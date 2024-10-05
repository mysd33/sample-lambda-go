/*
httpclient パッケージは、REST APIの呼び出し等のためのHTTPクライアントの機能を提供するパッケージです。
*/

package httpclient

import (
	"net/http"

	"example.com/appbase/pkg/retry"
)

// Option は、HTTPクライアントのFunctaional Optionパターンによるオプションの関数です。
type Option func(*Options)

// Options は、HTTPクライアント実行時のオプションを保持します。
type Options struct {
	// CheckRetrayable は、リトライ可能なエラーかどうかを判定する関数です。
	CheckRetrayable retry.CheckRetryable[*http.Response]
	// RetryOptions は、リトライ機能のオプションを保持します。
	RetryOptions []retry.Option
}

// WithCheckRetryable は、リトライ可能なエラーかどうかを判定する関数を追加するオプションを生成します。
func WithCheckRetryable(checkRetryable retry.CheckRetryable[*http.Response]) Option {
	return func(o *Options) {
		o.CheckRetrayable = checkRetryable
	}
}

// WithRetryOptions は、リトライ処理のオプションを追加するオプションを生成します。
func WithRetryOptions(retryOptions []retry.Option) Option {
	return func(o *Options) {
		o.RetryOptions = retryOptions
	}
}
