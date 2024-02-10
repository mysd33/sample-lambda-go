/*
handler パッケージは、Lambdaのハンドラメソッドに関する機能を提供するパッケージです。
*/
package handler

import (
	"context"

	"example.com/appbase/pkg/apcontext"
	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/logging"
	"github.com/cockroachdb/errors"
)

// SimpleLambdaHandlerFunc は、その他のトリガのLambdaのハンドラを表す関数です。
type SimpleLambdaHandlerFunc func(ctx context.Context, event any) (any, error)

// SimpleLambdaHandlerGenericFunc は、その他のトリガのLambdaのハンドラを表すジェネリクス対応関数です。
// Lambdaハンドラの引数がany型の場合、map[string]any型にマッピングされてしまうため
// 型パラメータで指定した型に変換できるようにするための関数です。
type SimpleLambdaHandlerGenericFunc[T any] func(ctx context.Context, event T) (any, error)

// SimpleLambdaHandlerGenericFuncAdapter は、SimpleLambdaHandlerFuncを型パラメータで指定したジェネリクス関数に変換します。
func SimpleLambdaHandlerGenericFuncAdapter[T any](f SimpleLambdaHandlerFunc) SimpleLambdaHandlerGenericFunc[T] {
	return func(ctx context.Context, event T) (any, error) {
		return f(ctx, event)
	}
}

// SimpleLambdaHandler は、その他のトリガのLambdaのハンドラを表す構造体です。
type SimpleLambdaHandler struct {
	config config.Config
	log    logging.Logger
}

// NewSimpleLambdaHandler は、SimpleLambdaHandlerを作成します。
func NewSimpleLambdaHandler(config config.Config,
	log logging.Logger) *SimpleLambdaHandler {
	return &SimpleLambdaHandler{
		config: config,
		log:    log,
	}
}

// Handle は、その他のトリガのLambdaのハンドラを実行します。
func (h *SimpleLambdaHandler) Handle(simpleControllerFunc SimpleControllerFunc) SimpleLambdaHandlerFunc {
	return func(ctx context.Context, event any) (response any, resultErr error) {
		// パニックのリカバリ処理
		defer func() {
			if v := recover(); v != nil {
				resultErr = errors.Errorf("recover from: %+v", v)
				// パニックのスタックトレース情報をログ出力
				h.log.ErrorWithUnexpectedError(resultErr)
			}
		}()
		// ctxをコンテキスト領域に格納
		apcontext.Context = ctx
		// リクエストIDをログの付加情報として追加
		h.log.ClearInfo()
		lc := apcontext.GetLambdaContext(ctx)
		h.log.AddInfo("AWS RequestID", lc.AwsRequestID)

		response, resultErr = simpleControllerFunc(ctx, event)
		return
	}
}
