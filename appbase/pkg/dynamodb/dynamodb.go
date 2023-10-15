/*
dynamodb パッケージは、DynamoDBアクセスに関する機能を提供するパッケージです。
*/
package dynamodb

import (
	"context"
	"os"

	"example.com/appbase/pkg/domain"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-xray-sdk-go/instrumentation/awsv2"
)

func CreateDynamoDBClient() (*dynamodb.Client, error) {
	// リージョン名
	region := os.Getenv("REGION")

	// AWS SDK for Go v2 Migration
	// https://github.com/aws/aws-sdk-go-v2
	// https://aws.github.io/aws-sdk-go-v2/docs/migrating/
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region))
	if err != nil {
		return nil, err
	}
	// Instrumenting AWS SDK v2
	// https://github.com/aws/aws-xray-sdk-go
	awsv2.AWSV2Instrumentor(&cfg.APIOptions)
	return dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		// DynamoDB Local起動先が指定されている場合
		dynamodbEndpoint := os.Getenv("DYNAMODB_LOCAL_ENDPOINT")
		if dynamodbEndpoint != "" {
			o.BaseEndpoint = aws.String(dynamodbEndpoint)
		}
	}), nil
}

// ExecuteTransaction は、Serviceの関数serviceFuncの実行前後でDynamoDBトランザクション実行します。
func ExecuteTransaction(serviceFunc domain.ServiceFunc) (interface{}, error) {
	// サービスの実行
	result, err := serviceFunc()
	// TODO: DynamoDBトランザクションオブジェクトが存在する場合はトランザクション実行の実装
	// コンテキスト領域にTransactWriteItemsInputや
	// TransactGetItemsInputがある場合にまとめてトランザクション実行

	if err != nil {
		return nil, err
	}
	return result, nil
}
