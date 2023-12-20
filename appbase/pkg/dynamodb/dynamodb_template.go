/*
dynamodb パッケージは、DynamoDBアクセスに関する機能を提供するパッケージです。
*/
package dynamodb

import (
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
	CreateOne(tableName tables.DynamoDBTableName, inputEntity any) error
	FindOneByPrimaryKey(tableName tables.DynamoDBTableName, input criteria.PkOnlyQueryInput, outEntity any) error
	FindSomeByPrimaryKey(tableName tables.DynamoDBTableName, input criteria.PkOnlyQueryInput, outEntities any) error
	FindSomeByGSI(tableName tables.DynamoDBTableName, input criteria.GsiQueryInput, outEntities any) error
	UpdateOne(tableName tables.DynamoDBTableName, input criteria.UpdateInput) error
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
	item, err := attributevalue.MarshalMap(inputEntity)
	if err != nil {
		return errors.WithStack(err)
	}
	// パーティションキーの重複判定条件
	partitonkeyName := tables.GetPrimaryKey(tableName).PartitionKey
	conditionExpression := aws.String("attribute_not_exists(#partition_key)")
	expressionAttributeNames := map[string]string{
		"#partition_key": partitonkeyName,
	}
	input := &dynamodb.PutItemInput{
		TableName:                aws.String(string(tableName)),
		Item:                     item,
		ConditionExpression:      conditionExpression,
		ExpressionAttributeNames: expressionAttributeNames,
	}
	_, err = t.dynamodbAccessor.PutItemSdk(input)
	if err != nil {
		var condErr *types.ConditionalCheckFailedException
		if errors.As(err, &condErr) {
			return ErrKeyDuplicaiton
		}
		return errors.WithStack(err)
	}
	return nil
}

// FindOneByPrimaryKey implements DynamoDBTemplate.
func (t *defaultDynamoDBTemplate) FindOneByPrimaryKey(tableName tables.DynamoDBTableName, input criteria.PkOnlyQueryInput, outEntity any) error {
	panic("unimplemented")
}

// FindSomeByGSI implements DynamoDBTemplate.
func (t *defaultDynamoDBTemplate) FindSomeByGSI(tableName tables.DynamoDBTableName, input criteria.GsiQueryInput, outEntities any) error {
	panic("unimplemented")
}

// FindSomeByPrimaryKey implements DynamoDBTemplate.
func (t *defaultDynamoDBTemplate) FindSomeByPrimaryKey(tableName tables.DynamoDBTableName, input criteria.PkOnlyQueryInput, outEntities any) error {
	panic("unimplemented")
}

// UpdateOne implements DynamoDBTemplate.
func (t *defaultDynamoDBTemplate) UpdateOne(tableName tables.DynamoDBTableName, input criteria.UpdateInput) error {
	panic("unimplemented")
}

// DeleteOne implements DynamoDBTemplate.
func (t *defaultDynamoDBTemplate) DeleteOne(tableName tables.DynamoDBTableName, input criteria.DeleteInput) error {
	panic("unimplemented")
}
