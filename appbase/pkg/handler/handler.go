/*
handler パッケージは、Lambdaのハンドラメソッドに関する機能を提供するパッケージです。
*/
package handler

import (
	"context"

	"example.com/appbase/pkg/apcontext"
	"github.com/aws/aws-lambda-go/events"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
)

// APITriggeredLambdaHandlerFuncは、APIGatewayトリガのLambdaのハンドラメソッドを表す関数です。
type APITriggeredLambdaHandlerFunc func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)

// SQSTriggeredLambdaHandlerFuncは、SQSトリガのLambdaのハンドラメソッドを表す関数です。
type SQSTriggeredLambdaHandlerFunc func(ctx context.Context, event events.SQSEvent) (events.SQSEventResponse, error)

// SimpleLambdaHandlerFunc は、その他のトリガのLambdaのハンドラメソッドを表す関数です。
type SimpleLambdaHandlerFunc func(ctx context.Context) error

// ApiLambdaHandlerは、APIGatewayトリガーのLambdaHandlerFuncです。
func ApiLambdaHandler(ginLambda *ginadapter.GinLambda) APITriggeredLambdaHandlerFunc {
	return func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		// ctxをコンテキスト領域に格納
		apcontext.Context = ctx

		// AWS Lambda Go API Proxyでginと統合
		// https://github.com/awslabs/aws-lambda-go-api-proxy
		return ginLambda.ProxyWithContext(ctx, request)
	}
}

func AsyncLambdaHandler(asyncControllerFunc AsyncControllerFunc) SQSTriggeredLambdaHandlerFunc {
	return func(ctx context.Context, event events.SQSEvent) (events.SQSEventResponse, error) {
		// ctxをコンテキスト領域に格納
		apcontext.Context = ctx

		// TODO: DBとのデータ整合性の確認
		// TODO: 二重実行防止のチェック（メッセージIDの確認）
		// FIFOの対応（FIFOの場合はメッセージグループID毎にメッセージのソートも）

		var response events.SQSEventResponse
		for _, v := range event.Records {
			// SQSのメッセージを1件取得しコントローラを呼び出し
			err := asyncControllerFunc(v)
			if err != nil {
				// 失敗したメッセージIDをBatchItemFailuresに登録
				response.BatchItemFailures = append(response.BatchItemFailures, events.SQSBatchItemFailure{ItemIdentifier: v.MessageId})
			}
		}
		return response, nil
	}
}

// TODO:他のトリガのLambdaHandlerFuncの実装
func ScheduledBatchLambdaHandler() SimpleLambdaHandlerFunc {
	return func(ctx context.Context) error {
		//　TODO: 実装
		panic("not implement")
	}
}

func FlowBatchLambdaHandler() SimpleLambdaHandlerFunc {
	return func(ctx context.Context) error {
		//　TODO: 実装
		panic("not implement")
	}
}
