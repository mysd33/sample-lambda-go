/*
dynamodb パッケージは、DynamoDBアクセスに関する機能を提供するパッケージです。
*/
package dynamodb

import (
	"strings"

	"example.com/appbase/pkg/dynamodb/criteria"
	"example.com/appbase/pkg/dynamodb/tables"
	"example.com/appbase/pkg/logging"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/cockroachdb/errors"
)

var (
	ErrRecordNotFound     = errors.New("record not found")
	ErrKeyDuplicaiton     = errors.New("key duplication")
	ErrUpdateWithCondtion = errors.New("update with condition error")
	ErrDeleteWithCondtion = errors.New("delete with condition error")
)

// DynamoDBTemplate は、DynamoDBアクセスを定型化した高次のインタフェースです。
type DynamoDBTemplate interface {
	// CreateOne は、DynamoDBに項目を登録します。
	CreateOne(tableName tables.DynamoDBTableName, inputEntity any) error
	// FindOneByTableKey は、ベーステーブルのプライマリキーの完全一致でDynamoDBから項目を取得します。
	FindOneByTableKey(tableName tables.DynamoDBTableName, input criteria.PkOnlyQueryInput, outEntity any) error
	// FindOneByPrimaryKey は、ベーステーブルのプライマリキーによる条件でDynamoDBから複数件の項目を取得します。
	FindSomeByTableKey(tableName tables.DynamoDBTableName, input criteria.PkOnlyQueryInput, outEntities any) error
	// FindSomeByGSIKey は、GSIのプライマリキーによる条件でDynamoDBから項目を複数件取得します。
	FindSomeByGSIKey(tableName tables.DynamoDBTableName, input criteria.GsiQueryInput, outEntities any) error
	// UpdateOne は、DynamoDBの項目を更新します。
	UpdateOne(tableName tables.DynamoDBTableName, input criteria.UpdateInput) error
	// DeleteOne は、DynamoDBの項目を削除します。
	DeleteOne(tableName tables.DynamoDBTableName, input criteria.DeleteInput) error
}

// NewDynamoDBTemplate は、DynamoDBTemplateのインスタンスを生成します。
func NewDynamoDBTemplate(log logging.Logger, dynamodbAccessor DynamoDBAccessor) DynamoDBTemplate {
	return &defaultDynamoDBTemplate{
		log:              log,
		dynamodbAccessor: dynamodbAccessor,
	}
}

//TODO:　DynamoDBTemplateインタフェースの実装

type defaultDynamoDBTemplate struct {
	log              logging.Logger
	dynamodbAccessor DynamoDBAccessor
}

// CreateOne implements DynamoDBTemplate.
func (t *defaultDynamoDBTemplate) CreateOne(tableName tables.DynamoDBTableName, inputEntity any) error {
	attributes, err := attributevalue.MarshalMap(inputEntity)
	if err != nil {
		return errors.WithStack(err)
	}
	// パーティションキーの重複判定条件
	partitonkeyName := tables.GetPrimaryKey(tableName).PartitionKey
	conditionExpression := aws.String("attribute_not_exists(#partition_key)")
	expressionAttributeNames := map[string]string{
		"#partition_key": partitonkeyName,
	}
	item := &dynamodb.PutItemInput{
		TableName:                aws.String(string(tableName)),
		Item:                     attributes,
		ConditionExpression:      conditionExpression,
		ExpressionAttributeNames: expressionAttributeNames,
	}
	_, err = t.dynamodbAccessor.PutItemSdk(item)
	if err != nil {
		var condErr *types.ConditionalCheckFailedException
		if errors.As(err, &condErr) {
			return ErrKeyDuplicaiton
		}
		return errors.WithStack(err)
	}
	return nil
}

// FindOneByTableKey implements DynamoDBTemplate.
func (t *defaultDynamoDBTemplate) FindOneByTableKey(tableName tables.DynamoDBTableName, input criteria.PkOnlyQueryInput, outEntity any) error {
	// プライマリキーの条件
	keyMap, err := CreatePkAttributeValue(input.PrimarKey)
	if err != nil {
		return err
	}
	// 取得項目
	var projection *string
	if len(input.SelectAttributes) > 0 {
		projection = aws.String(strings.Join(input.SelectAttributes, ","))
	}
	// GetItemInput
	getItemInput := &dynamodb.GetItemInput{
		TableName:            aws.String(string(tableName)),
		Key:                  keyMap,
		ProjectionExpression: projection,
		ConsistentRead:       aws.Bool(input.ConsitentRead),
	}
	// GetItemの実行
	getItemOutput, err := t.dynamodbAccessor.GetItemSdk(getItemInput)
	if err != nil {
		return errors.WithStack(err)
	}
	if len(getItemOutput.Item) == 0 {
		return ErrRecordNotFound
	}
	if err := attributevalue.UnmarshalMap(getItemOutput.Item, &outEntity); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// FindSomeByGSIKey implements DynamoDBTemplate.
func (t *defaultDynamoDBTemplate) FindSomeByGSIKey(tableName tables.DynamoDBTableName, input criteria.GsiQueryInput, outEntities any) error {
	panic("unimplemented")
}

// FindSomeByTableKey implements DynamoDBTemplate.
func (t *defaultDynamoDBTemplate) FindSomeByTableKey(tableName tables.DynamoDBTableName, input criteria.PkOnlyQueryInput, outEntities any) error {
	panic("unimplemented")
}

// UpdateOne implements DynamoDBTemplate.
func (t *defaultDynamoDBTemplate) UpdateOne(tableName tables.DynamoDBTableName, input criteria.UpdateInput) error {
	// プライマリキーの条件
	keyMap, err := CreatePkAttributeValue(input.PrimarKey)
	if err != nil {
		return err
	}
	// 更新表現
	expr, err := CreateUpdateExpressionBuilder(input)
	if err != nil {
		return err
	}
	// UpdateItemInput
	updateItemInput := &dynamodb.UpdateItemInput{
		TableName:                 aws.String(string(tableName)),
		Key:                       keyMap,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		UpdateExpression:          expr.Update(),
		ConditionExpression:       expr.Condition(),
		ReturnValues:              types.ReturnValueAllNew,
	}
	// UpdateItemの実行
	_, err = t.dynamodbAccessor.UpdateItemSdk(updateItemInput)
	if err != nil {
		// 更新条件エラー
		var condErr *types.ConditionalCheckFailedException
		if errors.As(err, &condErr) {
			return ErrUpdateWithCondtion
		}
		return errors.WithStack(err)
	}
	return nil
}

// DeleteOne implements DynamoDBTemplate.
func (t *defaultDynamoDBTemplate) DeleteOne(tableName tables.DynamoDBTableName, input criteria.DeleteInput) error {
	// プライマリキーの条件
	keyMap, err := CreatePkAttributeValue(input.PrimarKey)
	if err != nil {
		return err
	}
	// 削除表現
	expr, err := CreateDeleteExpressionBuilder(input)
	if err != nil {
		return err
	}
	// DeleteItemInput
	deleteItemInput := &dynamodb.DeleteItemInput{
		TableName:                 aws.String(string(tableName)),
		Key:                       keyMap,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ReturnValues:              types.ReturnValueNone,
	}
	// DeleteItemの実行
	_, err = t.dynamodbAccessor.DeleteItemSdk(deleteItemInput)
	if err != nil {
		// 削除条件エラー
		var condErr *types.ConditionalCheckFailedException
		if errors.As(err, &condErr) {
			return ErrDeleteWithCondtion
		}
		return errors.WithStack(err)
	}
	return nil
}
