/*
transaction パッケージは、トランザクション管理に関する機能を提供するパッケージです。
*/
package transaction

import (
	"example.com/appbase/pkg/async"
	myConfig "example.com/appbase/pkg/config"
	"example.com/appbase/pkg/logging"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/cockroachdb/errors"
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
	TransactSendMessages(inputs []*Message) error
	// EndTransactionは、トランザクションを終了します。
	EndTransaction()
}

// NewTransactionalSQSAccessor は、TransactionalSQSAccessorを作成します。
func NewTransactionalSQSAccessor(log logging.Logger, myCfg myConfig.Config) (TransactionalSQSAccessor, error) {
	sqsAccessor, err := async.NewSQSAccessor(log, myCfg)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	messageRegisterer := NewMessageRegisterer(myCfg)
	return &defaultTransactionalSQSAccessor{
		log:               log,
		sqsAccessor:       sqsAccessor,
		messageRegisterer: messageRegisterer,
	}, nil
}

// defaultTransactionalSQSAccessor は、TransactionalSQSAccessorを実装する構造体です。
type defaultTransactionalSQSAccessor struct {
	log               logging.Logger
	sqsAccessor       async.SQSAccessor
	messageRegisterer MessageRegisterer
	transaction       Transaction
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
func (sa *defaultTransactionalSQSAccessor) TransactSendMessages(inputs []*Message) error {
	sa.log.Debug("TransactSendMessages")
	for _, v := range inputs {
		// SQSへメッセージ送信
		output, err := sa.SendMessageSdk(v.QueueName, v.Input)
		if err != nil {
			//TODO: forの途中でエラーを返却することハンドリングが問題ないか再考
			return errors.WithStack(err)
		}
		sa.log.Debug("Send Message Id=%s", *output.MessageId)

		// DBトランザクションにアイテムを追加
		queueMessageItem := &QueueMessageItem{}
		queueMessageItem.MessageId = *output.MessageId
		if v.Input.MessageGroupId != nil {
			queueMessageItem.MessageDeduplicationId = *v.Input.MessageDeduplicationId
		}
		//TODO: DeleteTime（delete_time）の値を設定
		//queueMessageItem.DeleteTime = 0

		if err := sa.messageRegisterer.RegisterMessage(sa.transaction, queueMessageItem); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

// EndTransaction implements TransactionalSQSAccessor.
func (sa *defaultTransactionalSQSAccessor) EndTransaction() {
	sa.transaction = nil
}