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
	// GetItemSdkWithContext は、AWS SDKによるGetItemをラップします。goroutine向けに、渡されたContextを利用して実行します。
	GetItemSdkWithContext(ctx context.Context, input *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	// QuerySdk は、AWS SDKによるQueryをラップします。
	QuerySdk(input *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
	// QuerySdkWithContext は、AWS SDKによるQueryをラップします。goroutine向けに、渡されたContextを利用して実行します。
	QuerySDKWithContext(ctx context.Context, input *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
	// QueryPagesSdk は、AWS SDKによるQueryのページング処理をラップします。
	QueryPagesSdk(input *dynamodb.QueryInput, fn func(*dynamodb.QueryOutput) bool, optFns ...func(*dynamodb.Options)) error
	// QueryPagesSdkWithContext は、AWS SDKによるQueryのページング処理をラップします。goroutine向けに、渡されたContextを利用して実行します。
	QueryPagesSdkWithContext(ctx context.Context, input *dynamodb.QueryInput, fn func(*dynamodb.QueryOutput) bool, optFns ...func(*dynamodb.Options)) error
	// PutItemSdk は、AWS SDKによるPutItemをラップします。
	PutItemSdk(input *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	// PutItemSdkWithContext は、AWS SDKによるPutItemをラップします。goroutine向けに、渡されたContextを利用して実行します。
	PutItemSdkWithContext(ctx context.Context, input *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	// UpdateItemSdk は、AWS SDKによるUpdateItemをラップします。
	UpdateItemSdk(input *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
	// UpdateItemSdkWithContext は、AWS SDKによるUpdateItemをラップします。goroutine向けに、渡されたContextを利用して実行します。
	UpdateItemSdkWithContext(ctx context.Context, input *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
	// DeleteItemSdk は、AWS SDKによるDeleteItemをラップします。
	DeleteItemSdk(input *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error)
	// DeleteItemSdkWithContext は、AWS SDKによるDeleteItemをラップします。goroutine向けに、渡されたContextを利用して実行します。
	DeleteItemSdkWithContext(ctx context.Context, input *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error)
	// BatchGetItemSdk は、AWS SDKによるBatchGetItemをラップします。
	BatchGetItemSdk(input *dynamodb.BatchGetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchGetItemOutput, error)
	// BatchGetItemSdkWithContext は、AWS SDKによるBatchGetItemをラップします。goroutine向けに、渡されたContextを利用して実行します。
	BatchGetItemSdkWithContext(ctx context.Context, input *dynamodb.BatchGetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchGetItemOutput, error)
	// BatchWriteItemSdk は、AWS SDKによるBatchWriteItemをラップします。
	BatchWriteItemSdk(input *dynamodb.BatchWriteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchWriteItemOutput, error)
	// BatchWriteItemSdkWithContext は、AWS SDKによるBatchWriteItemをラップします。goroutine向けに、渡されたContextを利用して実行します。
	BatchWriteItemSdkWithContext(ctx context.Context, input *dynamodb.BatchWriteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchWriteItemOutput, error)
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
	return da.GetItemSdkWithContext(apcontext.Context, input, optFns...)
}

// GetItemSdkWithContext implements DynamoDBAccessor.
func (da *defaultDynamoDBAccessor) GetItemSdkWithContext(ctx context.Context, input *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	if ReturnConsumedCapacity(da.config) {
		input.ReturnConsumedCapacity = types.ReturnConsumedCapacityTotal
	}
	output, err := da.dynamodbClient.GetItem(ctx, input, optFns...)
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
	return da.QuerySDKWithContext(apcontext.Context, input, optFns...)
}

// QuerySDKWithContext implements DynamoDBAccessor.
func (da *defaultDynamoDBAccessor) QuerySDKWithContext(ctx context.Context, input *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	if ReturnConsumedCapacity(da.config) {
		input.ReturnConsumedCapacity = types.ReturnConsumedCapacityTotal
	}
	output, err := da.dynamodbClient.Query(ctx, input, optFns...)
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
	return da.QueryPagesSdkWithContext(apcontext.Context, input, fn, optFns...)
}

// QueryPagesSdkWithContext implements DynamoDBAccessor.
func (da *defaultDynamoDBAccessor) QueryPagesSdkWithContext(ctx context.Context, input *dynamodb.QueryInput, fn func(*dynamodb.QueryOutput) bool, optFns ...func(*dynamodb.Options)) error {
	paginator := dynamodb.NewQueryPaginator(da.dynamodbClient, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx, optFns...)
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
	return da.PutItemSdkWithContext(apcontext.Context, input, optFns...)
}

// PutItemSdkWithContext implements DynamoDBAccessor.
func (da *defaultDynamoDBAccessor) PutItemSdkWithContext(ctx context.Context, input *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	if ReturnConsumedCapacity(da.config) {
		input.ReturnConsumedCapacity = types.ReturnConsumedCapacityTotal
	}
	output, err := da.dynamodbClient.PutItem(ctx, input, optFns...)
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
	return da.UpdateItemSdkWithContext(apcontext.Context, input, optFns...)
}

// UpdateItemSdkWithContext implements DynamoDBAccessor.
func (da *defaultDynamoDBAccessor) UpdateItemSdkWithContext(ctx context.Context, input *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
	if ReturnConsumedCapacity(da.config) {
		input.ReturnConsumedCapacity = types.ReturnConsumedCapacityTotal
	}
	output, err := da.dynamodbClient.UpdateItem(ctx, input, optFns...)
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
	return da.DeleteItemSdkWithContext(apcontext.Context, input, optFns...)
}

// DeleteItemSdkWithContext implements DynamoDBAccessor.
func (da *defaultDynamoDBAccessor) DeleteItemSdkWithContext(ctx context.Context, input *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
	if ReturnConsumedCapacity(da.config) {
		input.ReturnConsumedCapacity = types.ReturnConsumedCapacityTotal
	}
	output, err := da.dynamodbClient.DeleteItem(ctx, input, optFns...)
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
	return da.BatchGetItemSdkWithContext(apcontext.Context, input, optFns...)
}

// BatchGetItemSdkWithContext implements DynamoDBAccessor.
func (da *defaultDynamoDBAccessor) BatchGetItemSdkWithContext(ctx context.Context, input *dynamodb.BatchGetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchGetItemOutput, error) {
	if ReturnConsumedCapacity(da.config) {
		input.ReturnConsumedCapacity = types.ReturnConsumedCapacityTotal
	}
	output, err := da.dynamodbClient.BatchGetItem(ctx, input, optFns...)
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
	return da.BatchWriteItemSdkWithContext(apcontext.Context, input, optFns...)
}

// BatchWriteItemSdkWithContext implements DynamoDBAccessor.
func (da *defaultDynamoDBAccessor) BatchWriteItemSdkWithContext(ctx context.Context, input *dynamodb.BatchWriteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchWriteItemOutput, error) {
	if ReturnConsumedCapacity(da.config) {
		input.ReturnConsumedCapacity = types.ReturnConsumedCapacityTotal
	}
	output, err := da.dynamodbClient.BatchWriteItem(ctx, input, optFns...)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	for i, v := range output.ConsumedCapacity {
		da.logger.Debug("BatchWriteItem(%d番目)[%s]消費キャパシティユニット:%f", i+1, *v.TableName, *v.CapacityUnits)
	}
	return output, nil
}
