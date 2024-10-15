/*
transaction パッケージは、トランザクション管理に関する機能を提供するパッケージです。
*/
package transaction

import (
	"context"

	"example.com/appbase/pkg/apcontext"
	myConfig "example.com/appbase/pkg/config"
	myDynamoDB "example.com/appbase/pkg/dynamodb"
	"example.com/appbase/pkg/id"
	"example.com/appbase/pkg/logging"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/cockroachdb/errors"
)

// TransactionalDynamoDBAccessorは、トランザクション管理可能なDynamoDBアクセス用インタフェースです。
type TransactionalDynamoDBAccessor interface {
	myDynamoDB.DynamoDBAccessor
	// AppendTransactWriteItem は、トランザクション書き込みしたい場合に対象のTransactWriteItemを追加します。
	// なお、TransactWriteItemsの実行は、TransactionManagerのExecuteTransaction関数で実行されるdomain.ServiceFunc関数が終了する際にtransactionWriteItemsSDKを実施します。
	AppendTransactWriteItem(item *types.TransactWriteItem) error
	// AppendTransactWriteItemWithContext は、goroutine向けに渡されたContextを利用して、トランザクション書き込みしたい場合に対象のTransactWriteItemを追加します。
	// なお、TransactWriteItemsの実行は、TransactionManagerのExecuteTransactionWithContext関数で実行されるdomain.ServiceFuncWithContext関数が終了する際にtransactionWriteItemsSDKを実施します。
	AppendTransactWriteItemWithContext(ctx context.Context, item *types.TransactWriteItem) error
	// TransactWriteItemsSDK は、AWS SDKによるTransactWriteItemsをラップします。
	// なお、TransactWriteItemsの実行は、TransactionManagerが実行するため業務ロジックで利用する必要はありません。
	TransactWriteItemsSDK(items []types.TransactWriteItem, optFns ...func(*dynamodb.Options)) (*dynamodb.TransactWriteItemsOutput, error)
	// TransactWriteItemsSDKWithContext は、AWS SDKによるTransactWriteItemsをラップします。goroutine向けに、渡されたContextを利用して実行します。
	TransactWriteItemsSDKWithContext(ctx context.Context, items []types.TransactWriteItem, optFns ...func(*dynamodb.Options)) (*dynamodb.TransactWriteItemsOutput, error)
}

// NewTransactionalDynamoDBAccessor は、TransactionalDynamoDBAccessorを作成します。
func NewTransactionalDynamoDBAccessor(logger logging.Logger, myCfg myConfig.Config) (TransactionalDynamoDBAccessor, error) {
	dynamodbAccessor, err := myDynamoDB.NewDynamoDBAccessor(logger, myCfg)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &defaultTransactionalDynamoDBAccessor{
		logger:           logger,
		config:           myCfg,
		dynamodbAccessor: dynamodbAccessor,
	}, nil
}

type defaultTransactionalDynamoDBAccessor struct {
	logger           logging.Logger
	config           myConfig.Config
	idGenerator      id.IDGenerator
	dynamodbAccessor myDynamoDB.DynamoDBAccessor
}

// GetDynamoDBClient implements TransactionalDynamoDBAccessor.
func (da *defaultTransactionalDynamoDBAccessor) GetDynamoDBClient() *dynamodb.Client {
	return da.dynamodbAccessor.GetDynamoDBClient()
}

// GetItemSdk implements TransactionalDynamoDBAccessor.
func (da *defaultTransactionalDynamoDBAccessor) GetItemSdk(input *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	return da.dynamodbAccessor.GetItemSdk(input, optFns...)
}

// GetItemSdkWithContext implements TransactionalDynamoDBAccessor.
func (da *defaultTransactionalDynamoDBAccessor) GetItemSdkWithContext(ctx context.Context, input *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	return da.dynamodbAccessor.GetItemSdkWithContext(ctx, input, optFns...)
}

// PutItemSdk implements TransactionalDynamoDBAccessor.
func (da *defaultTransactionalDynamoDBAccessor) PutItemSdk(input *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	return da.dynamodbAccessor.PutItemSdk(input, optFns...)
}

// PutItemSdkWithContext implements TransactionalDynamoDBAccessor.
func (da *defaultTransactionalDynamoDBAccessor) PutItemSdkWithContext(ctx context.Context, input *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	return da.dynamodbAccessor.PutItemSdkWithContext(ctx, input, optFns...)
}

// QuerySdk implements TransactionalDynamoDBAccessor.
func (da *defaultTransactionalDynamoDBAccessor) QuerySdk(input *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	return da.dynamodbAccessor.QuerySdk(input, optFns...)
}

// QuerySDKWithContext implements TransactionalDynamoDBAccessor.
func (da *defaultTransactionalDynamoDBAccessor) QuerySDKWithContext(ctx context.Context, input *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	return da.dynamodbAccessor.QuerySDKWithContext(ctx, input, optFns...)
}

// QueryPagesSdk implements TransactionalDynamoDBAccessor.
func (da *defaultTransactionalDynamoDBAccessor) QueryPagesSdk(input *dynamodb.QueryInput, fn func(*dynamodb.QueryOutput) bool, optFns ...func(*dynamodb.Options)) error {
	return da.dynamodbAccessor.QueryPagesSdk(input, fn, optFns...)
}

// QueryPagesSdkWithContext implements TransactionalDynamoDBAccessor.
func (da *defaultTransactionalDynamoDBAccessor) QueryPagesSdkWithContext(ctx context.Context, input *dynamodb.QueryInput, fn func(*dynamodb.QueryOutput) bool, optFns ...func(*dynamodb.Options)) error {
	return da.dynamodbAccessor.QueryPagesSdkWithContext(ctx, input, fn, optFns...)
}

