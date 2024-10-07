/*
transaction パッケージは、トランザクション管理に関する機能を提供するパッケージです。
*/
package transaction

import (
	"context"

	"example.com/appbase/pkg/apcontext"
	mydynamodb "example.com/appbase/pkg/dynamodb"
	"example.com/appbase/pkg/dynamodb/input"
	"example.com/appbase/pkg/dynamodb/tables"
	"example.com/appbase/pkg/logging"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/cockroachdb/errors"
)

// TransactionalDynamoDBTemplate は、トランザクション管理対応のDynamoDBアクセスを定型化した高次のインタフェースです。
type TransactionalDynamoDBTemplate interface {
	mydynamodb.DynamoDBTemplate
	// CreateOneWithTransaction は、トランザクションでDynamoDBに項目を登録します。
	CreateOneWithTransaction(tableName tables.DynamoDBTableName, inputEntity any) error
	// CreateOneWithTransactionInContext は、goroutine向けに渡されたContextを利用して、トランザクションでDynamoDBに項目を登録します。
	CreateOneWithTransactionInContext(ctx context.Context, tableName tables.DynamoDBTableName, inputEntity any) error
	// UpdateOneWithTransaction は、トランザクションでDynamoDBに項目を更新します。
	UpdateOneWithTransaction(tableName tables.DynamoDBTableName, input input.UpdateInput) error
	// UpdateOneWithTransactionInContext は、goroutine向けに渡されたContextを利用して、トランザクションでDynamoDBに項目を更新します。
	UpdateOneWithTransactionInContext(ctx context.Context, tableName tables.DynamoDBTableName, input input.UpdateInput) error
	// DeleteOneWithTransaction は、トランザクションでDynamoDBに項目を削除します。
	DeleteOneWithTransaction(tableName tables.DynamoDBTableName, input input.DeleteInput) error
	// DeleteOneWithTransactionInContext は、goroutine向けに渡されたContextを利用して、トランザクションでDynamoDBに項目を削除します。
	DeleteOneWithTransactionInContext(ctx context.Context, tableName tables.DynamoDBTableName, input input.DeleteInput) error
}

func NewTransactionalDynamoDBTemplate(logger logging.Logger,
	transactionalDynamoDBAccessor TransactionalDynamoDBAccessor) TransactionalDynamoDBTemplate {
	dynamodbTemplate := mydynamodb.NewDynamoDBTemplate(logger, transactionalDynamoDBAccessor)
	return &defaultTransactionalDynamoDBTemplate{
		logger:                        logger,
		dynamodbTemplate:              dynamodbTemplate,
		transactionalDynamoDBAccessor: transactionalDynamoDBAccessor,
	}
}

type defaultTransactionalDynamoDBTemplate struct {
	logger                        logging.Logger
	dynamodbTemplate              mydynamodb.DynamoDBTemplate
	transactionalDynamoDBAccessor TransactionalDynamoDBAccessor
}

// CreateOne implements TransactinalDynamoDBTemplate.
func (t *defaultTransactionalDynamoDBTemplate) CreateOne(tableName tables.DynamoDBTableName, inputEntity any, optFns ...func(*dynamodb.Options)) error {
	return t.dynamodbTemplate.CreateOne(tableName, inputEntity, optFns...)
}

// CreateOneWithContext implements TransactionalDynamoDBTemplate.
func (t *defaultTransactionalDynamoDBTemplate) CreateOneWithContext(ctx context.Context, tableName tables.DynamoDBTableName, inputEntity any, optFns ...func(*dynamodb.Options)) error {
	return t.dynamodbTemplate.CreateOneWithContext(ctx, tableName, inputEntity, optFns...)
}

// FindOneByTableKey implements TransactinalDynamoDBTemplate.
func (t *defaultTransactionalDynamoDBTemplate) FindOneByTableKey(tableName tables.DynamoDBTableName, input input.PKOnlyQueryInput, outEntity any, optFns ...func(*dynamodb.Options)) error {
	return t.dynamodbTemplate.FindOneByTableKey(tableName, input, outEntity, optFns...)
}

// FindOneByTableKeyWithContext implements TransactionalDynamoDBTemplate.
func (t *defaultTransactionalDynamoDBTemplate) FindOneByTableKeyWithContext(ctx context.Context, tableName tables.DynamoDBTableName, input input.PKOnlyQueryInput, outEntity any, optFns ...func(*dynamodb.Options)) error {
	return t.dynamodbTemplate.FindOneByTableKeyWithContext(ctx, tableName, input, outEntity, optFns...)
}

