/*
dynamodb パッケージは、DynamoDBアクセスに関する機能を提供するパッケージです。
*/
package dynamodb

import (
	"context"

	"example.com/appbase/pkg/apcontext"
	myConfig "example.com/appbase/pkg/config"
	"example.com/appbase/pkg/constant"
	"example.com/appbase/pkg/env"
	"example.com/appbase/pkg/logging"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-xray-sdk-go/instrumentation/awsv2"
	"github.com/cockroachdb/errors"
)

// CreateDynamoDBClient は、DynamoDBClientを作成します。
func CreateDynamoDBClient(myCfg myConfig.Config) (*dynamodb.Client, error) {
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
		// ローカル実行のためDynamoDB Local起動先が指定されている場合
		dynamodbEndpoint := myCfg.Get(constant.DYNAMODB_LOCAL_ENDPOINT_NAME)
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
}

func NewDynamoDBAccessor(log logging.Logger, myCfg myConfig.Config) (DynamoDBAccessor, error) {
	dynamodbClient, err := CreateDynamoDBClient(myCfg)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &defaultDynamoDBAccessor{log: log, dynamodbClient: dynamodbClient}, nil
}

type defaultDynamoDBAccessor struct {
	log            logging.Logger
	dynamodbClient *dynamodb.Client
}

// GetDynamoDBClient implements DynamoDBAccessor.
func (da *defaultDynamoDBAccessor) GetDynamoDBClient() *dynamodb.Client {
	return da.dynamodbClient
}

// GetItemSdk implements DynamoDBAccessor.
func (da *defaultDynamoDBAccessor) GetItemSdk(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
	if !env.IsStragingOrProd() {
		input.ReturnConsumedCapacity = types.ReturnConsumedCapacityTotal
	}
	output, err := da.dynamodbClient.GetItem(apcontext.Context, input)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if output.ConsumedCapacity != nil {
		da.log.Debug("GetItem[%s]消費キャパシティユニット:%f", *output.ConsumedCapacity.TableName, *output.ConsumedCapacity.CapacityUnits)
	}
	return output, nil
}

// QuerySdk implements DynamoDBAccessor.
func (da *defaultDynamoDBAccessor) QuerySdk(input *dynamodb.QueryInput) (*dynamodb.QueryOutput, error) {
	if !env.IsStragingOrProd() {
		input.ReturnConsumedCapacity = types.ReturnConsumedCapacityTotal
	}
	output, err := da.dynamodbClient.Query(apcontext.Context, input)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if output.ConsumedCapacity != nil {
		da.log.Debug("Query[%s]消費キャパシティユニット:%f", *output.ConsumedCapacity.TableName, *output.ConsumedCapacity.CapacityUnits)
	}
	return output, nil
}

// QueryPagesSdk implements DynamoDBAccessor.
func (da *defaultDynamoDBAccessor) QueryPagesSdk(input *dynamodb.QueryInput, fn func(*dynamodb.QueryOutput) bool) error {
	paginator := dynamodb.NewQueryPaginator(da.dynamodbClient, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(apcontext.Context)
		if err != nil {
			return errors.WithStack(err)
		}
		if page.ConsumedCapacity != nil {
			da.log.Debug("Query[%s]消費キャパシティユニット:%f", *page.ConsumedCapacity.TableName, *page.ConsumedCapacity.CapacityUnits)
		}
		if fn(page) {
			break
		}
	}
	return nil
}

// PutItemSdk implements DynamoDBAccessor.
func (da *defaultDynamoDBAccessor) PutItemSdk(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	if !env.IsStragingOrProd() {
		input.ReturnConsumedCapacity = types.ReturnConsumedCapacityTotal
	}
	output, err := da.dynamodbClient.PutItem(apcontext.Context, input)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if output.ConsumedCapacity != nil {
		da.log.Debug("PutItem[%s]消費キャパシティユニット:%f", *output.ConsumedCapacity.TableName, *output.ConsumedCapacity.CapacityUnits)
	}
	return output, nil

}

// UpdateItemSdk implements DynamoDBAccessor.
func (da *defaultDynamoDBAccessor) UpdateItemSdk(input *dynamodb.UpdateItemInput) (*dynamodb.UpdateItemOutput, error) {
	if !env.IsStragingOrProd() {
		input.ReturnConsumedCapacity = types.ReturnConsumedCapacityTotal
	}
	output, err := da.dynamodbClient.UpdateItem(apcontext.Context, input)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if output.ConsumedCapacity != nil {
		da.log.Debug("UpdateItem[%s]消費キャパシティユニット:%f", *output.ConsumedCapacity.TableName, *output.ConsumedCapacity.CapacityUnits)
	}
	return output, nil
}

// DeleteItemSdk implements DynamoDBAccessor.
func (da *defaultDynamoDBAccessor) DeleteItemSdk(input *dynamodb.DeleteItemInput) (*dynamodb.DeleteItemOutput, error) {
	if !env.IsStragingOrProd() {
		input.ReturnConsumedCapacity = types.ReturnConsumedCapacityTotal
	}
	output, err := da.dynamodbClient.DeleteItem(apcontext.Context, input)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if output.ConsumedCapacity != nil {
		da.log.Debug("DeleteItem[%s]消費キャパシティユニット:%f", *output.ConsumedCapacity.TableName, *output.ConsumedCapacity.CapacityUnits)
	}
	return output, nil
}

// BatchGetItemSdk implements DynamoDBAccessor.
func (da *defaultDynamoDBAccessor) BatchGetItemSdk(input *dynamodb.BatchGetItemInput) (*dynamodb.BatchGetItemOutput, error) {
	if !env.IsStragingOrProd() {
		input.ReturnConsumedCapacity = types.ReturnConsumedCapacityTotal
	}
	output, err := da.dynamodbClient.BatchGetItem(apcontext.Context, input)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	for i, v := range output.ConsumedCapacity {
		da.log.Debug("BatchGetItem(%d)[%s]消費キャパシティユニット:%f", i, *v.TableName, *v.CapacityUnits)
	}
	return output, nil
}

// BatchWriteItemSdk implements DynamoDBAccessor.
func (da *defaultDynamoDBAccessor) BatchWriteItemSdk(input *dynamodb.BatchWriteItemInput) (*dynamodb.BatchWriteItemOutput, error) {
	if !env.IsStragingOrProd() {
		input.ReturnConsumedCapacity = types.ReturnConsumedCapacityTotal
	}
	output, err := da.dynamodbClient.BatchWriteItem(apcontext.Context, input)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	for i, v := range output.ConsumedCapacity {
		da.log.Debug("BatchWriteItem(%d)[%s]消費キャパシティユニット:%f", i, *v.TableName, *v.CapacityUnits)
	}
	return output, nil
}
