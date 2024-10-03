/*
dynamodb パッケージは、DynamoDBアクセスに関する機能を提供するパッケージです。
*/
package dynamodb

import (
	"context"

	"example.com/appbase/pkg/apcontext"
	"example.com/appbase/pkg/awssdk"
	myConfig "example.com/appbase/pkg/config"
	"example.com/appbase/pkg/logging"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-xray-sdk-go/instrumentation/awsv2"
	"github.com/cockroachdb/errors"
)

const (
	DYNAMODB_LOCAL_ENDPOINT_NAME = "DYNAMODB_LOCAL_ENDPOINT"
)

// CreateDynamoDBClient は、DynamoDBClientを作成します。
func CreateDynamoDBClient(myCfg myConfig.Config) (*dynamodb.Client, error) {
	// カスタムHTTPClientの作成
	sdkHTTPClient := awssdk.NewHTTPClient(myCfg)
	// ClientLogModeの取得
	clientLogMode, found := awssdk.GetClientLogMode(myCfg)
	// AWS SDK for Go v2 Migration
	// https://github.com/aws/aws-sdk-go-v2
	// https://aws.github.io/aws-sdk-go-v2/docs/migrating/
	var cfg aws.Config
	var err error
	if found {
		cfg, err = config.LoadDefaultConfig(context.TODO(), config.WithHTTPClient(sdkHTTPClient), config.WithClientLogMode(clientLogMode))
	} else {
		cfg, err = config.LoadDefaultConfig(context.TODO(), config.WithHTTPClient(sdkHTTPClient))
	}
	if err != nil {
		return nil, errors.WithStack(err)
	}
	// Instrumenting AWS SDK v2
	// https://github.com/aws/aws-xray-sdk-go
	awsv2.AWSV2Instrumentor(&cfg.APIOptions)
	return dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		// ローカル実行のためDynamoDB Local起動先が指定されている場合
		dynamodbEndpoint := myCfg.Get(DYNAMODB_LOCAL_ENDPOINT_NAME, "")
		if dynamodbEndpoint != "" {
			o.BaseEndpoint = aws.String(dynamodbEndpoint)
		}
	}), nil
}

