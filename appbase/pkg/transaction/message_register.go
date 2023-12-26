/*
transaction パッケージは、トランザクション管理に関する機能を提供するパッケージです。
*/
package transaction

import (
	"example.com/appbase/pkg/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/cockroachdb/errors"
)

// QueueMessageItem は、QueueMessageテーブルのアイテムを表す構造体です。
type QueueMessageItem struct {
	MessageId              string `dynamodbav:"message_id"`
	DeleteTime             string `dynamodbav:"delete_time"`
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

// MessageRegisterer は、メッセージをトランザクションに登録するためのインターフェースです。
type MessageRegisterer interface {
	RegisterMessage(transaction Transaction, queueMessage *QueueMessageItem) error
}

// NewMessageRegisterer は、MessageRegistererを作成します。
func NewMessageRegisterer(config config.Config) MessageRegisterer {
	return &defaultMessageRegisterer{config: config}
}

// defaultMessageRegisterer は、MessageRegistererの実装です。
type defaultMessageRegisterer struct {
	config config.Config
}

// RegisterMessage implements MessageRegisterer.
func (*defaultMessageRegisterer) RegisterMessage(transaction Transaction, queueMessage *QueueMessageItem) error {
	av, err := attributevalue.MarshalMap(queueMessage)
	if err != nil {
		return errors.WithStack(err)
	}

	// TODO: キュー管理テーブルへのアクセス処理を、QueueMessageItemRepositoryへ切り出す予定
	put := &types.Put{
		Item: av,
		//TODO: テーブル名をプロパティ管理(Config.Getで取得)で設定切り出し
		TableName: aws.String("queue_message"),
	}
	item := &types.TransactWriteItem{Put: put}
	transaction.AppendTransactWriteItem(item)
	return nil
}
