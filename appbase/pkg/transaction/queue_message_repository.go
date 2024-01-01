package transaction

import (
	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/constant"
	"example.com/appbase/pkg/dynamodb"
	"example.com/appbase/pkg/dynamodb/input"
	"example.com/appbase/pkg/dynamodb/tables"
	"example.com/appbase/pkg/logging"
	"example.com/appbase/pkg/transaction/entity"
	mytables "example.com/appbase/pkg/transaction/tables"
	"github.com/cockroachdb/errors"
)

const (
	QUEUE_MESSAGE_TABLE_NAME = "QUEUE_MESSAGE_TABLE_NAME"
)

// QueueMessageItemRepository は、キューメッセージ管理テーブルのリポジトリインタフェースです。
type QueueMessageItemRepository interface {
	FindOne(messageId string, deleteTime string) (*entity.QueueMessageItem, error)
	CreateOneWithTx(queueMessage *entity.QueueMessageItem) error
	UpdateOneWithTx(queueMessage *entity.QueueMessageItem) error
}

// NewQueueMessageItemRepository は、QueueMessageItemRepositoryを作成します。
func NewQueueMessageItemRepository(config config.Config,
	log logging.Logger,
	dynamodbTemplate TransactionalDynamoDBTemplate) QueueMessageItemRepository {
	// テーブル名取得
	//TODO: テーブル名をプロパティ管理(Config.Getで取得)で設定切り出し
	//tableName := tables.DynamoDBTableName(config.Get("QUEUE_MESSAGE_TABLE_NAME"))
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

// defaultQueueMessageItemRepository は、QueueMessageItemRepositoryを実装する構造体です。
type defaultQueueMessageItemRepository struct {
	log              logging.Logger
	dynamodbTemplate TransactionalDynamoDBTemplate
	tableName        tables.DynamoDBTableName
	primaryKey       *tables.PKKeyPair
}

// FindOne implements QueueMessageItemRepository.
func (r *defaultQueueMessageItemRepository) FindOne(messageId string, deleteTime string) (*entity.QueueMessageItem, error) {
	r.log.Debug("partitionKey: %s", r.primaryKey.PartitionKey)
	input := input.PKQueryInput{
		PrimaryKey: input.PrimaryKey{
			PartitionKey: input.Attribute{
				Name:  r.primaryKey.PartitionKey,
				Value: messageId,
			},
		},
		// TODO: なぜ、元ネタでは、delete_timeでのFilterしている？
		WhereClauses: []*input.WhereClause{
			{
				Attribute: input.Attribute{
					Name:  constant.DELETE_TIME_NAME,
					Value: deleteTime,
				},
				WhereOp: input.WHERE_EQUAL,
			},
		},
		ConsitentRead: true,
	}
	var queueMessageItems []entity.QueueMessageItem
	// Itemの取得
	err := r.dynamodbTemplate.FindSomeByTableKey(r.tableName, input, &queueMessageItems)
	if err != nil {
		if errors.Is(err, dynamodb.ErrRecordNotFound) {
			return &entity.QueueMessageItem{}, nil
		}
		return nil, err
	}
	return &queueMessageItems[0], nil
}

// CreateOneWithTx implements QueueMessageItemRepository.
func (r *defaultQueueMessageItemRepository) CreateOneWithTx(queueMessage *entity.QueueMessageItem) error {
	err := r.dynamodbTemplate.CreateOneWithTransaction(r.tableName, queueMessage)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// UpdateOneWithTx implements QueueMessageItemRepository.
func (r *defaultQueueMessageItemRepository) UpdateOneWithTx(queueMessage *entity.QueueMessageItem) error {
	r.log.Debug("partitionKey: %s", r.primaryKey.PartitionKey)
	input := input.UpdateInput{
		PrimaryKey: input.PrimaryKey{
			PartitionKey: input.Attribute{
				Name:  r.primaryKey.PartitionKey,
				Value: queueMessage.MessageId,
			},
		},
		UpdateAttributes: []*input.Attribute{
			// TODO: 本来は、処理済みのStatus列を定義して更新すべき
			// メッセージ重複排除IDを登録する（処理済みフラグの代わりに使用している）
			{
				Name:  constant.MESSAGE_DEDUPLICATION_ID,
				Value: queueMessage.MessageDeduplicationId,
			},
		},
		WhereClauses: []*input.WhereClause{
			{
				Attribute: input.Attribute{
					Name:  r.primaryKey.PartitionKey,
					Value: queueMessage.MessageId,
				},
				WhereOp: input.WHERE_EQUAL,
			},
		},
	}
	r.log.Debug("メッセージ重複排除ID: %s", queueMessage.MessageDeduplicationId)
	err := r.dynamodbTemplate.UpdateOneWithTransaction(r.tableName, input)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
