// awssdk パッケージは、AWS SDKを利用する際のユーティリティを提供します。
package awssdk

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
)

// customRetryer は、AWS SDKのカスタムのリトライ機能を提供する構造体です。
type customRetryer struct {
	// 標準のリトライを埋め込み
	retry.Standard
	// カスタムのリトライ判定関数
	isErrorRetryableFunc func(err error) bool
}

// NewAWSRetryer は、カスタムのリトライ機能を提供するAWS SDKのRetryerを生成します。
func NewAWSRetryer(isErrorRetryableFunc func(err error) bool) aws.Retryer {
	return &customRetryer{
		isErrorRetryableFunc: isErrorRetryableFunc,
	}
}

// IsErrorRetryable implements the Retryer interface.
func (r *customRetryer) IsErrorRetryable(err error) bool {
	if r.isErrorRetryableFunc(err) {
		return true
	}
	return r.Standard.IsErrorRetryable(err)
}
