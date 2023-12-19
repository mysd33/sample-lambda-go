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

// TODO: Mock化しやすいよう全て公開メソッド化する

// TransactionalDynamoDBAccessorは、トランザクション管理可能なDynamoDBアクセス用インタフェースです。
type TransactionalDynamoDBAccessor interface {
	myDynamoDB.DynamoDBAccessor
	// startTransaction は、トランザクションを開始します。
	startTransaction(transaction transaction)
	// AppendTransactWriteItem は、トランザクション書き込みしたい場合に対象のTransactWriteItemを追加します。
	// なお、TransactWriteItemsの実行は、TransactionManagerのExecuteTransaction関数で実行されるdomain.ServiceFunc関数が終了する際にtransactionWriteItemsSDKを実施します。
	AppendTransactWriteItem(item *types.TransactWriteItem)
	// transactWriteItemsSDK は、AWS SDKによるTransactWriteItemsをラップします。
	// なお、TransactWriteItemsの実行は、TransactionManagerが実行するため非公開にしています。
	transactWriteItemsSDK(items []types.TransactWriteItem) (*dynamodb.TransactWriteItemsOutput, error)
	// endTransactionは、トランザクションを終了します。
	endTransaction()
}

// NewDynamoDBAccessor は、TransactionalDynamoDBAccessorを作成します。
func NewDynamoDBAccessor(log logging.Logger, myCfg myConfig.Config) (TransactionalDynamoDBAccessor, error) {
	dynamodbClient, err := myDynamoDB.CreateDynamoDBClient(myCfg)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &defaultDynamoDBAccessor{log: log, dynamodbClient: dynamodbClient}, nil
}

type defaultDynamoDBAccessor struct {
	log            logging.Logger
	dynamodbClient *dynamodb.Client
	transaction    transaction
}

// GetItemSdk implements DynamoDBAccessor.
func (d *defaultDynamoDBAccessor) GetItemSdk(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
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
func (d *defaultDynamoDBAccessor) QuerySdk(input *dynamodb.QueryInput) (*dynamodb.QueryOutput, error) {
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
func (d *defaultDynamoDBAccessor) QueryPagesSdk(input *dynamodb.QueryInput, fn func(*dynamodb.QueryOutput) bool) error {
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
func (d *defaultDynamoDBAccessor) PutItemSdk(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
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
func (d *defaultDynamoDBAccessor) UpdateItemSdk(input *dynamodb.UpdateItemInput) (*dynamodb.UpdateItemOutput, error) {
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
func (d *defaultDynamoDBAccessor) DeleteItemSdk(input *dynamodb.DeleteItemInput) (*dynamodb.DeleteItemOutput, error) {
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
func (d *defaultDynamoDBAccessor) BatchGetItemSdk(input *dynamodb.BatchGetItemInput) (*dynamodb.BatchGetItemOutput, error) {
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
func (d *defaultDynamoDBAccessor) BatchWriteItemSdk(input *dynamodb.BatchWriteItemInput) (*dynamodb.BatchWriteItemOutput, error) {
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

// startTransaction implements DynamoDBAccessor.
func (d *defaultDynamoDBAccessor) startTransaction(transaction transaction) {
	// TODO: 本当はここがスレッド毎にトランザクション管理できるとgoroutineセーフにできるが、現状難しい
	d.transaction = transaction
}

// AppendTransactWriteItem implements DynamoDBAccessor.
func (d *defaultDynamoDBAccessor) AppendTransactWriteItem(item *types.TransactWriteItem) {
	// TODO: 本当はここがスレッド毎にトランザクション管理できるとgoroutineセーフにできるが、現状難しい
	d.log.Debug("AppendTransactWriteItem")
	d.transaction.appendTransactWriteItem(item)
}

// transactWriteItemsSDK implements DynamoDBAccessor.
func (d *defaultDynamoDBAccessor) transactWriteItemsSDK(items []types.TransactWriteItem) (*dynamodb.TransactWriteItemsOutput, error) {
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

// endTransaction implements DynamoDBAccessor.
func (d *defaultDynamoDBAccessor) endTransaction() {
	d.transaction = nil
}
