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

type Message struct {
	QueueName string
	Input     *sqs.SendMessageInput
}

// TODO: Mock化しやすいよう全て公開メソッド化する

// TransactionalSQSDBAccessorは、トランザクション管理可能なSQSアクセス用インタフェースです。
type TransactionalSQSAccessor interface {
	async.SQSAccessor
	// startTransaction は、トランザクションを開始します。
	startTransaction(transaction transaction)
	// AppendMessage は、送信するメッセージをトランザクション管理したい場合に対象をメッセージを追加します
	AppendTransactMessage(queueName string, input *sqs.SendMessageInput) error
	// transactSendMessages は、トランザクション管理されたメッセージを送信します。
	// メッセージの送信は、TransactionManagerTransactionManagerが実行するため非公開にしています。
	transactSendMessages(inputs []*Message) error
	// endTransactionは、トランザクションを終了します。
	endTransaction()
}

// NewSQSAccessor は、TransactionalSQSAccessorを作成します。
func NewSQSAccessor(log logging.Logger, myCfg myConfig.Config) (TransactionalSQSAccessor, error) {
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

type defaultSQSAccessor struct {
	config      myConfig.Config
	log         logging.Logger
	sqsClient   *sqs.Client
	transaction transaction
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
	if input.MessageGroupId != nil {
		sa.log.Debug("MessageGroupId=%s, MessageDeduplicationId=%s, Message=%s",
			*input.MessageGroupId,
			*input.MessageDeduplicationId,
			*input.MessageBody)
	} else {
		sa.log.Debug("Message=%s", *input.MessageBody)
	}

	input.QueueUrl = queueUrlOutput.QueueUrl

	output, err := sa.sqsClient.SendMessage(apcontext.Context, input)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return output, nil
}

// startTransaction implements TransactionalSQSAccessor.
func (sa *defaultSQSAccessor) startTransaction(transaction transaction) {
	sa.transaction = transaction
}

// AppendTransactMessage implements TransactionalSQSAccessor.
func (sa *defaultSQSAccessor) AppendTransactMessage(queueName string, input *sqs.SendMessageInput) error {
	sa.log.Debug("AppendTransactMessage")
	sa.transaction.appendTransactMessage(&Message{QueueName: queueName, Input: input})
	return nil
}

// transactSendMessages implements TransactionalSQSAccessor.
func (sa *defaultSQSAccessor) transactSendMessages(inputs []*Message) error {
	for _, v := range inputs {
		// SQSへメッセージ送信
		output, err := sa.SendMessageSdk(v.QueueName, v.Input)
		if err != nil {
			//TODO: forの途中でエラーを返却することハンドリングが問題ないか再考
			return errors.WithStack(err)
		}
		sa.log.Debug("Send Message Id=%s", *output.MessageId)

		//TODO: QueueMessageRepositoryを作って別メソッドに切り出す
		// DBトランザクションにアイテムを追加
		queueMessageItem := &QueueMessageItem{}
		queueMessageItem.MessageId = *output.MessageId
		if v.Input.MessageGroupId != nil {
			queueMessageItem.MessageDeduplicationId = *v.Input.MessageDeduplicationId
		}
		//TODO: delete_timeの設定
		av, err := attributevalue.MarshalMap(queueMessageItem)
		if err != nil {
			//TODO: forの途中でエラーを返却することハンドリングが問題ないか再考
			return errors.WithStack(err)
		}
		put := &types.Put{
			Item: av,
			//TODO: テーブル名をプロパティ管理で設定切り出し
			TableName: aws.String("queue_message"),
		}
		// TransactWriteItemの追加
		item := &types.TransactWriteItem{Put: put}
		sa.transaction.appendTransactWriteItem(item)
	}

	return nil
}

// endTransaction implements TransactionalSQSAccessor.
func (sa *defaultSQSAccessor) endTransaction() {
	sa.transaction = nil
}

// QueueMessageItem は、QueueMessageテーブル用のEntityの構造体です。
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
