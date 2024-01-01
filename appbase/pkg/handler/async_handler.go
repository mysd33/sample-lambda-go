/*
handler パッケージは、Lambdaのハンドラメソッドに関する機能を提供するパッケージです。
*/
package handler

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"example.com/appbase/internal/pkg/entity"
	"example.com/appbase/internal/pkg/repository"
	"example.com/appbase/pkg/apcontext"
	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/logging"
	"example.com/appbase/pkg/message"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/cockroachdb/errors"
)

var ErrMessageIdNotFound = errors.New("メッセージIDが取得できません")

// SQSTriggeredLambdaHandlerFuncは、SQSトリガのLambdaのハンドラを表す関数です。
type SQSTriggeredLambdaHandlerFunc func(ctx context.Context, event events.SQSEvent) (events.SQSEventResponse, error)

// AsyncLambdaHandler は、SQSトリガの非同期処理のLambdaのハンドラを管理する構造体です。
type AsyncLambdaHandler struct {
	config                     config.Config
	log                        logging.Logger
	queueMessageItemRepository repository.QueueMessageItemRepository
}

// NewAsyncLambdaHandler は、AsyncLambdaHandlerを作成します。
func NewAsyncLambdaHandler(config config.Config,
	log logging.Logger,
	queueMessageItemRepository repository.QueueMessageItemRepository) *AsyncLambdaHandler {
	return &AsyncLambdaHandler{
		config:                     config,
		log:                        log,
		queueMessageItemRepository: queueMessageItemRepository,
	}
}

// Handleは、SQSトリガのLambdaのハンドラを実行します。
func (h *AsyncLambdaHandler) Handle(asyncControllerFunc AsyncControllerFunc) SQSTriggeredLambdaHandlerFunc {
	return func(ctx context.Context, event events.SQSEvent) (response events.SQSEventResponse, err error) {
		defer func() {
			// パニックのリカバリ処理
			if v := recover(); v != nil {
				err = fmt.Errorf("recover from: %+v", v)
				// パニックのスタックトレース情報をログ出力
				h.log.ErrorWithUnexpectedError(err)
				// 全てのメッセージを失敗扱いにする
				response.BatchItemFailures = append(response.BatchItemFailures, events.SQSBatchItemFailure{ItemIdentifier: ""})
			}
		}()
		// FIFOの対応（FIFOの場合はメッセージグループID毎にメッセージのソート）
		isFifo := event.Records[0].Attributes[string(types.QueueAttributeNameFifoQueue)] == "true"

		if isFifo {
			// FIFOの場合はメッセージをソート
			h.sortMessages(event.Records)
		}
		// ctxをコンテキスト領域に格納
		apcontext.Context = ctx

		for _, v := range event.Records {
			// SQSのメッセージを1件取得しコントローラを呼び出し
			// TODO: DBとのデータ整合性の確認
			// TODO: 二重実行防止のチェック（メッセージIDの確認
			h.doHandle(v, response, isFifo, asyncControllerFunc)
			if err != nil {
				// 失敗したメッセージIDをBatchItemFailuresに登録
				response.BatchItemFailures = append(response.BatchItemFailures, events.SQSBatchItemFailure{ItemIdentifier: v.MessageId})
			}
		}
		return
	}
}

