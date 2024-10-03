/*
transaction パッケージは、トランザクション管理に関する機能を提供するパッケージです。
*/
package transaction

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// Options は、トランザクション実行時のオプションを保持します。
type Options struct {
	DynamoDBOptions []func(*dynamodb.Options)
	SqsOptions      []func(*sqs.Options)
}

// WithDynamoDBOptions は、DynamoDBのオプションを追加するオプションを生成します。
func WithDynamoDBOptions(options []func(*dynamodb.Options)) func(*Options) {
	return func(o *Options) {
		o.DynamoDBOptions = append(o.DynamoDBOptions, options...)
	}
}

// WithSQSOptions は、SQSのオプションを追加するオプションを生成します。
func WithSQSOptions(options []func(*sqs.Options)) func(*Options) {
	return func(o *Options) {
		o.SqsOptions = append(o.SqsOptions, options...)
	}
}
