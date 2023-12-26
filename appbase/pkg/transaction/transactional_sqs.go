/*
transaction パッケージは、トランザクション管理に関する機能を提供するパッケージです。
*/
package transaction

import (
	"maps"

	"example.com/appbase/pkg/async"
	myConfig "example.com/appbase/pkg/config"
	"example.com/appbase/pkg/logging"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
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
	TransactSendMessages(inputs []*Message, hasDBTranaction bool) error
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
	//TODO: SendMessageInputに、DeleteTime（delete_time）の値を設定
	sa.transaction.AppendTransactMessage(&Message{QueueName: queueName, Input: input})
	return nil
}

// TransactSendMessages implements TransactionalSQSAccessor.
func (sa *defaultTransactionalSQSAccessor) TransactSendMessages(inputs []*Message, hasDbTrancation bool) error {
	sa.log.Debug("TransactSendMessages")
	for _, v := range inputs {
		// 業務テーブルでのDynamoDBトランザクション処理がない場合は、メッセージにフラグ情報を送る
		addIsTableCheckFlag(v, hasDbTrancation)
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
			queueMessageItem := &QueueMessageItem{}
			// (キュー名) + "_" + (メッセージID)をパーティションキーとする
			queueMessageItem.MessageId = v.QueueName + "_" + *output.MessageId
			if v.Input.MessageGroupId != nil {
				queueMessageItem.MessageDeduplicationId = *v.Input.MessageDeduplicationId
			}
			//TODO: DeleteTime（delete_time）の値を設定
			//queueMessageItem.DeleteTime = 0
			if err := sa.messageRegisterer.RegisterMessage(sa.transaction, queueMessageItem); err != nil {
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

// addIsTableCheckFlag は、業務テーブルでのDynamoDBトランザクション処理がない場合にメッセージにフラグ情報を追加します
func addIsTableCheckFlag(v *Message, hasDbTrancation bool) {
	if hasDbTrancation {
		return
	}
	isTableChecked := map[string]types.MessageAttributeValue{
		// TODO: 定数化
		"is_table_check": {
			DataType:    aws.String("String"),
			StringValue: aws.String("false"),
		},
	}
	if v.Input.MessageAttributes == nil {
		v.Input.MessageAttributes = isTableChecked
	} else {
		maps.Copy(v.Input.MessageAttributes, isTableChecked)
	}
}
