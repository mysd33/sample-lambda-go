/*
dynamodb パッケージは、DynamoDBアクセスに関する機能を提供するパッケージです。
*/
package dynamodb

import (
	"context"
	"os"

	"example.com/appbase/pkg/apcontext"
	"example.com/appbase/pkg/constant"
	"example.com/appbase/pkg/domain"
	"example.com/appbase/pkg/logging"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-xray-sdk-go/instrumentation/awsv2"
	"github.com/cockroachdb/errors"
)

var (
	// DynamoDBクライアント
	dynamodbClient *dynamodb.Client
	// 書き込みトランザクション
	transactWriteItems []types.TransactWriteItem
	// TODO: 読み込みトランザクションTransactGetItems
	// transactGetItems []types.TransactGetItem
)

// TODO:トランザクション関連のデバッグログの出力

// ExecuteTransaction は、Serviceの関数serviceFuncの実行前後でDynamoDBトランザクション実行します。
func ExecuteTransaction(serviceFunc domain.ServiceFunc) (interface{}, error) {
	var err error
	// DynamoDBClientの作成
	dynamodbClient, err = createDynamoDBClient()
	if err != nil {
		return nil, err
	}
	// サービスの実行
	result, err := serviceFunc()
	// DynamoDBのトランザクションを終了
	_, err = endTransaction(err)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// createDynamoDBClient は、DynamoDBClientを作成します。
func createDynamoDBClient() (*dynamodb.Client, error) {
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
		dynamodbEndpoint := os.Getenv(constant.DYNAMODB_LOCAL_ENDPOINT_NAME)
		if dynamodbEndpoint != "" {
			o.BaseEndpoint = aws.String(dynamodbEndpoint)
		}
	}), nil
}

// checkTransactWriteItems は、TransactWriteItemが存在するかを確認します。
func checkTransactWriteItems() bool {
	return len(transactWriteItems) > 0
}

// clearTransactWriteItems() は、TransactWriteItemをクリアします。
func clearTransactWriteItems() {
	transactWriteItems = nil
}

// endTransaction は、エラーがなければ、AWS SDKによるTransactionWriteItemsを実行しトランザクション実行し、エラーがある場合には実行しません。
// TODO: TransactGetItemsの考慮
func endTransaction(err error) (*dynamodb.TransactWriteItemsOutput, error) {
	if !checkTransactWriteItems() {
		return nil, nil
	}
	// 処理結果がどんな場合でもTransactWriteItemをクリア
	defer clearTransactWriteItems()
	if err != nil {
		// Serviceの処理結果がエラー場合は、トランザクションを実行せず、元のエラーを返却し終了
		return nil, err
	}
	// トランザクション実行
	output, err := dynamodbClient.TransactWriteItems(apcontext.Context, &dynamodb.TransactWriteItemsInput{
		TransactItems: transactWriteItems,
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return output, nil
}

// NewDynamoDBAccessor は、Acccessorを作成します。
func NewDynamoDBAccessor(log logging.Logger) (DynamoDBAccessor, error) {
	return &defaultDynamoDBAccessor{log: log}, nil
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
	// AppendTransactWriteItemは、トランザクション書き込みしたい場合に対象のTransactWriteItemを追加します。
	// なお、TransactWriteItemsの実行は、ExecuteTransaction関数で実行されるdomain.ServiceFunc関数が終了する際に実施します。
	AppendTransactWriteItem(item types.TransactWriteItem)
}

type defaultDynamoDBAccessor struct {
	log logging.Logger
}

// GetItemSdk implements DynamoDBAccessor.
func (d *defaultDynamoDBAccessor) GetItemSdk(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
	return dynamodbClient.GetItem(apcontext.Context, input)
}

// QuerySdk implements DynamoDBAccessor.
func (d *defaultDynamoDBAccessor) QuerySdk(input *dynamodb.QueryInput) (*dynamodb.QueryOutput, error) {
	return dynamodbClient.Query(apcontext.Context, input)
}

// QueryPagesSdk implements DynamoDBAccessor.
func (d *defaultDynamoDBAccessor) QueryPagesSdk(input *dynamodb.QueryInput, fn func(*dynamodb.QueryOutput) bool) error {
	paginator := dynamodb.NewQueryPaginator(dynamodbClient, input)
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
	return dynamodbClient.PutItem(apcontext.Context, input)
}

// UpdateItemSdk implements DynamoDBAccessor.
func (d *defaultDynamoDBAccessor) UpdateItemSdk(input *dynamodb.UpdateItemInput) (*dynamodb.UpdateItemOutput, error) {
	return dynamodbClient.UpdateItem(apcontext.Context, input)
}

// DeleteItemSdk implements DynamoDBAccessor.
func (d *defaultDynamoDBAccessor) DeleteItemSdk(input *dynamodb.DeleteItemInput) (*dynamodb.DeleteItemOutput, error) {
	return dynamodbClient.DeleteItem(apcontext.Context, input)
}

// BatchGetItemSdk implements DynamoDBAccessor.
func (d *defaultDynamoDBAccessor) BatchGetItemSdk(input *dynamodb.BatchGetItemInput) (*dynamodb.BatchGetItemOutput, error) {
	return dynamodbClient.BatchGetItem(apcontext.Context, input)
}

// BatchWriteItemSdk implements DynamoDBAccessor.
func (d *defaultDynamoDBAccessor) BatchWriteItemSdk(input *dynamodb.BatchWriteItemInput) (*dynamodb.BatchWriteItemOutput, error) {
	return dynamodbClient.BatchWriteItem(apcontext.Context, input)
}

// AppendTransactWriteItem implements DynamoDBAccessor.
func (d *defaultDynamoDBAccessor) AppendTransactWriteItem(item types.TransactWriteItem) {
	transactWriteItems = append(transactWriteItems, item)
}
