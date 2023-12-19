/*
dynamodb パッケージは、DynamoDBアクセスに関する機能を提供するパッケージです。
*/
package dynamodb

import (
	"context"

	myConfig "example.com/appbase/pkg/config"
	"example.com/appbase/pkg/constant"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
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
