/*
idempotency パッケージは、イベントの重複によるLambdaの二重実行を防止し冪等性を担保するための機能を提供します。
*/
package idempotency

import (
	"fmt"
	"strconv"
	"time"

	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/date"
	mydynamodb "example.com/appbase/pkg/dynamodb"
	"example.com/appbase/pkg/dynamodb/input"
	"example.com/appbase/pkg/dynamodb/tables"
	"example.com/appbase/pkg/idempotency/entity"
	mytables "example.com/appbase/pkg/idempotency/tables"
	"example.com/appbase/pkg/logging"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go/aws"
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
func NewIdempotencyRepository(log logging.Logger, dynamodbAccessor mydynamodb.DynamoDBAccessor,
	dynamodbTemplate mydynamodb.DynamoDBTemplate, dateManager date.DateManager, config config.Config) IdempotencyRepository {
	// テーブル名取得
	tableName := tables.DynamoDBTableName(config.Get(IDEMPOTENCY_TABLE_NAME, "idempotency"))
	// テーブル定義の設定
	mytables.IdempotencyTable{}.InitPK(tableName)
	// プライマリキーの設定
	primaryKey := tables.GetPrimaryKey(tableName)
	return &defaultIdempotencyRepository{
		log:              log,
		dynamodbTemplate: dynamodbTemplate,
		dateManager:      dateManager,
		tableName:        tableName,
		primaryKey:       primaryKey,
	}
}

// defaultIdempotencyRepository は、DuplicationCheckRepositoryのデフォルト実装です。
type defaultIdempotencyRepository struct {
	log              logging.Logger
	dynamodbAccessor mydynamodb.DynamoDBAccessor
	dynamodbTemplate mydynamodb.DynamoDBTemplate
	dateManager      date.DateManager
	tableName        tables.DynamoDBTableName
	primaryKey       *tables.PKKeyPair
}

// FindOne implements DuplicationCheckRepository.
func (r *defaultIdempotencyRepository) FindOne(idempotencyKey string) (*entity.IdempotencyItem, error) {
	r.log.Debug("partitionKey: %s", r.primaryKey.PartitionKey)
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
	idempotencyKeyNotExist := "attribute_not_exists(#idempotencyKey)"
	// 2. アイテムの有効期限（TTL）expiryが過ぎている
	idempotencyExpiryExpired := "#expiry < :now"
	// 3. ステータスが処理中のアイテムの処理中状態の有効期限が過ぎている
	inprogressExpiryExpired := "#status = :inprogress AND attribute_not_exists(#inprogressExpiry) AND #inprogressExpiry < :nowInMillis"
	conditionExpression := fmt.Sprintf("(%s) OR (%s) OR (%s)", idempotencyKeyNotExist, idempotencyExpiryExpired, inprogressExpiryExpired)

	attributes, err := attributevalue.MarshalMap(idempotencyItem)
	if err != nil {
		return errors.Wrap(err, "CreateOneで構造体をAttributeValueのMap変換時にエラー")
	}

	// TODO: DynamoDBTemplate化して呼び出すようにしたい

	input := &dynamodb.PutItemInput{
		TableName:           aws.String(string(r.tableName)),
		Item:                attributes,
		ConditionExpression: aws.String(conditionExpression),
		ExpressionAttributeNames: map[string]string{
			"#idempotencyKey":   r.primaryKey.PartitionKey,
			"#expiry":           mytables.EXPIRY,
			"#inprogressExpiry": mytables.INPROGRESS_EXPIRY,
			"#status":           mytables.STATUS,
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":now":         &types.AttributeValueMemberN{Value: strconv.FormatInt(now.Unix(), 10)},
			":nowInMillis": &types.AttributeValueMemberN{Value: strconv.FormatInt(now.UnixNano()/int64(time.Millisecond), 10)},
			":inprogress":  &types.AttributeValueMemberS{Value: mytables.STATUS_INPROGRESS},
		},
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
