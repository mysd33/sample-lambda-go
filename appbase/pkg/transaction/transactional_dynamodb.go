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
	dynamodbClient, err := myDynamoDB.CreateDynamoDBClient(myCfg)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &defaultTransactionalDynamoDBAccessor{log: log, dynamodbClient: dynamodbClient}, nil
}

type defaultTransactionalDynamoDBAccessor struct {
	log            logging.Logger
	dynamodbClient *dynamodb.Client
	transaction    Transaction
}

// GetItemSdk implements DynamoDBAccessor.
func (d *defaultTransactionalDynamoDBAccessor) GetItemSdk(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
	if !env.IsStragingOrProd() {
		input.ReturnConsumedCapacity = types.ReturnConsumedCapacityTotal
	}
	output, err := d.dynamodbClient.GetItem(apcontext.Context, input)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if output.ConsumedCapacity != nil {
		d.log.Debug("GetItem[%s]消費キャパシティユニット:%f", *output.ConsumedCapacity.TableName, *output.ConsumedCapacity.CapacityUnits)
	}
	return output, nil
}

// QuerySdk implements DynamoDBAccessor.
func (d *defaultTransactionalDynamoDBAccessor) QuerySdk(input *dynamodb.QueryInput) (*dynamodb.QueryOutput, error) {
	if !env.IsStragingOrProd() {
		input.ReturnConsumedCapacity = types.ReturnConsumedCapacityTotal
	}
	output, err := d.dynamodbClient.Query(apcontext.Context, input)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if output.ConsumedCapacity != nil {
		d.log.Debug("Query[%s]消費キャパシティユニット:%f", *output.ConsumedCapacity.TableName, *output.ConsumedCapacity.CapacityUnits)
	}
	return output, nil
}

// QueryPagesSdk implements DynamoDBAccessor.
func (d *defaultTransactionalDynamoDBAccessor) QueryPagesSdk(input *dynamodb.QueryInput, fn func(*dynamodb.QueryOutput) bool) error {
	paginator := dynamodb.NewQueryPaginator(d.dynamodbClient, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(apcontext.Context)
		if err != nil {
			return errors.WithStack(err)
		}
		if page.ConsumedCapacity != nil {
			d.log.Debug("Query[%s]消費キャパシティユニット:%f", *page.ConsumedCapacity.TableName, *page.ConsumedCapacity.CapacityUnits)
		}
		if fn(page) {
			break
		}
	}
	return nil
}

// PutItemSdk implements DynamoDBAccessor.
func (d *defaultTransactionalDynamoDBAccessor) PutItemSdk(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	if !env.IsStragingOrProd() {
		input.ReturnConsumedCapacity = types.ReturnConsumedCapacityTotal
	}
	output, err := d.dynamodbClient.PutItem(apcontext.Context, input)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if output.ConsumedCapacity != nil {
		d.log.Debug("PutItem[%s]消費キャパシティユニット:%f", *output.ConsumedCapacity.TableName, *output.ConsumedCapacity.CapacityUnits)
	}
	return output, nil

}

// UpdateItemSdk implements DynamoDBAccessor.
func (d *defaultTransactionalDynamoDBAccessor) UpdateItemSdk(input *dynamodb.UpdateItemInput) (*dynamodb.UpdateItemOutput, error) {
	if !env.IsStragingOrProd() {
		input.ReturnConsumedCapacity = types.ReturnConsumedCapacityTotal
	}
	output, err := d.dynamodbClient.UpdateItem(apcontext.Context, input)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if output.ConsumedCapacity != nil {
		d.log.Debug("UpdateItem[%s]消費キャパシティユニット:%f", *output.ConsumedCapacity.TableName, *output.ConsumedCapacity.CapacityUnits)
	}
	return output, nil
}

// DeleteItemSdk implements DynamoDBAccessor.
func (d *defaultTransactionalDynamoDBAccessor) DeleteItemSdk(input *dynamodb.DeleteItemInput) (*dynamodb.DeleteItemOutput, error) {
	if !env.IsStragingOrProd() {
		input.ReturnConsumedCapacity = types.ReturnConsumedCapacityTotal
	}
	output, err := d.dynamodbClient.DeleteItem(apcontext.Context, input)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if output.ConsumedCapacity != nil {
		d.log.Debug("DeleteItem[%s]消費キャパシティユニット:%f", *output.ConsumedCapacity.TableName, *output.ConsumedCapacity.CapacityUnits)
	}
	return output, nil
}

// BatchGetItemSdk implements DynamoDBAccessor.
func (d *defaultTransactionalDynamoDBAccessor) BatchGetItemSdk(input *dynamodb.BatchGetItemInput) (*dynamodb.BatchGetItemOutput, error) {
	if !env.IsStragingOrProd() {
		input.ReturnConsumedCapacity = types.ReturnConsumedCapacityTotal
	}
	output, err := d.dynamodbClient.BatchGetItem(apcontext.Context, input)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	for i, v := range output.ConsumedCapacity {
		d.log.Debug("BatchGetItem(%d)[%s]消費キャパシティユニット:%f", i, *v.TableName, *v.CapacityUnits)
	}
	return output, nil
}

// BatchWriteItemSdk implements DynamoDBAccessor.
func (d *defaultTransactionalDynamoDBAccessor) BatchWriteItemSdk(input *dynamodb.BatchWriteItemInput) (*dynamodb.BatchWriteItemOutput, error) {
	if !env.IsStragingOrProd() {
		input.ReturnConsumedCapacity = types.ReturnConsumedCapacityTotal
	}
	output, err := d.dynamodbClient.BatchWriteItem(apcontext.Context, input)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	for i, v := range output.ConsumedCapacity {
		d.log.Debug("BatchWriteItem(%d)[%s]消費キャパシティユニット:%f", i, *v.TableName, *v.CapacityUnits)
	}
	return output, nil
}

// StartTransaction implements DynamoDBAccessor.
func (d *defaultTransactionalDynamoDBAccessor) StartTransaction(transaction Transaction) {
	// TODO: 本当はここがスレッド毎にトランザクション管理できるとgoroutineセーフにできるが、現状難しい
	d.transaction = transaction
}

// AppendTransactWriteItem implements DynamoDBAccessor.
func (d *defaultTransactionalDynamoDBAccessor) AppendTransactWriteItem(item *types.TransactWriteItem) {
	// TODO: 本当はここがスレッド毎にトランザクション管理できるとgoroutineセーフにできるが、現状難しい
	d.log.Debug("AppendTransactWriteItem")
	d.transaction.AppendTransactWriteItem(item)
}

// TransactWriteItemsSDK implements DynamoDBAccessor.
func (d *defaultTransactionalDynamoDBAccessor) TransactWriteItemsSDK(items []types.TransactWriteItem) (*dynamodb.TransactWriteItemsOutput, error) {
	input := &dynamodb.TransactWriteItemsInput{TransactItems: items}
	if !env.IsStragingOrProd() {
		input.ReturnConsumedCapacity = types.ReturnConsumedCapacityTotal
	}
	output, err := d.dynamodbClient.TransactWriteItems(apcontext.Context, input)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	for i, v := range output.ConsumedCapacity {
		d.log.Debug("TransactWriteItems(%d)[%s]消費キャパシティユニット:%f", i, *v.TableName, *v.CapacityUnits)
	}
	return output, nil
}

// EndTransaction implements DynamoDBAccessor.
func (d *defaultTransactionalDynamoDBAccessor) EndTransaction() {
	d.transaction = nil
}
