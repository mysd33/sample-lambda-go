/*
apcontext パッケージは、アプリケーションで格納するコンテキスト領域の操作機能を扱うパッケージです。
*/
package apcontext

import (
	"context"

	"github.com/aws/aws-lambda-go/lambdacontext"
)

// Contextは、アプリケーションで格納するコンテキスト領域です。
var Context context.Context

// ContextKey は、Contextのキーを表す型です。
type ContextKey string

// GetLambdaContext は、Lambdaのコンテキストを取得します。
func GetLambdaContext(ctx context.Context) *lambdacontext.LambdaContext {
	lc, _ := lambdacontext.FromContext(ctx)
	return lc
}
