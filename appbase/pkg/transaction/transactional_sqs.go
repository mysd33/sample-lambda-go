/*
transaction パッケージは、トランザクション管理に関する機能を提供するパッケージです。
*/
package transaction

import (
	"context"

	"example.com/appbase/pkg/apcontext"
	"example.com/appbase/pkg/async"
	myConfig "example.com/appbase/pkg/config"
	"example.com/appbase/pkg/constant"
	"example.com/appbase/pkg/logging"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-xray-sdk-go/instrumentation/awsv2"
	"github.com/cockroachdb/errors"
)

// Message は、送信先のSQSのキューとメッセージのペアを管理する構造体です。
type Message struct {
	// キュー名
	QueueName string
	// SQS送信するメッセージ
	Input *sqs.SendMessageInput
}

// QueueMessageItem は、QueueMessageテーブルのアイテムを表す構造体です。
type QueueMessageItem struct {
	MessageId              string `dynamodbav:"message_id"`
	DeleteTime             int    `dynamodbav:"delete_time"`
	MessageDeduplicationId string `dynamodbav:"message_deduplication_id"`
}

// GetKey は、DynamoDBのキー情報を取得します。
func (m QueueMessageItem) GetKey() (map[string]types.AttributeValue, error) {
	id, err := attributevalue.Marshal(m.MessageId)
	if err != nil {
		return nil, err
	}

	return map[string]types.AttributeValue{"message_id": id}, nil
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
	// TODO: カスタムHTTPClientの作成
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, errors.WithStack(err)
	}
	awsv2.AWSV2Instrumentor(&cfg.APIOptions)
	sqlClient := sqs.NewFromConfig(cfg, func(o *sqs.Options) {
		// ローカル実行のためDynamoDB Local起動先が指定されている場合
		sqsEndpoint := myCfg.Get(constant.SQS_LOCAL_ENDPOINT_NAME)
		if sqsEndpoint != "" {
			o.BaseEndpoint = aws.String(sqsEndpoint)
		}
	})
	return &defaultSQSAccessor{
		config:    myCfg,
		log:       log,
		sqsClient: sqlClient,
	}, nil
}

// defaultSQSAccessor は、TransactionalSQSAccessorを実装する構造体です。
type defaultSQSAccessor struct {
	config      myConfig.Config
	log         logging.Logger
	sqsClient   *sqs.Client
	transaction Transaction
}

// SendMessageSdk implements SQSAccessor.
func (sa *defaultSQSAccessor) SendMessageSdk(queueName string, input *sqs.SendMessageInput) (*sqs.SendMessageOutput, error) {
	// QueueのURLの取得・設定
	queueUrlOutput, err := sa.sqsClient.GetQueueUrl(apcontext.Context, &sqs.GetQueueUrlInput{
		QueueName: aws.String(queueName),
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	sa.log.Debug("QueueURL=%s", *queueUrlOutput.QueueUrl)
	// 送信先の設定
	input.QueueUrl = queueUrlOutput.QueueUrl

	if input.MessageGroupId != nil {
		sa.log.Debug("MessageGroupId=%s, MessageDeduplicationId=%s, Message=%s",
			*input.MessageGroupId,
			*input.MessageDeduplicationId,
			*input.MessageBody)
	} else {
		sa.log.Debug("Message=%s", *input.MessageBody)
	}
	//　SQSへメッセージ送信する
	output, err := sa.sqsClient.SendMessage(apcontext.Context, input)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return output, nil
}

// StartTransaction implements TransactionalSQSAccessor.
func (sa *defaultSQSAccessor) StartTransaction(transaction Transaction) {
	sa.transaction = transaction
}

// AppendTransactMessage implements TransactionalSQSAccessor.
func (sa *defaultSQSAccessor) AppendTransactMessage(queueName string, input *sqs.SendMessageInput) error {
	sa.log.Debug("AppendTransactMessage")
	sa.transaction.AppendTransactMessage(&Message{QueueName: queueName, Input: input})
	return nil
}

// TransactSendMessages implements TransactionalSQSAccessor.
func (sa *defaultSQSAccessor) TransactSendMessages(inputs []*Message) error {
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
		if err := sa.appendTransactWriteItemForQueueMessage(queueMessageItem); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

// EndTransaction implements TransactionalSQSAccessor.
func (sa *defaultSQSAccessor) EndTransaction() {
	sa.transaction = nil
}

// appendTransactWriteItemForQueueMessage は、QueueMessageテーブルにTransactWriteItem操作を追加します。
func (sa *defaultSQSAccessor) appendTransactWriteItemForQueueMessage(queueMessageItem *QueueMessageItem) error {
	av, err := attributevalue.MarshalMap(queueMessageItem)
	if err != nil {
		return errors.WithStack(err)
	}
	put := &types.Put{
		Item: av,
		//TODO: テーブル名をプロパティ管理で設定切り出し
		TableName: aws.String("queue_message"),
	}
	item := &types.TransactWriteItem{Put: put}
	sa.transaction.AppendTransactWriteItem(item)
	return nil
}
