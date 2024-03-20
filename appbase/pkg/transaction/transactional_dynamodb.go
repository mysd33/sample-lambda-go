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
	TransactWriteItemsSDK(items []types.TransactWriteItem) (*dynamodb.TransactWriteItemsOutput, error)
}

// NewTransactionalDynamoDBAccessor は、TransactionalDynamoDBAccessorを作成します。
func NewTransactionalDynamoDBAccessor(log logging.Logger, myCfg myConfig.Config) (TransactionalDynamoDBAccessor, error) {
	dynamodbAccessor, err := myDynamoDB.NewDynamoDBAccessor(log, myCfg)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &defaultTransactionalDynamoDBAccessor{log: log, config: myCfg, dynamodbAccessor: dynamodbAccessor}, nil
}

type defaultTransactionalDynamoDBAccessor struct {
	log              logging.Logger
	config           myConfig.Config
	dynamodbAccessor myDynamoDB.DynamoDBAccessor
}

// GetDynamoDBClient implements TransactionalDynamoDBAccessor.
func (da *defaultTransactionalDynamoDBAccessor) GetDynamoDBClient() *dynamodb.Client {
	return da.dynamodbAccessor.GetDynamoDBClient()
}

// GetItemSdk implements TransactionalDynamoDBAccessor.
func (da *defaultTransactionalDynamoDBAccessor) GetItemSdk(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
	return da.dynamodbAccessor.GetItemSdk(input)
}

// PutItemSdk implements TransactionalDynamoDBAccessor.
func (da *defaultTransactionalDynamoDBAccessor) PutItemSdk(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	return da.dynamodbAccessor.PutItemSdk(input)
}

// QueryPagesSdk implements TransactionalDynamoDBAccessor.
func (da *defaultTransactionalDynamoDBAccessor) QueryPagesSdk(input *dynamodb.QueryInput, fn func(*dynamodb.QueryOutput) bool) error {
	return da.dynamodbAccessor.QueryPagesSdk(input, fn)
}

// QuerySdk implements TransactionalDynamoDBAccessor.
func (da *defaultTransactionalDynamoDBAccessor) QuerySdk(input *dynamodb.QueryInput) (*dynamodb.QueryOutput, error) {
	return da.dynamodbAccessor.QuerySdk(input)
}

// UpdateItemSdk implements TransactionalDynamoDBAccessor.
func (da *defaultTransactionalDynamoDBAccessor) UpdateItemSdk(input *dynamodb.UpdateItemInput) (*dynamodb.UpdateItemOutput, error) {
	return da.dynamodbAccessor.UpdateItemSdk(input)
}

// DeleteItemSdk implements TransactionalDynamoDBAccessor.
func (da *defaultTransactionalDynamoDBAccessor) DeleteItemSdk(input *dynamodb.DeleteItemInput) (*dynamodb.DeleteItemOutput, error) {
	return da.dynamodbAccessor.DeleteItemSdk(input)
}

// BatchGetItemSdk implements TransactionalDynamoDBAccessor.
func (da *defaultTransactionalDynamoDBAccessor) BatchGetItemSdk(input *dynamodb.BatchGetItemInput) (*dynamodb.BatchGetItemOutput, error) {
	return da.dynamodbAccessor.BatchGetItemSdk(input)
}

// BatchWriteItemSdk implements TransactionalDynamoDBAccessor.
func (da *defaultTransactionalDynamoDBAccessor) BatchWriteItemSdk(input *dynamodb.BatchWriteItemInput) (*dynamodb.BatchWriteItemOutput, error) {
	return da.dynamodbAccessor.BatchWriteItemSdk(input)
}

// AppendTransactWriteItem implements TransactionalDynamoDBAccessor.
func (da *defaultTransactionalDynamoDBAccessor) AppendTransactWriteItem(item *types.TransactWriteItem) error {
	da.log.Debug("AppendTransactWriteItem")
	return da.AppendTransactWriteItemWithContext(apcontext.Context, item)
}

// AppendTransactWriteItemWithContext implements TransactionalDynamoDBAccessor.
func (da *defaultTransactionalDynamoDBAccessor) AppendTransactWriteItemWithContext(ctx context.Context, item *types.TransactWriteItem) error {
	da.log.Debug("AppendTransactWriteItemWithContext")
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
func (da *defaultTransactionalDynamoDBAccessor) TransactWriteItemsSDK(items []types.TransactWriteItem) (*dynamodb.TransactWriteItemsOutput, error) {
	da.log.Debug("TransactWriteItemsSDK: %d件", len(items))
	input := &dynamodb.TransactWriteItemsInput{TransactItems: items}
	if myDynamoDB.ReturnConsumedCapacity(da.config) {
		input.ReturnConsumedCapacity = types.ReturnConsumedCapacityTotal
	}
	output, err := da.GetDynamoDBClient().TransactWriteItems(apcontext.Context, input)

	if err != nil {
		return nil, errors.WithStack(err)
	}
	if output.ConsumedCapacity != nil && len(output.ConsumedCapacity) > 0 {
		da.log.Debug("消費キャパシティユニット: %d件", len(output.ConsumedCapacity))
		for i, v := range output.ConsumedCapacity {
			da.log.Debug("TransactWriteItems(%d番目)[%s]消費キャパシティユニット:%f", i+1, *v.TableName, *v.CapacityUnits)
		}
	}
	return output, nil
}
