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

	"example.com/appbase/pkg/apcontext"
	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/constant"
	"example.com/appbase/pkg/logging"
	"example.com/appbase/pkg/message"
	"example.com/appbase/pkg/transaction"
	"example.com/appbase/pkg/transaction/entity"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/cockroachdb/errors"
)

const (
	// TODO: メッセージ受信時、キューメッセージ管理テーブルのアイテムがない場合のリトライ回数の設定切り出し
	QUEUE_MESSAGE_DELETE_RETRY_COUNT = 1
	// TODO: リトライ間隔の設定切り出し
	TABLE_ACESS_RETRY_DURATION = time.Duration(300)
	// TODO: リトライ回数の設定切り出し
	TABLE_ACESS_RETRY_COUNT = 5
)

var ErrMessageIdNotFound = errors.New("メッセージIDが取得できません")

// SQSTriggeredLambdaHandlerFuncは、SQSトリガのLambdaのハンドラを表す関数です。
type SQSTriggeredLambdaHandlerFunc func(ctx context.Context, event events.SQSEvent) (events.SQSEventResponse, error)

// AsyncLambdaHandler は、SQSトリガの非同期処理のLambdaのハンドラを管理する構造体です。
type AsyncLambdaHandler struct {
	config                     config.Config
	log                        logging.Logger
	queueMessageItemRepository transaction.QueueMessageItemRepository
}

// NewAsyncLambdaHandler は、AsyncLambdaHandlerを作成します。
func NewAsyncLambdaHandler(config config.Config,
	log logging.Logger,
	queueMessageItemRepository transaction.QueueMessageItemRepository) *AsyncLambdaHandler {
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
		// ctxをコンテキスト領域に格納
		apcontext.Context = ctx

		// FIFOの対応（FIFOの場合はメッセージグループID毎にメッセージのソート）
		isFIFO := event.Records[0].Attributes[string(types.MessageSystemAttributeNameMessageGroupId)] != ""
		h.log.Debug("isFIFO: %t", isFIFO)
		if isFIFO {
			// FIFOの場合はメッセージをソート
			h.sortMessages(event.Records)
		}
		for _, v := range event.Records {
			// SQSのメッセージを1件取得しコントローラを呼び出し
			h.doHandle(v, response, isFIFO, asyncControllerFunc)
			if err != nil {
				// 失敗したメッセージIDをBatchItemFailuresに登録
				response.BatchItemFailures = append(response.BatchItemFailures, events.SQSBatchItemFailure{ItemIdentifier: v.MessageId})
			}
		}
		return
	}
}

