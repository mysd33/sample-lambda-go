/*
transaction パッケージは、トランザクション管理に関する機能を提供するパッケージです。
*/
package transaction

import (
	"context"
	"maps"
	"strconv"
	"time"

	"example.com/appbase/pkg/apcontext"
	"example.com/appbase/pkg/async"
	myConfig "example.com/appbase/pkg/config"
	"example.com/appbase/pkg/constant"
	"example.com/appbase/pkg/logging"
	"example.com/appbase/pkg/transaction/model"
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
	// AppendMessage は、送信するメッセージをトランザクション管理したい場合に対象をメッセージを追加します
	// なお、メッセージの送信は、TransactionManagerのExecuteTransaction関数で実行されるdomain.ServiceFunc関数が終了する際にtransactionWriteItemsSDKを実施します。
	AppendTransactMessage(queueName string, input *sqs.SendMessageInput) error
	// AppendTransactMessageWithContext は、goroutine向けに渡されたContextを利用して、送信するメッセージをトランザクション管理したい場合に対象をメッセージを追加します
	// なお、メッセージの送信は、TransactionManagerのExecuteTransactionWithContext関数で実行されるdomain.ServiceFuncWithContext関数が終了する際にtransactionWriteItemsSDKを実施します。
	AppendTransactMessageWithContext(ctx context.Context, queueName string, input *sqs.SendMessageInput) error
	// TransactSendMessages は、トランザクション管理されたメッセージを送信します。
	// なお、TransactSendMessagesの実行は、TransactionManagerが実行するため業務ロジックで利用する必要はありません。
	TransactSendMessages(inputs []*Message, optFns ...func(*sqs.Options)) error
	// TransactSendMessagesWithContext は、goroutine向けに渡されたContextを利用して、
	/// トランザクション管理されたメッセージを送信します。
	// なお、TransactSendMessagesWithContextの実行は、TransactionManagerが実行するため業務ロジックで利用する必要はありません。
	TransactSendMessagesWithContext(ctx context.Context, inputs []*Message, optFns ...func(*sqs.Options)) error
}

// NewTransactionalSQSAccessor は、TransactionalSQSAccessorを作成します。
func NewTransactionalSQSAccessor(logger logging.Logger, myCfg myConfig.Config, messageRegisterer MessageRegisterer) (TransactionalSQSAccessor, error) {
	sqsAccessor, err := async.NewSQSAccessor(logger, myCfg)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	// TTL（時間）の取得
	ttl := myCfg.GetInt(QUEUE_MESSAGE_TABLE_TTL_HOUR, 24*4)
	return &defaultTransactionalSQSAccessor{
		logger:            logger,
		config:            myCfg,
		sqsAccessor:       sqsAccessor,
		messageRegisterer: messageRegisterer,
		ttl:               ttl,
	}, nil
}

// defaultTransactionalSQSAccessor は、TransactionalSQSAccessorを実装する構造体です。
type defaultTransactionalSQSAccessor struct {
	logger            logging.Logger
	config            myConfig.Config
	sqsAccessor       async.SQSAccessor
	messageRegisterer MessageRegisterer
	ttl               int
}

// SendMessageSdk implements TransactionalSQSAccessor.
func (sa *defaultTransactionalSQSAccessor) SendMessageSdk(queueName string, input *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
	return sa.sqsAccessor.SendMessageSdk(queueName, input, optFns...)
}

// SendMessageSdkWithContext implements TransactionalSQSAccessor.
func (sa *defaultTransactionalSQSAccessor) SendMessageSdkWithContext(ctx context.Context, queueName string, input *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
	return sa.sqsAccessor.SendMessageSdkWithContext(ctx, queueName, input, optFns...)
}

// AppendTransactMessage implements TransactionalSQSAccessor.
func (sa *defaultTransactionalSQSAccessor) AppendTransactMessage(queueName string, input *sqs.SendMessageInput) error {
	sa.logger.Debug("AppendTransactMessage")
	return sa.AppendTransactMessageWithContext(apcontext.Context, queueName, input)
}

// AppendTransactMessageWithContext implements TransactionalSQSAccessor.
func (sa *defaultTransactionalSQSAccessor) AppendTransactMessageWithContext(ctx context.Context, queueName string, input *sqs.SendMessageInput) error {
	sa.logger.Debug("AppendTransactMessageWithContext")
	if ctx == nil {
		ctx = apcontext.Context
	}
	value := ctx.Value(TRANSACTION_CTX_KEY)
	if value == nil {
		return errors.New("トランザクションが開始されていません")
	}
	transaction, ok := value.(Transaction)
	if !ok {
		return errors.New("トランザクションが開始されていません")
	}
	transaction.AppendTransactMessage(&Message{QueueName: queueName, Input: input})
	return nil
}

// TransactSendMessages implements TransactionalSQSAccessor.
func (sa *defaultTransactionalSQSAccessor) TransactSendMessages(inputs []*Message, optFns ...func(*sqs.Options)) error {
	return sa.TransactSendMessagesWithContext(apcontext.Context, inputs, optFns...)
}

// TransactSendMessagesWithContext implements TransactionalSQSAccessor.
func (sa *defaultTransactionalSQSAccessor) TransactSendMessagesWithContext(ctx context.Context, inputs []*Message, optFns ...func(*sqs.Options)) error {
	sa.logger.Debug("TransactSendMessages: %d件", len(inputs))
	if ctx == nil {
		ctx = apcontext.Context
	}
	for _, v := range inputs {
		// メッセージに削除時間を追加する
		sa.addDeleteTime(v)
		// SQSへメッセージ送信
		output, err := sa.SendMessageSdkWithContext(ctx, v.QueueName, v.Input, optFns...)
		if err != nil {
			//TODO: forの途中でエラーを返却することハンドリングが問題ないか再考
			return errors.WithStack(err)
		}
		sa.logger.Debug("Send Message Id=%s", *output.MessageId)

		// メッセージ管理テーブル用のアイテムのトランザクション登録処理を追加
		queueMessageItem := &model.QueueMessageItem{}
		// (キュー名) + "_" + (メッセージID)をパーティションキーとする
		queueMessageItem.MessageId = v.QueueName + "_" + *output.MessageId

		// FIFOでメッセージグループIDが設定されている場合は、メッセージグループIDを設定
		if v.Input.MessageGroupId != nil {
			queueMessageItem.MessageGroupId = *v.Input.MessageGroupId
		}

		// ステータスは送信時は格納していない
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

	return nil
}

// addDeleteTime は、削除時間をメッセージに追加します。
func (sa *defaultTransactionalSQSAccessor) addDeleteTime(v *Message) {
	nowTime := time.Now()
	delTimeStr := strconv.FormatInt(nowTime.Add(time.Duration(sa.ttl)*time.Hour).Unix(), 10)
	deleteTime := map[string]types.MessageAttributeValue{
		constant.QUEUE_MESSAGE_DELETE_TIME_NAME: {
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
