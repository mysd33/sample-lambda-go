/*
apcontext パッケージは、アプリケーションで格納するコンテキスト領域の操作機能を扱うパッケージです。
*/
package apcontext

import (
	"context"
)

// Contextは、アプリケーションで格納するコンテキスト領域です。
var Context context.Context

// ContextKey は、Contextのキーを表す型です。
type ContextKey string
