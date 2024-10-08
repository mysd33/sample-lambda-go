package transaction

import (
	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/constant"
	"example.com/appbase/pkg/dynamodb"
	"example.com/appbase/pkg/dynamodb/input"
	"example.com/appbase/pkg/dynamodb/tables"
	"example.com/appbase/pkg/logging"
	"example.com/appbase/pkg/transaction/model"
	mytables "example.com/appbase/pkg/transaction/tables"
	"github.com/cockroachdb/errors"
)

// キューメッセージ管理テーブル名のプロパティ名
const QUEUE_MESSAGE_TABLE_NAME = "QUEUE_MESSAGE_TABLE_NAME"

// QueueMessageItemRepository は、キューメッセージ管理テーブルのリポジトリインタフェースです。
type QueueMessageItemRepository interface {
	FindOne(messageId string, deleteTime int) (*model.QueueMessageItem, error)
	CreateOneWithTx(queueMessage *model.QueueMessageItem) error
	UpdateOneWithTx(queueMessage *model.QueueMessageItem) error
}

// NewQueueMessageItemRepository は、QueueMessageItemRepositoryを作成します。
func NewQueueMessageItemRepository(config config.Config,
	logger logging.Logger,
	dynamodbTemplate TransactionalDynamoDBTemplate) QueueMessageItemRepository {
	// テーブル名取得
	tableName := tables.DynamoDBTableName(config.Get(QUEUE_MESSAGE_TABLE_NAME, "queue_message"))
	// テーブル定義の設定
	mytables.QueueMessageTable{}.InitPK(tableName)
	// プライマリキーの設定
	primaryKey := tables.GetPrimaryKey(tableName)
	return &defaultQueueMessageItemRepository{
		logger:           logger,
		dynamodbTemplate: dynamodbTemplate,
		tableName:        tableName,
		primaryKey:       primaryKey,
	}
}

// defaultQueueMessageItemRepository は、QueueMessageItemRepositoryを実装する構造体です。
type defaultQueueMessageItemRepository struct {
	logger           logging.Logger
	dynamodbTemplate TransactionalDynamoDBTemplate
	tableName        tables.DynamoDBTableName
	primaryKey       *tables.PKKeyPair
}

// FindOne implements QueueMessageItemRepository.
func (r *defaultQueueMessageItemRepository) FindOne(messageId string, deleteTime int) (*model.QueueMessageItem, error) {
	r.logger.Debug("partitionKey: %s", r.primaryKey.PartitionKey)
	input := input.PKQueryInput{
		PrimaryKey: input.PrimaryKey{
			PartitionKey: input.Attribute{
				Name:  r.primaryKey.PartitionKey,
				Value: messageId,
			},
		},
		// 念のため、delete_timeが一致してるかのFilterのチェック
		WhereClauses: []*input.WhereClause{
			{
				Attribute: input.Attribute{
					Name:  constant.QUEUE_MESSAGE_DELETE_TIME_NAME,
					Value: deleteTime,
				},
				WhereOp: input.WHERE_EQUAL,
			},
		},
		ConsitentRead: true,
	}
	var queueMessageItems []model.QueueMessageItem
	// Itemの取得
	err := r.dynamodbTemplate.FindSomeByTableKey(r.tableName, input, &queueMessageItems)
	if err != nil {
		if errors.Is(err, dynamodb.ErrRecordNotFound) {
			return &model.QueueMessageItem{}, nil
		}
		return nil, err
	}
	return &queueMessageItems[0], nil
}

// CreateOneWithTx implements QueueMessageItemRepository.
func (r *defaultQueueMessageItemRepository) CreateOneWithTx(queueMessage *model.QueueMessageItem) error {
	err := r.dynamodbTemplate.CreateOneWithTransaction(r.tableName, queueMessage)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// UpdateOneWithTx implements QueueMessageItemRepository.
func (r *defaultQueueMessageItemRepository) UpdateOneWithTx(queueMessage *model.QueueMessageItem) error {
	r.logger.Debug("partitionKey: %s", r.primaryKey.PartitionKey)
	input := input.UpdateInput{
		PrimaryKey: input.PrimaryKey{
			PartitionKey: input.Attribute{
				Name:  r.primaryKey.PartitionKey,
				Value: queueMessage.MessageId,
			},
		},
		UpdateAttributes: []*input.Attribute{
			// Status列を更新
			{
				Name:  constant.QUEUE_MESSAGE_STATUS,
				Value: queueMessage.Status,
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
	r.logger.Debug("ステータス: %s", queueMessage.Status)
	err := r.dynamodbTemplate.UpdateOneWithTransaction(r.tableName, input)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
