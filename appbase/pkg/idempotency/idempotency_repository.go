/*
idempotency パッケージは、イベントの重複によるLambdaの二重実行を防止し冪等性を担保するための機能を提供します。
*/
package idempotency

import (
	"time"

	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/date"
	mydynamodb "example.com/appbase/pkg/dynamodb"
	"example.com/appbase/pkg/dynamodb/input"
	"example.com/appbase/pkg/dynamodb/tables"
	"example.com/appbase/pkg/idempotency/entity"
	mytables "example.com/appbase/pkg/idempotency/tables"
	"example.com/appbase/pkg/logging"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/cockroachdb/errors"
)

// 冪等性管理テーブル名のプロパティ名
const IDEMPOTENCY_TABLE_NAME = "IDEMPOTENCY_TABLE_NAME"

// IdempotencyRepository は、冪等性管理テーブルのためのリポジトリインターフェースです。
type IdempotencyRepository interface {
	// FindOne は、冪等性テーブルからアイテムを取得します。
	FindOne(idempotencyKey string) (*entity.IdempotencyItem, error)
	// CreateOne は、冪等性テーブルにアイテムを作成します。
	// 既に同一のidempotencyKeyが存在する場合はエラーを返します。
	CreateOne(idempotencyItem *entity.IdempotencyItem) error
	// UpdateOne は、冪等性テーブルのアイテムを更新します。
	UpdateOne(idempotencyItem *entity.IdempotencyItem) error
	// DeleteOne は、冪等性テーブルのアイテムを削除します。
	DeleteOne(idempotencyKey string) error
}

// NewIdempotencyRepository は、IdempotencyRepositoryを作成します。
func NewIdempotencyRepository(logger logging.Logger, dynamodbAccessor mydynamodb.DynamoDBAccessor,
	dynamodbTemplate mydynamodb.DynamoDBTemplate, dateManager date.DateManager, config config.Config) IdempotencyRepository {
	// テーブル名取得
	tableName := tables.DynamoDBTableName(config.Get(IDEMPOTENCY_TABLE_NAME, "idempotency"))
	// テーブル定義の設定
	mytables.IdempotencyTable{}.InitPK(tableName)
	// プライマリキーの設定
	primaryKey := tables.GetPrimaryKey(tableName)
	return &defaultIdempotencyRepository{
		logger:           logger,
		dynamodbAccessor: dynamodbAccessor,
		dynamodbTemplate: dynamodbTemplate,
		dateManager:      dateManager,
		tableName:        tableName,
		primaryKey:       primaryKey,
	}
}

// defaultIdempotencyRepository は、DuplicationCheckRepositoryのデフォルト実装です。
type defaultIdempotencyRepository struct {
	logger           logging.Logger
	dynamodbAccessor mydynamodb.DynamoDBAccessor
	dynamodbTemplate mydynamodb.DynamoDBTemplate
	dateManager      date.DateManager
	tableName        tables.DynamoDBTableName
	primaryKey       *tables.PKKeyPair
}

// FindOne implements DuplicationCheckRepository.
func (r *defaultIdempotencyRepository) FindOne(idempotencyKey string) (*entity.IdempotencyItem, error) {
	r.logger.Debug("partitionKey: %s", r.primaryKey.PartitionKey)
	input := input.PKOnlyQueryInput{
		PrimaryKey: input.PrimaryKey{
			PartitionKey: input.Attribute{
				Name:  r.primaryKey.PartitionKey,
				Value: idempotencyKey,
			},
		},
	}
	var idempotencyItem entity.IdempotencyItem
	err := r.dynamodbTemplate.FindOneByTableKey(r.tableName, input, &idempotencyItem)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &idempotencyItem, nil
}

// CreateOne implements DuplicationCheckRepository.
func (r *defaultIdempotencyRepository) CreateOne(idempotencyItem *entity.IdempotencyItem) error {
	now := r.dateManager.GetSystemDate()
	// 以下の条件のいずれかが満たされた場合は、新しいアイテムを作成する
	// 1. idempotencyKeyが存在しない
	idempotencyKeyNotExistExpr := expression.AttributeNotExists(expression.Name(mytables.IDEMPOTENCY_KEY))
	// 2. アイテムの有効期限（TTL）expiryが過ぎている
	idempotencyExpiryExpiredExpr := expression.Name(mytables.EXPIRY).LessThan(expression.Value(now.Unix()))
	// 3. ステータスが処理中のアイテムの処理中状態の有効期限が過ぎている
	inprogressExpiryExpiredExpr := expression.Name(mytables.STATUS).Equal(expression.Value(mytables.STATUS_INPROGRESS)).
		And(expression.AttributeExists(expression.Name(mytables.INPROGRESS_EXPIRY))).
		And(expression.Name(mytables.INPROGRESS_EXPIRY).LessThan(expression.Value(now.UnixNano() / int64(time.Millisecond))))
	// 1、2、 3の条件をORで結合
	conditionExpressionExpr := expression.Or(idempotencyKeyNotExistExpr, idempotencyExpiryExpiredExpr, inprogressExpiryExpiredExpr)
	expr, err := expression.NewBuilder().WithCondition(conditionExpressionExpr).Build()
	if err != nil {
		return errors.Wrap(err, "CreateOneでConditionExpressionの構築時にエラー")
	}
	attributes, err := attributevalue.MarshalMap(idempotencyItem)
	if err != nil {
		return errors.Wrap(err, "CreateOneで構造体をAttributeValueのMap変換時にエラー")
	}
	input := &dynamodb.PutItemInput{
		TableName:                 aws.String(string(r.tableName)),
		Item:                      attributes,
		ConditionExpression:       expr.Condition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}
	_, err = r.dynamodbAccessor.PutItemSdk(input)
	if err != nil {
		var condErr *types.ConditionalCheckFailedException
		if errors.As(err, &condErr) {
			return mydynamodb.ErrKeyDuplicaiton
		}
		return errors.Wrap(err, "CreateOneで登録実行時エラー")
	}
	return nil
}

// UpdateOne implements DuplicationCheckRepository.
func (r *defaultIdempotencyRepository) UpdateOne(idempotencyItem *entity.IdempotencyItem) error {
	input := input.UpdateInput{
		PrimaryKey: input.PrimaryKey{
			PartitionKey: input.Attribute{
				Name:  r.primaryKey.PartitionKey,
				Value: idempotencyItem.IdempotencyKey,
			},
		},
		UpdateAttributes: []*input.Attribute{
			// 有効期限を更新
			{
				Name:  mytables.EXPIRY,
				Value: idempotencyItem.Expiry,
			},
			// ステータス列を更新
			{
				Name:  mytables.STATUS,
				Value: idempotencyItem.Status,
			},
		},
	}

	err := r.dynamodbTemplate.UpdateOne(r.tableName, input)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// DeleteOne implements DuplicationCheckRepository.
func (r *defaultIdempotencyRepository) DeleteOne(idempotencyKey string) error {
	input := input.DeleteInput{
		PrimaryKey: input.PrimaryKey{
			PartitionKey: input.Attribute{
				Name:  r.primaryKey.PartitionKey,
				Value: idempotencyKey,
			},
		},
	}
	err := r.dynamodbTemplate.DeleteOne(r.tableName, input)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