// UpdateItemSdk implements TransactionalDynamoDBAccessor.
func (da *defaultTransactionalDynamoDBAccessor) UpdateItemSdk(input *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
	return da.dynamodbAccessor.UpdateItemSdk(input, optFns...)
}

// UpdateItemSdkWithContext implements TransactionalDynamoDBAccessor.
func (da *defaultTransactionalDynamoDBAccessor) UpdateItemSdkWithContext(ctx context.Context, input *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
	return da.dynamodbAccessor.UpdateItemSdkWithContext(ctx, input, optFns...)
}

// DeleteItemSdk implements TransactionalDynamoDBAccessor.
func (da *defaultTransactionalDynamoDBAccessor) DeleteItemSdk(input *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
	return da.dynamodbAccessor.DeleteItemSdk(input, optFns...)
}

// DeleteItemSdkWithContext implements TransactionalDynamoDBAccessor.
func (da *defaultTransactionalDynamoDBAccessor) DeleteItemSdkWithContext(ctx context.Context, input *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
	return da.dynamodbAccessor.DeleteItemSdkWithContext(ctx, input, optFns...)
}

// BatchGetItemSdk implements TransactionalDynamoDBAccessor.
func (da *defaultTransactionalDynamoDBAccessor) BatchGetItemSdk(input *dynamodb.BatchGetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchGetItemOutput, error) {
	return da.dynamodbAccessor.BatchGetItemSdk(input, optFns...)
}

// BatchGetItemSdkWithContext implements TransactionalDynamoDBAccessor.
func (da *defaultTransactionalDynamoDBAccessor) BatchGetItemSdkWithContext(ctx context.Context, input *dynamodb.BatchGetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchGetItemOutput, error) {
	return da.dynamodbAccessor.BatchGetItemSdkWithContext(ctx, input, optFns...)
}

// BatchWriteItemSdk implements TransactionalDynamoDBAccessor.
func (da *defaultTransactionalDynamoDBAccessor) BatchWriteItemSdk(input *dynamodb.BatchWriteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchWriteItemOutput, error) {
	return da.dynamodbAccessor.BatchWriteItemSdk(input, optFns...)
}

// BatchWriteItemSdkWithContext implements TransactionalDynamoDBAccessor.
func (da *defaultTransactionalDynamoDBAccessor) BatchWriteItemSdkWithContext(ctx context.Context, input *dynamodb.BatchWriteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchWriteItemOutput, error) {
	return da.dynamodbAccessor.BatchWriteItemSdkWithContext(ctx, input, optFns...)
}

// AppendTransactWriteItem implements TransactionalDynamoDBAccessor.
func (da *defaultTransactionalDynamoDBAccessor) AppendTransactWriteItem(item *types.TransactWriteItem) error {
	da.logger.Debug("AppendTransactWriteItem")
	return da.AppendTransactWriteItemWithContext(apcontext.Context, item)
}

// AppendTransactWriteItemWithContext implements TransactionalDynamoDBAccessor.
func (da *defaultTransactionalDynamoDBAccessor) AppendTransactWriteItemWithContext(ctx context.Context, item *types.TransactWriteItem) error {
	da.logger.Debug("AppendTransactWriteItemWithContext")
	if ctx == nil {
		ctx = apcontext.Context
	}
	value := ctx.Value(TRANSACTION_CTX_KEY)
	if value == nil {
		return errors.New("トランザクションが開始されていません")
	}
	transaction, ok := value.(Transaction)
	if !ok {
		return errors.New("トランザクションが開始されていません")
	}
	transaction.AppendTransactWriteItem(item)
	return nil
}

// TransactWriteItemsSDK implements TransactionalDynamoDBAccessor.
func (da *defaultTransactionalDynamoDBAccessor) TransactWriteItemsSDK(items []types.TransactWriteItem, optFns ...func(*dynamodb.Options)) (*dynamodb.TransactWriteItemsOutput, error) {
	return da.TransactWriteItemsSDKWithContext(apcontext.Context, items, optFns...)
}

// TransactWriteItemsSDKWithContext implements TransactionalDynamoDBAccessor.
func (da *defaultTransactionalDynamoDBAccessor) TransactWriteItemsSDKWithContext(ctx context.Context, items []types.TransactWriteItem, optFns ...func(*dynamodb.Options)) (*dynamodb.TransactWriteItemsOutput, error) {
	da.logger.Debug("TransactWriteItemsSDK: %d件", len(items))
	if ctx == nil {
		ctx = apcontext.Context
	}
	input := &dynamodb.TransactWriteItemsInput{TransactItems: items}
	// ReturnConsumedCapacityを設定
	if myDynamoDB.ReturnConsumedCapacity(da.config) {
		input.ReturnConsumedCapacity = types.ReturnConsumedCapacityTotal

	}
	output, err := da.GetDynamoDBClient().TransactWriteItems(ctx, input, optFns...)

	if err != nil {
		return nil, errors.WithStack(err)
	}
	if output.ConsumedCapacity != nil && len(output.ConsumedCapacity) > 0 {
		da.logger.Debug("消費キャパシティユニット: %d件", len(output.ConsumedCapacity))
		for i, v := range output.ConsumedCapacity {
			da.logger.Debug("TransactWriteItems(%d番目)[%s]消費キャパシティユニット:%f", i+1, *v.TableName, *v.CapacityUnits)
		}
	}
	return output, nil
}
