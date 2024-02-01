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