// FindSomeByTableKey implements TransactinalDynamoDBTemplate.
func (t *defaultTransactionalDynamoDBTemplate) FindSomeByTableKey(tableName tables.DynamoDBTableName, input input.PKQueryInput, outEntities any, optFns ...func(*dynamodb.Options)) error {
	return t.dynamodbTemplate.FindSomeByTableKey(tableName, input, outEntities, optFns...)
}

// FindSomeByTableKeyWithContext implements TransactionalDynamoDBTemplate.
func (t *defaultTransactionalDynamoDBTemplate) FindSomeByTableKeyWithContext(ctx context.Context, tableName tables.DynamoDBTableName, input input.PKQueryInput, outEntities any, optFns ...func(*dynamodb.Options)) error {
	return t.dynamodbTemplate.FindSomeByTableKeyWithContext(ctx, tableName, input, outEntities, optFns...)
}

// FindSomeByGSIKey implements TransactinalDynamoDBTemplate.
func (t *defaultTransactionalDynamoDBTemplate) FindSomeByGSIKey(tableName tables.DynamoDBTableName, input input.GsiQueryInput, outEntities any, optFns ...func(*dynamodb.Options)) error {
	return t.dynamodbTemplate.FindSomeByGSIKey(tableName, input, outEntities, optFns...)
}

// FindSomeByGSIKeyWithContext implements TransactionalDynamoDBTemplate.
func (t *defaultTransactionalDynamoDBTemplate) FindSomeByGSIKeyWithContext(ctx context.Context, tableName tables.DynamoDBTableName, input input.GsiQueryInput, outEntities any, optFns ...func(*dynamodb.Options)) error {
	return t.dynamodbTemplate.FindSomeByGSIKeyWithContext(ctx, tableName, input, outEntities, optFns...)
}

// UpdateOne implements TransactinalDynamoDBTemplate.
func (t *defaultTransactionalDynamoDBTemplate) UpdateOne(tableName tables.DynamoDBTableName, input input.UpdateInput, optFns ...func(*dynamodb.Options)) error {
	return t.dynamodbTemplate.UpdateOne(tableName, input, optFns...)
}

// UpdateOneWithContext implements TransactionalDynamoDBTemplate.
func (t *defaultTransactionalDynamoDBTemplate) UpdateOneWithContext(ctx context.Context, tableName tables.DynamoDBTableName, input input.UpdateInput, optFns ...func(*dynamodb.Options)) error {
	return t.dynamodbTemplate.UpdateOneWithContext(ctx, tableName, input, optFns...)
}

// DeleteOne implements TransactinalDynamoDBTemplate.
func (t *defaultTransactionalDynamoDBTemplate) DeleteOne(tableName tables.DynamoDBTableName, input input.DeleteInput, optFns ...func(*dynamodb.Options)) error {
	return t.dynamodbTemplate.DeleteOne(tableName, input, optFns...)
}

// DeleteOneWithContext implements TransactionalDynamoDBTemplate.
func (t *defaultTransactionalDynamoDBTemplate) DeleteOneWithContext(ctx context.Context, tableName tables.DynamoDBTableName, input input.DeleteInput, optFns ...func(*dynamodb.Options)) error {
	return t.dynamodbTemplate.DeleteOneWithContext(ctx, tableName, input, optFns...)
}

// CreateOneWithTransaction implements TransactinalDynamoDBTemplate.
func (t *defaultTransactionalDynamoDBTemplate) CreateOneWithTransaction(tableName tables.DynamoDBTableName, inputEntity any) error {
	return t.CreateOneWithTransactionInContext(apcontext.Context, tableName, inputEntity)
}

// CreateOneWithTransactionInContext implements TransactionalDynamoDBTemplate.
func (t *defaultTransactionalDynamoDBTemplate) CreateOneWithTransactionInContext(ctx context.Context, tableName tables.DynamoDBTableName, inputEntity any) error {
	// TransactWriteItemの作成
	item, err := t.newPutTransactionWriteItem(tableName, inputEntity)
	if err != nil {
		return err
	}
	// TransactWriteItemの追加
	return t.transactionalDynamoDBAccessor.AppendTransactWriteItemWithContext(ctx, item)

}