// doHandle は、SQSのメッセージを1件に対して、ディレード処理（ジョブ）を実行します。
func (h *AsyncLambdaHandler) doHandle(sqsMsg events.SQSMessage, response events.SQSEventResponse, isFIFO bool, asyncControllerFunc AsyncControllerFunc) error {
	// FIFOの場合は、以前のメッセージが失敗している場合は、当該メッセージもエラー処理をスキップする
	if isFIFO && response.BatchItemFailures != nil {
		return errors.New("以前のメッセージが失敗しているためエラー")
	}
	queueName := h.getQueueName(sqsMsg)
	messageId := sqsMsg.MessageId

	h.log.Debug("doHandle[QueueName: %s, MessageId: %s]", queueName, messageId)
	// キューメッセージテーブルのキーを作成
	deduplicationId, err := h.checkMessageId(sqsMsg)
	if err != nil {
		// メッセージIDが取得できない場合
		if errors.Is(err, ErrMessageIdNotFound) {
			// 受信回数が閾値を超えている場合、メッセージを削除
			receiveCount, _ := strconv.Atoi(sqsMsg.Attributes[string(types.MessageSystemAttributeNameApproximateReceiveCount)])
			if receiveCount >= QUEUE_MESSAGE_DELETE_RETRY_COUNT {
				// Errorログの出力
				h.log.Error(message.E_FW_9002, err)

				// メッセージ削除させるため正常終了
				return nil
			}
		}
		return err
	}
	// 処理済のメッセージ（重複排除IDがメッセージ管理テーブルに入っている）の場合
	if deduplicationId != "" {
		h.log.Debug("処理済のメッセージです。[QueueName: %s, MessageId: %s]", queueName, messageId)
		// 重複して処理しないよう正常終了
		return nil
	}
	// Contextに非同期処理情報を格納
	h.addAsyncInfoToContext(sqsMsg, isFIFO)

	// Controllerの実行（実際にはインタセプターを経由）
	return asyncControllerFunc(sqsMsg)
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
func (h *AsyncLambdaHandler) checkMessageId(sqsMsg events.SQSMessage) (string, error) {
	queueMessageTableId := h.getQueueMessageTableId(sqsMsg)
	// フラグが立っている場合は、DynamoDBのキューメッセージ管理テーブルを確認しない
	// Message1Attributeからis_table_checkの値を取得
	isTableCheck := sqsMsg.MessageAttributes[constant.IS_TABLE_CHECK_NAME].StringValue
	if isTableCheck != nil && *isTableCheck == "false" {
		// DBを確認しないため、メッセージ重複排除IDは空文字を返却
		h.log.Debug(constant.IS_TABLE_CHECK_NAME + ":あり")
		return "", nil
	}
	h.log.Debug("キューメッセージテーブルID: %s", queueMessageTableId)
	deleteTime := sqsMsg.MessageAttributes[constant.DELETE_TIME_NAME].StringValue
	retryCount := 0
	var queueMessageItem *entity.QueueMessageItem
	var err error
	for {
		// メッセージIDに対応するキューメッセージ管理テーブルからのアイテムを取得
		queueMessageItem, err = h.queueMessageItemRepository.FindOne(queueMessageTableId, *deleteTime)
		if err != nil {
			return "", err
		}
		// Lambdaタイムアウトの観点か閾値の回数か、メッセージIDを取得できた場合はループを抜ける
		if retryCount == TABLE_ACESS_RETRY_COUNT || queueMessageItem.MessageId != "" {
			break
		}
		retryCount++
		// メッセージIDを取得できなかった場合のWARNログの追加
		h.log.Warn(message.W_FW_8003, queueMessageTableId)
		// キューメッセージ管理テーブルへのアクセスリトライ時間待機
		time.Sleep(TABLE_ACESS_RETRY_DURATION * time.Millisecond)
	}
	if queueMessageItem.MessageId == "" {
		return "", ErrMessageIdNotFound
	}
	return queueMessageItem.MessageDeduplicationId, nil
}

// addAsyncInfoToContext は、非同期処理情報をContextに格納します。
func (h *AsyncLambdaHandler) addAsyncInfoToContext(sqsMsg events.SQSMessage, isFIFO bool) {
	var messageDeduplicationId string
	if isFIFO {
		// FIFOキューの場合は、実際のメッセージ重複排除IDを設定
		messageDeduplicationId = sqsMsg.Attributes[string(types.MessageSystemAttributeNameMessageDeduplicationId)]
	} else {
		// 標準キューの場合は、ダミーのメッセージ重複排除IDを設定
		messageDeduplicationId = "standard-queue-dummy"
	}
	// 処理成功時にメッセージ管理テーブルを更新するため、Context領域に非同期処理情報を格納しておく
	apcontext.Context = context.WithValue(apcontext.Context, constant.ASYNC_HANDLER_INFO_CTX_KEY,
		&entity.QueueMessageItem{
			MessageId:              h.getQueueMessageTableId(sqsMsg),
			MessageDeduplicationId: messageDeduplicationId,
			DeleteTime:             *sqsMsg.MessageAttributes[constant.DELETE_TIME_NAME].StringValue,
		},
	)
}

// getQueueMessageTableId は、キューメッセージ管理テーブルのキーを作成します。
func (h *AsyncLambdaHandler) getQueueName(sqsMsg events.SQSMessage) string {
	queueArn := strings.Split(sqsMsg.EventSourceARN, ":")
	queueName := queueArn[len(queueArn)-1]
	return queueName
}

// getQueueMessageTableId は、キューメッセージ管理テーブルのキーを作成します。
func (h *AsyncLambdaHandler) getQueueMessageTableId(sqsMsg events.SQSMessage) string {
	queueName := h.getQueueName(sqsMsg)
	queueMessageId := queueName + "_" + sqsMsg.MessageId
	return queueMessageId
}
