/*
dynamodb パッケージは、DynamoDBアクセスに関する機能を提供するパッケージです。
*/
package dynamodb

import (
	"context"

	"example.com/appbase/pkg/apcontext"
	myConfig "example.com/appbase/pkg/config"
	"example.com/appbase/pkg/constant"
	"example.com/appbase/pkg/logging"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-xray-sdk-go/instrumentation/awsv2"
	"github.com/cockroachdb/errors"
)

// createDynamoDBClient は、DynamoDBClientを作成します。
func createDynamoDBClient(myCfg myConfig.Config) (*dynamodb.Client, error) {
	//TODO: カスタムHTTPClientの作成

	// AWS SDK for Go v2 Migration
	// https://github.com/aws/aws-sdk-go-v2
	// https://aws.github.io/aws-sdk-go-v2/docs/migrating/
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, errors.WithStack(err)
	}
	// Instrumenting AWS SDK v2
	// https://github.com/aws/aws-xray-sdk-go
	awsv2.AWSV2Instrumentor(&cfg.APIOptions)
	return dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		// DynamoDB Local起動先が指定されている場合
		dynamodbEndpoint := myCfg.Get(constant.DYNAMODB_LOCAL_ENDPOINT_NAME)
		if dynamodbEndpoint != "" {
			o.BaseEndpoint = aws.String(dynamodbEndpoint)
		}
	}), nil
}

// DynamoDBAccessor は、AWS SDKを使ったDynamoDBアクセスの実装をラップしカプセル化するインタフェースです。
type DynamoDBAccessor interface {
	// GetItemSdk は、AWS SDKによるGetItemをラップします。
	GetItemSdk(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error)
	// QuerySdk は、AWS SDKによるQueryをラップします。
	QuerySdk(input *dynamodb.QueryInput) (*dynamodb.QueryOutput, error)
	// QueryPagesSdk は、AWS SDKによるQueryのページング処理をラップします。
	QueryPagesSdk(input *dynamodb.QueryInput, fn func(*dynamodb.QueryOutput) bool) error
	// PutItemSdk は、AWS SDKによるPutItemをラップします。
	PutItemSdk(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error)
	// UpdateItemSdk は、AWS SDKによるUpdateItemをラップします。
	UpdateItemSdk(input *dynamodb.UpdateItemInput) (*dynamodb.UpdateItemOutput, error)
	// DeleteItemSdk は、AWS SDKによるDeleteItemをラップします。
	DeleteItemSdk(input *dynamodb.DeleteItemInput) (*dynamodb.DeleteItemOutput, error)
	// BatchGetItemSdk は、AWS SDKによるBatchGetItemをラップします。
	BatchGetItemSdk(input *dynamodb.BatchGetItemInput) (*dynamodb.BatchGetItemOutput, error)
	// BatchWriteItemSdk は、AWS SDKによるBatchWriteItemをラップします。
	BatchWriteItemSdk(input *dynamodb.BatchWriteItemInput) (*dynamodb.BatchWriteItemOutput, error)
	// startTransactionは、トランザクションを開始します。
	startTransaction(transaction transaction)
	// AppendTransactWriteItemは、トランザクション書き込みしたい場合に対象のTransactWriteItemを追加します。
	// なお、TransactWriteItemsの実行は、TransactionManagerのExecuteTransaction関数で実行されるdomain.ServiceFunc関数が終了する際にtransactionWriteItemsSDKを実施します。
	AppendTransactWriteItem(item *types.TransactWriteItem)
	// transactWriteItemsSDK は、AWS SDKによるTransactWriteItemsをラップします。
	// なお、TransactWriteItemsの実行は、TransactionManagerが実行するため非公開にしています。
	transactWriteItemsSDK(items []types.TransactWriteItem) (*dynamodb.TransactWriteItemsOutput, error)
	// endTransactionは、トランザクションを終了します。
	endTransaction()
}

// NewDynamoDBAccessor は、Acccessorを作成します。
func NewDynamoDBAccessor(log logging.Logger, myCfg myConfig.Config) (DynamoDBAccessor, error) {
	dynamodbClient, err := createDynamoDBClient(myCfg)
	if err != nil {
		return nil, err
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
	return d.dynamodbClient.GetItem(apcontext.Context, input)
}

// QuerySdk implements DynamoDBAccessor.
func (d *defaultDynamoDBAccessor) QuerySdk(input *dynamodb.QueryInput) (*dynamodb.QueryOutput, error) {
	return d.dynamodbClient.Query(apcontext.Context, input)
}

// QueryPagesSdk implements DynamoDBAccessor.
func (d *defaultDynamoDBAccessor) QueryPagesSdk(input *dynamodb.QueryInput, fn func(*dynamodb.QueryOutput) bool) error {
	paginator := dynamodb.NewQueryPaginator(d.dynamodbClient, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(apcontext.Context)
		if err != nil {
			return errors.WithStack(err)
		}
		if fn(page) {
			break
		}
	}
	return nil
}

// PutItemSdk implements DynamoDBAccessor.
func (d *defaultDynamoDBAccessor) PutItemSdk(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	return d.dynamodbClient.PutItem(apcontext.Context, input)
}

// UpdateItemSdk implements DynamoDBAccessor.
func (d *defaultDynamoDBAccessor) UpdateItemSdk(input *dynamodb.UpdateItemInput) (*dynamodb.UpdateItemOutput, error) {
	return d.dynamodbClient.UpdateItem(apcontext.Context, input)
}

// DeleteItemSdk implements DynamoDBAccessor.
func (d *defaultDynamoDBAccessor) DeleteItemSdk(input *dynamodb.DeleteItemInput) (*dynamodb.DeleteItemOutput, error) {
	return d.dynamodbClient.DeleteItem(apcontext.Context, input)
}

// BatchGetItemSdk implements DynamoDBAccessor.
func (d *defaultDynamoDBAccessor) BatchGetItemSdk(input *dynamodb.BatchGetItemInput) (*dynamodb.BatchGetItemOutput, error) {
	return d.dynamodbClient.BatchGetItem(apcontext.Context, input)
}

// BatchWriteItemSdk implements DynamoDBAccessor.
func (d *defaultDynamoDBAccessor) BatchWriteItemSdk(input *dynamodb.BatchWriteItemInput) (*dynamodb.BatchWriteItemOutput, error) {
	return d.dynamodbClient.BatchWriteItem(apcontext.Context, input)
}

// startTransaction implements DynamoDBAccessor.
func (d *defaultDynamoDBAccessor) startTransaction(transaction transaction) {
	// TODO: 本当はここがスレッド毎にトランザクション管理できるとgoroutineセーフにできるが、現状難しい
	d.transaction = transaction
}

// AppendTransactWriteItem implements DynamoDBAccessor.
func (d *defaultDynamoDBAccessor) AppendTransactWriteItem(item *types.TransactWriteItem) {
	// TODO: 本当はここがスレッド毎にトランザクション管理できるとgoroutineセーフにできるが、現状難しい
	d.transaction.appendTransactWriteItem(item)
}

// transactWriteItemsSDK implements DynamoDBAccessor.
func (d *defaultDynamoDBAccessor) transactWriteItemsSDK(items []types.TransactWriteItem) (*dynamodb.TransactWriteItemsOutput, error) {
	return d.dynamodbClient.TransactWriteItems(apcontext.Context, &dynamodb.TransactWriteItemsInput{
		TransactItems: items,
	})
}

// endTransaction implements DynamoDBAccessor.
func (d *defaultDynamoDBAccessor) endTransaction() {
	d.transaction = nil
}
