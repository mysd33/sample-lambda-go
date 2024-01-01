/*
transaction パッケージは、トランザクション管理に関する機能を提供するパッケージです。
*/
package transaction

import (
	"example.com/appbase/internal/pkg/entity"

	"example.com/appbase/pkg/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/cockroachdb/errors"
)

// MessageRegisterer は、メッセージをトランザクションに登録するためのインターフェースです。
type MessageRegisterer interface {
	RegisterMessage(transaction Transaction, queueMessage *entity.QueueMessageItem) error
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
func (*defaultMessageRegisterer) RegisterMessage(transaction Transaction, queueMessage *entity.QueueMessageItem) error {
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
