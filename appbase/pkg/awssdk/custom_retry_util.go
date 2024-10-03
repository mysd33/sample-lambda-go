// awssdk パッケージは、AWS SDKを利用する際のユーティリティを提供します。
package awssdk

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// WithCustomRetryerDynamoDBOption は、カスタムのリトライ機能を追加するDynamoDBのオプションを生成します。
func WithCustomRetryerDynamoDBOption(isErrorRetryableFunc func(err error) bool) func(*dynamodb.Options) {
	retryer := NewAWSRetryer(isErrorRetryableFunc)
	return func(o *dynamodb.Options) {
		o.Retryer = retryer
	}
}

// WithCustomRetryerS3Option は、カスタムのリトライ機能を追加するS3のオプションを生成します。
func WithCustomRetryerS3Option(isErrorRetryableFunc func(err error) bool) func(*s3.Options) {
	retryer := NewAWSRetryer(isErrorRetryableFunc)
	return func(o *s3.Options) {
		o.Retryer = retryer
	}
}

// WithCustomRetryerSQSOption は、カスタムのリトライ機能を追加するSQSのオプションを生成します。
func WithCustomRetryerSQSOption(isErrorRetryableFunc func(err error) bool) func(*sqs.Options) {
	retryer := NewAWSRetryer(isErrorRetryableFunc)
	return func(o *sqs.Options) {
		o.Retryer = retryer
	}
}
