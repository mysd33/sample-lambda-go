/*
handler パッケージは、Lambdaのハンドラメソッドに関する機能を提供するパッケージです。
*/
package handler

import (
	"context"
	"fmt"

	"example.com/appbase/pkg/apcontext"
	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/logging"
	"github.com/aws/aws-lambda-go/events"
)

// SQSTriggeredLambdaHandlerFuncは、SQSトリガのLambdaのハンドラを表す関数です。
type SQSTriggeredLambdaHandlerFunc func(ctx context.Context, event events.SQSEvent) (events.SQSEventResponse, error)

// AsyncLambdaHandler は、SQSトリガの非同期処理のLambdaのハンドラを管理する構造体です。
type AsyncLambdaHandler struct {
	config config.Config
	log    logging.Logger
}

// NewAsyncLambdaHandler は、AsyncLambdaHandlerを作成します。
func NewAsyncLambdaHandler(config config.Config,
	log logging.Logger) *AsyncLambdaHandler {
	return &AsyncLambdaHandler{
		config: config,
		log:    log,
	}
}

// Handleは、SQSトリガのLambdaのハンドラを実行します。
func (h *AsyncLambdaHandler) Handle(asyncControllerFunc AsyncControllerFunc) SQSTriggeredLambdaHandlerFunc {
	return func(ctx context.Context, event events.SQSEvent) (response events.SQSEventResponse, err error) {
		// 非同期処理の場合
		defer func() {
			if v := recover(); v != nil {
				err = fmt.Errorf("recover from: %+v", v)
				h.log.ErrorWithUnexpectedError(err)
				// 全てのメッセージを失敗扱いにする
				response.BatchItemFailures = append(response.BatchItemFailures, events.SQSBatchItemFailure{ItemIdentifier: ""})
			}
		}()
		// ctxをコンテキスト領域に格納
		apcontext.Context = ctx

		// TODO: DBとのデータ整合性の確認
		// TODO: 二重実行防止のチェック（メッセージIDの確認）
		// FIFOの対応（FIFOの場合はメッセージグループID毎にメッセージのソートも）

		for _, v := range event.Records {
			// SQSのメッセージを1件取得しコントローラを呼び出し
			err := asyncControllerFunc(v)
			if err != nil {
				// 失敗したメッセージIDをBatchItemFailuresに登録
				response.BatchItemFailures = append(response.BatchItemFailures, events.SQSBatchItemFailure{ItemIdentifier: v.MessageId})
			}
		}
		return
	}
}
