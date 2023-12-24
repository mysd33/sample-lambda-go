/*
transaction パッケージは、トランザクション管理に関する機能を提供するパッケージです。
*/
package transaction

import (
	mydynamodb "example.com/appbase/pkg/dynamodb"
	"example.com/appbase/pkg/dynamodb/input"
	"example.com/appbase/pkg/dynamodb/tables"
	"example.com/appbase/pkg/logging"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/cockroachdb/errors"
)

// TransactionalDynamoDBTemplate は、トランザクション管理対応のDynamoDBアクセスを定型化した高次のインタフェースです。
type TransactinalDynamoDBTemplate interface {
	mydynamodb.DynamoDBTemplate
	// CreateOneWithTransaction は、トランザクションでDynamoDBに項目を登録します。
	CreateOneWithTransaction(tableName tables.DynamoDBTableName, inputEntity any) error
	// UpdateOneWithTransaction は、トランザクションでDynamoDBに項目を更新します。
	UpdateOneWithTransaction(tableName tables.DynamoDBTableName, input input.UpdateInput) error
	// DeleteOneWithTransaction は、トランザクションでDynamoDBに項目を削除します。
	DeleteOneWithTransaction(tableName tables.DynamoDBTableName, input input.DeleteInput) error
}

func NewTransactionalDynamoDBTemplate(log logging.Logger,
	transactionalDynamoDBAccessor TransactionalDynamoDBAccessor) TransactinalDynamoDBTemplate {
	dynamodbTemplate := mydynamodb.NewDynamoDBTemplate(log, transactionalDynamoDBAccessor)
	return &defaultTransactionalDynamoDBTemplate{
		log:                           log,
		dynamodbTemplate:              dynamodbTemplate,
		transactionalDynamoDBAccessor: transactionalDynamoDBAccessor,
	}
}

type defaultTransactionalDynamoDBTemplate struct {
	log                           logging.Logger
	dynamodbTemplate              mydynamodb.DynamoDBTemplate
	transactionalDynamoDBAccessor TransactionalDynamoDBAccessor
}

// CreateOne implements TransactinalDynamoDBTemplate.
func (t *defaultTransactionalDynamoDBTemplate) CreateOne(tableName tables.DynamoDBTableName, inputEntity any) error {
	return t.dynamodbTemplate.CreateOne(tableName, inputEntity)
}

// FindOneByTableKey implements TransactinalDynamoDBTemplate.
func (t *defaultTransactionalDynamoDBTemplate) FindOneByTableKey(tableName tables.DynamoDBTableName, input input.PkOnlyQueryInput, outEntity any) error {
	return t.dynamodbTemplate.FindOneByTableKey(tableName, input, outEntity)
}

// FindSomeByTableKey implements TransactinalDynamoDBTemplate.
func (t *defaultTransactionalDynamoDBTemplate) FindSomeByTableKey(tableName tables.DynamoDBTableName, input input.PkQueryInput, outEntities any) error {
	return t.dynamodbTemplate.FindSomeByTableKey(tableName, input, outEntities)
}

// FindSomeByGSIKey implements TransactinalDynamoDBTemplate.
func (t *defaultTransactionalDynamoDBTemplate) FindSomeByGSIKey(tableName tables.DynamoDBTableName, input input.GsiQueryInput, outEntities any) error {
	return t.dynamodbTemplate.FindSomeByGSIKey(tableName, input, outEntities)
}

// UpdateOne implements TransactinalDynamoDBTemplate.
func (t *defaultTransactionalDynamoDBTemplate) UpdateOne(tableName tables.DynamoDBTableName, input input.UpdateInput) error {
	return t.dynamodbTemplate.UpdateOne(tableName, input)
}

// DeleteOne implements TransactinalDynamoDBTemplate.
func (t *defaultTransactionalDynamoDBTemplate) DeleteOne(tableName tables.DynamoDBTableName, input input.DeleteInput) error {
	return t.dynamodbTemplate.DeleteOne(tableName, input)
}

// CreateOneWithTransaction implements TransactinalDynamoDBTemplate.
func (t *defaultTransactionalDynamoDBTemplate) CreateOneWithTransaction(tableName tables.DynamoDBTableName, inputEntity any) error {
	attributes, err := attributevalue.MarshalMap(inputEntity)
	if err != nil {
		return errors.Wrap(err, "CreateOneWithTransactionで構造体をAttributeValueのMap変換時にエラー")
	}
	// パーティションキーの重複判定条件
	partitonkeyName := tables.GetPrimaryKey(tableName).PartitionKey
	conditionExpression := aws.String("attribute_not_exists(#partition_key)")
	expressionAttributeNames := map[string]string{
		"#partition_key": partitonkeyName,
	}
	// TransactWriteItemの作成
	item := types.TransactWriteItem{
		Put: &types.Put{
			TableName:                aws.String(string(tableName)),
			Item:                     attributes,
			ConditionExpression:      conditionExpression,
			ExpressionAttributeNames: expressionAttributeNames,
		},
	}
	// TransactWriteItemの追加
	t.transactionalDynamoDBAccessor.AppendTransactWriteItem(&item)
	return nil
}

// UpdateOneWithTransaction implements TransactinalDynamoDBTemplate.
func (t *defaultTransactionalDynamoDBTemplate) UpdateOneWithTransaction(tableName tables.DynamoDBTableName, input input.UpdateInput) error {
	// プライマリキーの条件
	keyMap, err := mydynamodb.CreatePkAttributeValue(input.PrimaryKey)
	if err != nil {
		return errors.Wrap(err, "UpdateOneWithTransactionで更新対象条件の生成時エラー")
	}
	// 更新表現
	expr, err := mydynamodb.CreateUpdateExpression(input)
	if err != nil {
		return errors.Wrap(err, "UpdateOneWithTransactionで更新条件の生成時エラー")
	}
	// TransactWriteItemの作成
	item := types.TransactWriteItem{
		Update: &types.Update{
			TableName:                 aws.String(string(tableName)),
			Key:                       keyMap,
			ExpressionAttributeNames:  expr.Names(),
			ExpressionAttributeValues: expr.Values(),
			UpdateExpression:          expr.Update(),
			ConditionExpression:       expr.Condition(),
		},
	}
	// TransactWriteItemの追加
	t.transactionalDynamoDBAccessor.AppendTransactWriteItem(&item)
	return nil
}

// DeleteOneWithTransaction implements TransactinalDynamoDBTemplate.
func (t *defaultTransactionalDynamoDBTemplate) DeleteOneWithTransaction(tableName tables.DynamoDBTableName, input input.DeleteInput) error {
	// プライマリキーの条件
	keyMap, err := mydynamodb.CreatePkAttributeValue(input.PrimaryKey)
	if err != nil {
		return errors.Wrap(err, "DeleteOneWithTransactionで削除対象条件の生成時エラー")
	}
	// 削除表現
	expr, err := mydynamodb.CreateDeleteExpression(input)
	if err != nil {
		return errors.Wrap(err, "DelteOneWithTransactionで削除条件の生成時エラー")
	}
	// TransactWriteItemの作成
	item := types.TransactWriteItem{
		Delete: &types.Delete{
			TableName:                 aws.String(string(tableName)),
			Key:                       keyMap,
			ExpressionAttributeNames:  expr.Names(),
			ExpressionAttributeValues: expr.Values(),
		},
	}
	// TransactWriteItemの追加
	t.transactionalDynamoDBAccessor.AppendTransactWriteItem(&item)
	return nil
}
