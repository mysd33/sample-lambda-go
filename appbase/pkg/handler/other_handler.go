/*
handler パッケージは、Lambdaのハンドラメソッドに関する機能を提供するパッケージです。
*/
package handler

import (
	"context"

	"example.com/appbase/pkg/apcontext"
	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/idempotency"
	"example.com/appbase/pkg/logging"
	"github.com/cockroachdb/errors"
)

const (
	// 当該機能により二重実行検知した場合に、当該ハンドラによるLambdaの戻り値を設定
	// 当該機能を使ったLambdaを実行し、StepFunctionsで定義したフロー上で重複エラーを判定できるようにする
	IDEMPOTENCCY_RESPONSE_NAME = "idempotency_check"
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
	logger logging.Logger
}

// NewSimpleLambdaHandler は、SimpleLambdaHandlerを作成します。
func NewSimpleLambdaHandler(config config.Config,
	logger logging.Logger) *SimpleLambdaHandler {
	return &SimpleLambdaHandler{
		config: config,
		logger: logger,
	}
}

// Handle は、その他のトリガのLambdaのハンドラを実行します。
func (h *SimpleLambdaHandler) Handle(simpleControllerFunc SimpleControllerFunc) SimpleLambdaHandlerFunc {
	return func(ctx context.Context, event any) (response any, resultErr error) {
		defer func() {
			// パニックのリカバリ処理
			if v := recover(); v != nil {
				resultErr = errors.Errorf("recover from: %+v", v)
				// パニックのスタックトレース情報をログ出力
				h.logger.ErrorWithUnexpectedError(resultErr)
			}
			// ログのフラッシュ
			h.logger.Sync()
		}()
		// ctxをコンテキスト領域に格納
		apcontext.Context = ctx
		// リクエストIDをログの付加情報として追加
		h.logger.ClearInfo()
		lc := apcontext.GetLambdaContext(ctx)
		h.logger.AddInfo("AWS RequestID", lc.AwsRequestID)

		response, resultErr = simpleControllerFunc(ctx, event)
		if resultErr != nil {
			if errors.Is(resultErr, idempotency.CompletedProcessIdempotencyError) || errors.Is(resultErr, idempotency.InprogressProcessIdempotencyError) {
				// 二重実行防止（冪等性）機能で、業務のController側で未ハンドリングの二重実行エラーを検知した場合は、
				// （実行中、実行済の処理かに関わらず）エラーをnilにし、正常終了とする
				// DynamoDBStreamsやKinesisDataStreams等、SimpleLambdaHandlerを使ったイベントソースマッピングのトリガに関しては、
				// 業務側のControllerで、エラーをハンドハンドリングしておく必要があるので注意すること。
				// ReportBatchItemFailuresによる一部再実行を行う場合は、業務側のControllerで、二重実行エラーが発生したレコードのIDをBatchItemFailuresに設定する
				// ReportBatchItemFailuresを使わない場合は、業務側のControllerで、別のエラー原因となるOtherErrorを返却し、再実行可能な配慮を行う
				resultErr = nil
				// 当該機能の結果を表す戻り値を設定しておくことで、
				// StepFunctionsで、当該機能を使ったLambdaを実行した場合に、フロー上で戻り値をもとに二重実行を判定できるようにしている
				response = map[string]any{
					IDEMPOTENCCY_RESPONSE_NAME: true,
				}
			}
		}

		return
	}
}
