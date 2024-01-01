package repository

import (
	"example.com/appbase/internal/pkg/entity"
	mytables "example.com/appbase/internal/pkg/repository/tables"
	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/dynamodb/input"
	"example.com/appbase/pkg/dynamodb/tables"
	"example.com/appbase/pkg/logging"
	"example.com/appbase/pkg/transaction"
	"github.com/cockroachdb/errors"
)

type QueueMessageItemRepository interface {
	FindOne(messageId string) (*entity.QueueMessageItem, error)
	CreateOneWithTx(queueMessage *entity.QueueMessageItem) error
}

func NewQueueMessageItemRepository(config config.Config,
	log logging.Logger,
	dynamodbTemplate transaction.TransactionalDynamoDBTemplate) QueueMessageItemRepository {
	// テーブル名取得
	//TODO: テーブル名をプロパティ管理(Config.Getで取得)で設定切り出し
	tableName := tables.DynamoDBTableName("queue_message")
	// テーブル定義の設定
	mytables.QueueMessageTable{}.InitPK(tableName)
	// プライマリキーの設定
	primaryKey := tables.GetPrimaryKey(tableName)
	return &defaultQueueMessageItemRepository{
		log:              log,
		dynamodbTemplate: dynamodbTemplate,
		tableName:        tableName,
		primaryKey:       primaryKey,
	}
}

type defaultQueueMessageItemRepository struct {
	log              logging.Logger
	dynamodbTemplate transaction.TransactionalDynamoDBTemplate
	tableName        tables.DynamoDBTableName
	primaryKey       *tables.PKKeyPair
}

// CreateOneWithTx implements QueueMessageItemRepository.
func (r *defaultQueueMessageItemRepository) CreateOneWithTx(queueMessage *entity.QueueMessageItem) error {
	err := r.dynamodbTemplate.CreateOneWithTransaction(r.tableName, queueMessage)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// FindOne implements QueueMessageItemRepository.
func (r *defaultQueueMessageItemRepository) FindOne(messageId string) (*entity.QueueMessageItem, error) {
	input := input.PKOnlyQueryInput{
		PrimaryKey: input.PrimaryKey{
			PartitionKey: input.Attribute{
				Name:  r.primaryKey.PartitionKey,
				Value: messageId,
			},
		},
	}
	var queueMessageItem entity.QueueMessageItem
	// Itemの取得
	err := r.dynamodbTemplate.FindOneByTableKey(r.tableName, input, &queueMessageItem)
	if err != nil {
		return nil, err
	}
	return &queueMessageItem, nil
}
