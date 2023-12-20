/*
transaction パッケージは、トランザクション管理に関する機能を提供するパッケージです。
*/
package transaction

import (
	"example.com/appbase/pkg/apcontext"
	myConfig "example.com/appbase/pkg/config"
	myDynamoDB "example.com/appbase/pkg/dynamodb"
	"example.com/appbase/pkg/env"
	"example.com/appbase/pkg/logging"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/cockroachdb/errors"
)

// TransactionalDynamoDBAccessorは、トランザクション管理可能なDynamoDBアクセス用インタフェースです。
type TransactionalDynamoDBAccessor interface {
	myDynamoDB.DynamoDBAccessor
	// StartTransaction は、トランザクションを開始します。
	StartTransaction(transaction Transaction)
	// AppendTransactWriteItem は、トランザクション書き込みしたい場合に対象のTransactWriteItemを追加します。
	// なお、TransactWriteItemsの実行は、TransactionManagerのExecuteTransaction関数で実行されるdomain.ServiceFunc関数が終了する際にtransactionWriteItemsSDKを実施します。
	AppendTransactWriteItem(item *types.TransactWriteItem)
	// TransactWriteItemsSDK は、AWS SDKによるTransactWriteItemsをラップします。
	// なお、TransactWriteItemsの実行は、TransactionManagerが実行するため非公開にしています。
	TransactWriteItemsSDK(items []types.TransactWriteItem) (*dynamodb.TransactWriteItemsOutput, error)
	// EndTransactionは、トランザクションを終了します。
	EndTransaction()
}

// NewTransactionalDynamoDBAccessor は、TransactionalDynamoDBAccessorを作成します。
func NewTransactionalDynamoDBAccessor(log logging.Logger, myCfg myConfig.Config) (TransactionalDynamoDBAccessor, error) {
	dynamodbAccessor, err := myDynamoDB.NewDynamoDBAccessor(log, myCfg)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &defaultTransactionalDynamoDBAccessor{log: log, dynamodbAccessor: dynamodbAccessor}, nil
}

type defaultTransactionalDynamoDBAccessor struct {
	log              logging.Logger
	dynamodbAccessor myDynamoDB.DynamoDBAccessor
	transaction      Transaction
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

// StartTransaction implements TransactionalDynamoDBAccessor.
func (da *defaultTransactionalDynamoDBAccessor) StartTransaction(transaction Transaction) {
	// TODO: 本当はここがスレッド毎にトランザクション管理できるとgoroutineセーフにできるが、現状難しい
	da.transaction = transaction
}

// AppendTransactWriteItem implements TransactionalDynamoDBAccessor.
func (da *defaultTransactionalDynamoDBAccessor) AppendTransactWriteItem(item *types.TransactWriteItem) {
	// TODO: 本当はここがスレッド毎にトランザクション管理できるとgoroutineセーフにできるが、現状難しい
	da.log.Debug("AppendTransactWriteItem")
	da.transaction.AppendTransactWriteItem(item)
}

// TransactWriteItemsSDK implements TransactionalDynamoDBAccessor.
func (da *defaultTransactionalDynamoDBAccessor) TransactWriteItemsSDK(items []types.TransactWriteItem) (*dynamodb.TransactWriteItemsOutput, error) {
	da.log.Debug("TransactWriteItemsSDK")
	input := &dynamodb.TransactWriteItemsInput{TransactItems: items}
	if !env.IsStragingOrProd() {
		input.ReturnConsumedCapacity = types.ReturnConsumedCapacityTotal
	}
	output, err := da.GetDynamoDBClient().TransactWriteItems(apcontext.Context, input)

	// TODO: TransactWriteItems実行時に主キー重複（ErrKeyDuplicaiton）や
	// 更新条件、削除条件のエラー（ErrUpdateWithCondtion、ErrDeleteWithCondtion）相当が
	// あった場合に、業務AP側でどうハンドリングするのか？
	if err != nil {
		return nil, errors.WithStack(err)
	}
	for i, v := range output.ConsumedCapacity {
		da.log.Debug("TransactWriteItems(%d)[%s]消費キャパシティユニット:%f", i, *v.TableName, *v.CapacityUnits)
	}
	return output, nil
}

// EndTransaction implements DynamoDBAccessor.
func (da *defaultTransactionalDynamoDBAccessor) EndTransaction() {
	da.transaction = nil
}