// newPutTransactionWriteItem は、PutのためのTransactWriteItemを作成します。
func (t *defaultTransactionalDynamoDBTemplate) newPutTransactionWriteItem(tableName tables.DynamoDBTableName, inputEntity any) (*types.TransactWriteItem, error) {
	attributes, err := attributevalue.MarshalMap(inputEntity)
	if err != nil {
		return nil, errors.Wrap(err, "CreateOneWithTransactionで構造体をAttributeValueのMap変換時にエラー")
	}
	// パーティションキーの重複判定条件
	partitonkeyName := tables.GetPrimaryKey(tableName).PartitionKey
	conditionExpression := aws.String("attribute_not_exists(#partition_key)")
	expressionAttributeNames := map[string]string{
		"#partition_key": partitonkeyName,
	}
	// TransactWriteItem
	item := types.TransactWriteItem{
		Put: &types.Put{
			TableName:                aws.String(string(tableName)),
			Item:                     attributes,
			ConditionExpression:      conditionExpression,
			ExpressionAttributeNames: expressionAttributeNames,
		},
	}
	return &item, nil
}

// UpdateOneWithTransaction implements TransactinalDynamoDBTemplate.
func (t *defaultTransactionalDynamoDBTemplate) UpdateOneWithTransaction(tableName tables.DynamoDBTableName, input input.UpdateInput) error {
	return t.CreateOneWithTransactionInContext(apcontext.Context, tableName, input)
}

// UpdateOneWithTransactionInContext implements TransactionalDynamoDBTemplate.
func (t *defaultTransactionalDynamoDBTemplate) UpdateOneWithTransactionInContext(ctx context.Context, tableName tables.DynamoDBTableName, input input.UpdateInput) error {
	item, err := t.newUpdateTransactionWriteItem(tableName, input)
	if err != nil {
		return err
	}
	// TransactWriteItemの追加
	return t.transactionalDynamoDBAccessor.AppendTransactWriteItemWithContext(ctx, item)
}

// newUpdateTransactionWriteItem は、UpdateのためのTransactWriteItemを作成します。
func (t *defaultTransactionalDynamoDBTemplate) newUpdateTransactionWriteItem(tableName tables.DynamoDBTableName, input input.UpdateInput) (*types.TransactWriteItem, error) {
	// プライマリキーの条件
	keyMap, err := mydynamodb.CreatePkAttributeValue(input.PrimaryKey)
	if err != nil {
		return nil, errors.Wrap(err, "UpdateOneWithTransactionで更新対象条件の生成時エラー")
	}
	// 更新表現
	expr, err := mydynamodb.CreateUpdateExpression(input)
	if err != nil {
		return nil, errors.Wrap(err, "UpdateOneWithTransactionで更新条件の生成時エラー")
	}
	// TransactWriteItem
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
	return &item, nil
}

// DeleteOneWithTransaction implements TransactinalDynamoDBTemplate.
func (t *defaultTransactionalDynamoDBTemplate) DeleteOneWithTransaction(tableName tables.DynamoDBTableName, input input.DeleteInput) error {
	return t.DeleteOneWithTransactionInContext(apcontext.Context, tableName, input)
}

// DeleteOneWithTransactionInContext implements TransactionalDynamoDBTemplate.
func (t *defaultTransactionalDynamoDBTemplate) DeleteOneWithTransactionInContext(ctx context.Context, tableName tables.DynamoDBTableName, input input.DeleteInput) error {
	item, err := t.newDeleteTransactionWriteItem(tableName, input)
	if err != nil {
		return err
	}
	// TransactWriteItemの追加
	return t.transactionalDynamoDBAccessor.AppendTransactWriteItemWithContext(ctx, item)
}

// newDeleteTransactionWriteItem は、DeleteのためのTransactWriteItemを作成します。
func (t *defaultTransactionalDynamoDBTemplate) newDeleteTransactionWriteItem(tableName tables.DynamoDBTableName, input input.DeleteInput) (*types.TransactWriteItem, error) {
	// プライマリキーの条件
	keyMap, err := mydynamodb.CreatePkAttributeValue(input.PrimaryKey)
	if err != nil {
		return nil, errors.Wrap(err, "DeleteOneWithTransactionで削除対象条件の生成時エラー")
	}
	// 削除表現
	expr, err := mydynamodb.CreateDeleteExpression(input)
	if err != nil {
		return nil, errors.Wrap(err, "DelteOneWithTransactionで削除条件の生成時エラー")
	}
	// TransactWriteItemの作成
	item := types.TransactWriteItem{
		Delete: &types.Delete{
			TableName:                 aws.String(string(tableName)),
			Key:                       keyMap,
			ExpressionAttributeNames:  expr.Names(),
			ExpressionAttributeValues: expr.Values(),
			ConditionExpression:       expr.Condition(),
		},
	}
	return &item, nil
}