// DynamoDBAccessor は、AWS SDKを使ったDynamoDBアクセスの実装をラップしカプセル化するインタフェースです。
type DynamoDBAccessor interface {
	// GetDynamoDBClient は、DyanmoDBClientを返却します。
	GetDynamoDBClient() *dynamodb.Client
	// GetItemSdk は、AWS SDKによるGetItemをラップします。
	GetItemSdk(input *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	// QuerySdk は、AWS SDKによるQueryをラップします。
	QuerySdk(input *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
	// QueryPagesSdk は、AWS SDKによるQueryのページング処理をラップします。
	QueryPagesSdk(input *dynamodb.QueryInput, fn func(*dynamodb.QueryOutput) bool, optFns ...func(*dynamodb.Options)) error
	// PutItemSdk は、AWS SDKによるPutItemをラップします。
	PutItemSdk(input *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	// UpdateItemSdk は、AWS SDKによるUpdateItemをラップします。
	UpdateItemSdk(input *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
	// DeleteItemSdk は、AWS SDKによるDeleteItemをラップします。
	DeleteItemSdk(input *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error)
	// BatchGetItemSdk は、AWS SDKによるBatchGetItemをラップします。
	BatchGetItemSdk(input *dynamodb.BatchGetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchGetItemOutput, error)
	// BatchWriteItemSdk は、AWS SDKによるBatchWriteItemをラップします。
	BatchWriteItemSdk(input *dynamodb.BatchWriteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchWriteItemOutput, error)
}

// NewDynamoDBAccessor は、DynamoDBAccessor を作成します。
func NewDynamoDBAccessor(logger logging.Logger, myCfg myConfig.Config) (DynamoDBAccessor, error) {
	dynamodbClient, err := CreateDynamoDBClient(myCfg)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &defaultDynamoDBAccessor{logger: logger, config: myCfg, dynamodbClient: dynamodbClient}, nil
}

// defaultDynamoDBAccessor は、DynamoDBAccessor のデフォルト実装です。
type defaultDynamoDBAccessor struct {
	logger         logging.Logger
	config         myConfig.Config
	dynamodbClient *dynamodb.Client
}

// GetDynamoDBClient implements DynamoDBAccessor.
func (da *defaultDynamoDBAccessor) GetDynamoDBClient() *dynamodb.Client {
	return da.dynamodbClient
}

// GetItemSdk implements DynamoDBAccessor.
func (da *defaultDynamoDBAccessor) GetItemSdk(input *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	if ReturnConsumedCapacity(da.config) {
		input.ReturnConsumedCapacity = types.ReturnConsumedCapacityTotal
	}
	output, err := da.dynamodbClient.GetItem(apcontext.Context, input, optFns...)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if output.ConsumedCapacity != nil {
		da.logger.Debug("GetItem[%s]消費キャパシティユニット:%f", *output.ConsumedCapacity.TableName, *output.ConsumedCapacity.CapacityUnits)
	}
	return output, nil
}

// QuerySdk implements DynamoDBAccessor.
func (da *defaultDynamoDBAccessor) QuerySdk(input *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	if ReturnConsumedCapacity(da.config) {
		input.ReturnConsumedCapacity = types.ReturnConsumedCapacityTotal
	}
	output, err := da.dynamodbClient.Query(apcontext.Context, input, optFns...)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if output.ConsumedCapacity != nil {
		da.logger.Debug("Query[%s]消費キャパシティユニット:%f", *output.ConsumedCapacity.TableName, *output.ConsumedCapacity.CapacityUnits)
	}
	return output, nil
}

// QueryPagesSdk implements DynamoDBAccessor.
func (da *defaultDynamoDBAccessor) QueryPagesSdk(input *dynamodb.QueryInput, fn func(*dynamodb.QueryOutput) bool,
	optFns ...func(*dynamodb.Options)) error {
	paginator := dynamodb.NewQueryPaginator(da.dynamodbClient, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(apcontext.Context, optFns...)
		if err != nil {
			return errors.WithStack(err)
		}
		if page.ConsumedCapacity != nil {
			da.logger.Debug("Query[%s]消費キャパシティユニット:%f", *page.ConsumedCapacity.TableName, *page.ConsumedCapacity.CapacityUnits)
		}
		if fn(page) {
			break
		}
	}
	return nil
}

// PutItemSdk implements DynamoDBAccessor.
func (da *defaultDynamoDBAccessor) PutItemSdk(input *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	if ReturnConsumedCapacity(da.config) {
		input.ReturnConsumedCapacity = types.ReturnConsumedCapacityTotal
	}
	output, err := da.dynamodbClient.PutItem(apcontext.Context, input, optFns...)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if output.ConsumedCapacity != nil {
		da.logger.Debug("PutItem[%s]消費キャパシティユニット:%f", *output.ConsumedCapacity.TableName, *output.ConsumedCapacity.CapacityUnits)
	}
	return output, nil

}

// UpdateItemSdk implements DynamoDBAccessor.
func (da *defaultDynamoDBAccessor) UpdateItemSdk(input *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
	if ReturnConsumedCapacity(da.config) {
		input.ReturnConsumedCapacity = types.ReturnConsumedCapacityTotal
	}
	output, err := da.dynamodbClient.UpdateItem(apcontext.Context, input, optFns...)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if output.ConsumedCapacity != nil {
		da.logger.Debug("UpdateItem[%s]消費キャパシティユニット:%f", *output.ConsumedCapacity.TableName, *output.ConsumedCapacity.CapacityUnits)
	}
	return output, nil
}

// DeleteItemSdk implements DynamoDBAccessor.
func (da *defaultDynamoDBAccessor) DeleteItemSdk(input *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
	if ReturnConsumedCapacity(da.config) {
		input.ReturnConsumedCapacity = types.ReturnConsumedCapacityTotal
	}
	output, err := da.dynamodbClient.DeleteItem(apcontext.Context, input, optFns...)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if output.ConsumedCapacity != nil {
		da.logger.Debug("DeleteItem[%s]消費キャパシティユニット:%f", *output.ConsumedCapacity.TableName, *output.ConsumedCapacity.CapacityUnits)
	}
	return output, nil
}

// BatchGetItemSdk implements DynamoDBAccessor.
func (da *defaultDynamoDBAccessor) BatchGetItemSdk(input *dynamodb.BatchGetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchGetItemOutput, error) {
	if ReturnConsumedCapacity(da.config) {
		input.ReturnConsumedCapacity = types.ReturnConsumedCapacityTotal
	}
	output, err := da.dynamodbClient.BatchGetItem(apcontext.Context, input, optFns...)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	for i, v := range output.ConsumedCapacity {
		da.logger.Debug("BatchGetItem(%d番目)[%s]消費キャパシティユニット:%f", i+1, *v.TableName, *v.CapacityUnits)
	}
	return output, nil
}

// BatchWriteItemSdk implements DynamoDBAccessor.
func (da *defaultDynamoDBAccessor) BatchWriteItemSdk(input *dynamodb.BatchWriteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchWriteItemOutput, error) {
	if ReturnConsumedCapacity(da.config) {
		input.ReturnConsumedCapacity = types.ReturnConsumedCapacityTotal
	}
	output, err := da.dynamodbClient.BatchWriteItem(apcontext.Context, input, optFns...)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	for i, v := range output.ConsumedCapacity {
		da.logger.Debug("BatchWriteItem(%d番目)[%s]消費キャパシティユニット:%f", i+1, *v.TableName, *v.CapacityUnits)
	}
	return output, nil
}
