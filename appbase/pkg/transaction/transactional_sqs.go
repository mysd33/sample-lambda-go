/*
transaction パッケージは、トランザクション管理に関する機能を提供するパッケージです。
*/
package transaction

import (
	"maps"
	"strconv"
	"time"

	"example.com/appbase/pkg/async"
	myConfig "example.com/appbase/pkg/config"
	"example.com/appbase/pkg/constant"
	"example.com/appbase/pkg/logging"
	"example.com/appbase/pkg/transaction/entity"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/cockroachdb/errors"
)

const (
	QUEUE_MESSAGE_TABLE_TTL_HOUR = "QUEUE_MESSAGE_TABLE_TTL_HOUR"
)

// Message は、送信先のSQSのキューとメッセージのペアを管理する構造体です。
type Message struct {
	// キュー名
	QueueName string
	// SQS送信するメッセージ
	Input *sqs.SendMessageInput
}

// TransactionalSQSDBAccessorは、トランザクション管理可能なSQSアクセス用インタフェースです。
type TransactionalSQSAccessor interface {
	async.SQSAccessor
	// StartTransaction は、トランザクションを開始します。
	StartTransaction(transaction Transaction)
	// AppendMessage は、送信するメッセージをトランザクション管理したい場合に対象をメッセージを追加します
	AppendTransactMessage(queueName string, input *sqs.SendMessageInput) error
	// TransactSendMessages は、トランザクション管理されたメッセージを送信します。
	// メッセージの送信は、TransactionManagerTransactionManagerが実行するため非公開にしています。
	TransactSendMessages(inputs []*Message, hasDBTranaction bool) error
	// EndTransactionは、トランザクションを終了します。
	EndTransaction()
}

// NewTransactionalSQSAccessor は、TransactionalSQSAccessorを作成します。
func NewTransactionalSQSAccessor(log logging.Logger, myCfg myConfig.Config, messageRegisterer MessageRegisterer) (TransactionalSQSAccessor, error) {
	sqsAccessor, err := async.NewSQSAccessor(log, myCfg)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	// TTL（時間）の取得
	ttl, err := strconv.Atoi(myCfg.Get(QUEUE_MESSAGE_TABLE_TTL_HOUR))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &defaultTransactionalSQSAccessor{
		log:               log,
		config:            myCfg,
		sqsAccessor:       sqsAccessor,
		messageRegisterer: messageRegisterer,
		ttl:               ttl,
	}, nil
}

// defaultTransactionalSQSAccessor は、TransactionalSQSAccessorを実装する構造体です。
type defaultTransactionalSQSAccessor struct {
	log               logging.Logger
	config            myConfig.Config
	sqsAccessor       async.SQSAccessor
	messageRegisterer MessageRegisterer
	transaction       Transaction
	ttl               int
}

// SendMessageSdk implements TransactionalSQSAccessor.
func (sa *defaultTransactionalSQSAccessor) SendMessageSdk(queueName string, input *sqs.SendMessageInput) (*sqs.SendMessageOutput, error) {
	return sa.sqsAccessor.SendMessageSdk(queueName, input)
}

// StartTransaction implements TransactionalSQSAccessor.
func (sa *defaultTransactionalSQSAccessor) StartTransaction(transaction Transaction) {
	sa.transaction = transaction
}

// AppendTransactMessage implements TransactionalSQSAccessor.
func (sa *defaultTransactionalSQSAccessor) AppendTransactMessage(queueName string, input *sqs.SendMessageInput) error {
	sa.log.Debug("AppendTransactMessage")
	sa.transaction.AppendTransactMessage(&Message{QueueName: queueName, Input: input})
	return nil
}

// TransactSendMessages implements TransactionalSQSAccessor.
func (sa *defaultTransactionalSQSAccessor) TransactSendMessages(inputs []*Message, hasDbTrancation bool) error {
	sa.log.Debug("TransactSendMessages: %d件", len(inputs))

	for _, v := range inputs {
		// 業務テーブルでのDynamoDBトランザクション処理がある場合は、メッセージに削除時間を追加する
		sa.addDeleteTime(v, hasDbTrancation)
		// 業務テーブルでのDynamoDBトランザクション処理がない場合は、メッセージにフラグ情報を追加する。
		sa.addIsTableCheckFlag(v, hasDbTrancation)
		// SQSへメッセージ送信
		output, err := sa.SendMessageSdk(v.QueueName, v.Input)
		if err != nil {
			//TODO: forの途中でエラーを返却することハンドリングが問題ないか再考
			return errors.WithStack(err)
		}
		sa.log.Debug("Send Message Id=%s", *output.MessageId)

		// 業務テーブルでのDynamoDBトランザクション処理がある場合
		if hasDbTrancation {
			// メッセージ管理テーブル用のアイテムのトランザクション登録処理を追加
			queueMessageItem := &entity.QueueMessageItem{}
			// (キュー名) + "_" + (メッセージID)をパーティションキーとする
			queueMessageItem.MessageId = v.QueueName + "_" + *output.MessageId
			// メッセージ重複排除IDは送信時は格納しない（処理済みフラグ代わりに使用しているため）
			// DeleteTime（delete_time）の値を設定
			deleteTime, err := strconv.Atoi(*v.Input.MessageAttributes["delete_time"].StringValue)
			if err != nil {
				return errors.WithStack(err)
			}
			queueMessageItem.DeleteTime = deleteTime
			if err := sa.messageRegisterer.RegisterMessage(queueMessageItem); err != nil {
				return errors.WithStack(err)
			}
		}
	}

	return nil
}

// EndTransaction implements TransactionalSQSAccessor.
func (sa *defaultTransactionalSQSAccessor) EndTransaction() {
	sa.transaction = nil
}

// addDeleteTime は、業務テーブルでのDynamoDBトランザクション処理がある場合に削除時間をメッセージに追加します。
func (sa *defaultTransactionalSQSAccessor) addDeleteTime(v *Message, hasDbTrancation bool) {
	if !hasDbTrancation {
		return
	}
	nowTime := time.Now()
	delTimeStr := strconv.FormatInt(nowTime.Add(time.Duration(sa.ttl)*time.Hour).Unix(), 10)
	deleteTime := map[string]types.MessageAttributeValue{
		constant.DELETE_TIME_NAME: {
			DataType:    aws.String("String"),
			StringValue: aws.String(delTimeStr),
		},
	}

	if v.Input.MessageAttributes == nil {
		v.Input.MessageAttributes = deleteTime
	} else {
		maps.Copy(v.Input.MessageAttributes, deleteTime)
	}

}

// addIsTableCheckFlag は、業務テーブルでのDynamoDBトランザクション処理がない場合にメッセージにフラグ情報を追加します。
func (*defaultTransactionalSQSAccessor) addIsTableCheckFlag(v *Message, hasDbTrancation bool) {
	if hasDbTrancation {
		return
	}
	needsTableChecked := map[string]types.MessageAttributeValue{
		constant.NEEDS_TABLE_CHECK_NAME: {
			DataType:    aws.String("String"),
			StringValue: aws.String("false"),
		},
	}
	if v.Input.MessageAttributes == nil {
		v.Input.MessageAttributes = needsTableChecked
	} else {
		maps.Copy(v.Input.MessageAttributes, needsTableChecked)
	}
}
