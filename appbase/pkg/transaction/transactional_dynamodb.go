/*
transaction パッケージは、トランザクション管理に関する機能を提供するパッケージです。
*/
package transaction

import (
	"context"

	"example.com/appbase/pkg/apcontext"
	myConfig "example.com/appbase/pkg/config"
	myDynamoDB "example.com/appbase/pkg/dynamodb"
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
}

// NewTransactionalDynamoDBAccessor は、TransactionalDynamoDBAccessorを作成します。
func NewTransactionalDynamoDBAccessor(logger logging.Logger, myCfg myConfig.Config) (TransactionalDynamoDBAccessor, error) {
	dynamodbAccessor, err := myDynamoDB.NewDynamoDBAccessor(logger, myCfg)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &defaultTransactionalDynamoDBAccessor{logger: logger, config: myCfg, dynamodbAccessor: dynamodbAccessor}, nil
}

type defaultTransactionalDynamoDBAccessor struct {
	logger           logging.Logger
	config           myConfig.Config
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

// PutItemSdk implements TransactionalDynamoDBAccessor.
func (da *defaultTransactionalDynamoDBAccessor) PutItemSdk(input *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	return da.dynamodbAccessor.PutItemSdk(input, optFns...)
}

// QueryPagesSdk implements TransactionalDynamoDBAccessor.
func (da *defaultTransactionalDynamoDBAccessor) QueryPagesSdk(input *dynamodb.QueryInput, fn func(*dynamodb.QueryOutput) bool, optFns ...func(*dynamodb.Options)) error {
	return da.dynamodbAccessor.QueryPagesSdk(input, fn, optFns...)
}

// QuerySdk implements TransactionalDynamoDBAccessor.
func (da *defaultTransactionalDynamoDBAccessor) QuerySdk(input *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	return da.dynamodbAccessor.QuerySdk(input, optFns...)
}

// UpdateItemSdk implements TransactionalDynamoDBAccessor.
func (da *defaultTransactionalDynamoDBAccessor) UpdateItemSdk(input *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
	return da.dynamodbAccessor.UpdateItemSdk(input, optFns...)
}

// DeleteItemSdk implements TransactionalDynamoDBAccessor.
func (da *defaultTransactionalDynamoDBAccessor) DeleteItemSdk(input *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
	return da.dynamodbAccessor.DeleteItemSdk(input, optFns...)
}

// BatchGetItemSdk implements TransactionalDynamoDBAccessor.
func (da *defaultTransactionalDynamoDBAccessor) BatchGetItemSdk(input *dynamodb.BatchGetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchGetItemOutput, error) {
	return da.dynamodbAccessor.BatchGetItemSdk(input, optFns...)
}

// BatchWriteItemSdk implements TransactionalDynamoDBAccessor.
func (da *defaultTransactionalDynamoDBAccessor) BatchWriteItemSdk(input *dynamodb.BatchWriteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchWriteItemOutput, error) {
	return da.dynamodbAccessor.BatchWriteItemSdk(input, optFns...)
}

// AppendTransactWriteItem implements TransactionalDynamoDBAccessor.
func (da *defaultTransactionalDynamoDBAccessor) AppendTransactWriteItem(item *types.TransactWriteItem) error {
	da.logger.Debug("AppendTransactWriteItem")
	return da.AppendTransactWriteItemWithContext(apcontext.Context, item)
}

// AppendTransactWriteItemWithContext implements TransactionalDynamoDBAccessor.
func (da *defaultTransactionalDynamoDBAccessor) AppendTransactWriteItemWithContext(ctx context.Context, item *types.TransactWriteItem) error {
	da.logger.Debug("AppendTransactWriteItemWithContext")
	value := ctx.Value(TRANSACTION_CTX_KEY)
	if value == nil {
		// TODO: エラー処理
		return errors.New("トランザクションが開始されていません")
	}
	transaction, ok := value.(Transaction)
	if !ok {
		// TODO: エラー処理
		return errors.New("トランザクションが開始されていません")
	}
	transaction.AppendTransactWriteItem(item)
	return nil
}

// TransactWriteItemsSDK implements TransactionalDynamoDBAccessor.
func (da *defaultTransactionalDynamoDBAccessor) TransactWriteItemsSDK(items []types.TransactWriteItem, optFns ...func(*dynamodb.Options)) (*dynamodb.TransactWriteItemsOutput, error) {
	da.logger.Debug("TransactWriteItemsSDK: %d件", len(items))
	input := &dynamodb.TransactWriteItemsInput{TransactItems: items}
	if myDynamoDB.ReturnConsumedCapacity(da.config) {
		input.ReturnConsumedCapacity = types.ReturnConsumedCapacityTotal
	}
	output, err := da.GetDynamoDBClient().TransactWriteItems(apcontext.Context, input, optFns...)

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