func (h *AsyncLambdaHandler) doHandle(sqsMsg events.SQSMessage, response events.SQSEventResponse, isFifo bool, asyncControllerFunc AsyncControllerFunc) error {
	// Fifoの場合は、以前のメッセージが失敗している場合は、当該メッセージもエラー処理をスキップする
	if isFifo && response.BatchItemFailures != nil {
		return errors.New("以前のメッセージが失敗しているためエラー")
	}
	// キュー名取得
	queueArn := strings.Split(sqsMsg.EventSourceARN, ":")
	queueName := queueArn[len(queueArn)-1]
	// キューメッセージテーブルのキーを作成
	queueMessageId := queueName + "_" + sqsMsg.MessageId
	deduplicationId, err := h.checkMessageId(queueMessageId, sqsMsg)

	// TODO: メッセージ削除までのリトライ回数の設定切り出し
	QUEUE_MESSAGE_DELETE_RETRY_COUNT := 1

	if err != nil {
		// メッセージIDが取得できない場合は
		if errors.Is(err, ErrMessageIdNotFound) {
			// 受信回数
			receiveCount, _ := strconv.Atoi(sqsMsg.Attributes[string(types.MessageSystemAttributeNameApproximateReceiveCount)])
			if receiveCount >= QUEUE_MESSAGE_DELETE_RETRY_COUNT {
				//TODO: Errorログの出力

				// メッセージ削除させるため正常終了
				return nil
			}
		}
		return err
	}

	// 処理済のメッセージの場合
	if deduplicationId == "" {
		// 重複して処理しないよう正常終了
		return nil
	}

	// Controllerの実行（実際にはインタセプターを経由）
	err = asyncControllerFunc(sqsMsg)

	// TODO: この後、メッセージ管理テーブルにTTLの更新または削除が必要では？
	return err
}

// sortMessages は、メッセージをMessageGroupIdごとにSequenceNamberを昇順にします。
func (h *AsyncLambdaHandler) sortMessages(sqsMsgs []events.SQSMessage) {
	sort.Slice(sqsMsgs, func(i, j int) bool {
		iMessageGroupId := sqsMsgs[i].Attributes[string(types.MessageSystemAttributeNameMessageGroupId)]
		jMessageGroupId := sqsMsgs[j].Attributes[string(types.MessageSystemAttributeNameMessageGroupId)]
		if iMessageGroupId != jMessageGroupId {
			// MessageGroupIdが異なる場合はMessageGroupIdでソート
			return iMessageGroupId < jMessageGroupId
		}
		// MessageGroupIdが同じ場合はSequenceNumberでソート
		iSequenceNumber, _ := strconv.Atoi(sqsMsgs[i].Attributes[string(types.MessageSystemAttributeNameSequenceNumber)])
		jSequenceNumber, _ := strconv.Atoi(sqsMsgs[j].Attributes[string(types.MessageSystemAttributeNameSequenceNumber)])
		return iSequenceNumber < jSequenceNumber
	})
}

// checkMessagId は、キューメッセージ管理テーブルにメッセージIDが存在するか確認する
func (h *AsyncLambdaHandler) checkMessageId(queueMessageId string, sqsMsg events.SQSMessage) (string, error) {
	// フラグが立っている場合は、DynamoDBのキューメッセージ管理テーブルを確認しない
	// Message1Attributeからis_table_checkの値を取得
	// TODO: is_table_checkを定数化
	isTableCheck := sqsMsg.MessageAttributes["is_table_check"].StringValue
	if isTableCheck != nil && *isTableCheck == "false" {
		// DBを確認しないため、メッセージ重複排除IDは空文字を返却
		h.log.Debug("is_table_checkあり")
		return "", nil
	}
	// キューメッセージ管理テーブルから該当のメッセージアイテムを取得
	h.log.Debug("キューメッセージID: %s", queueMessageId)
	// TODO: リトライ回数、リトライ時間の設定切り出し
	RETRY_TIME := time.Duration(300)
	RETRY_COUNT := 5
	retryCount := 0
	var queueMessageItem *entity.QueueMessageItem
	var err error
	// Lambdaタイムアウトの観点か5回リトライする
	for {
		// TODO: なぜ、元ネタでは、delete_timeでのFilterしているのか？
		// 現状、実装できていない
		queueMessageItem, err = h.queueMessageItemRepository.FindOne(queueMessageId)
		if err != nil {
			// TODO: Errorログ
			h.log.Error(message.E_FW_9001, err)
			return "", err
		}
		if retryCount == RETRY_COUNT || queueMessageItem.MessageId != "" {
			break
		}
		retryCount++
		// TODO: 取得できなかった場合のWARNログの追加

		time.Sleep(RETRY_TIME * time.Millisecond)
	}
	if queueMessageItem.MessageId == "" {
		// TODO: Errorログの追加

		return "", ErrMessageIdNotFound
	}
	return queueMessageItem.MessageDeduplicationId, nil
}
