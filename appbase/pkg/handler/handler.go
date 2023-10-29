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

// LambdaHandlerFuncは、Lambdaのハンドラメソッドを表す関数です。
type LambdaHandlerFunc func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)

// ApiLambdaHandlerは、APIGatewayトリガーのLambdaHandlerFuncです。
func ApiLambdaHandler(ginLambda *ginadapter.GinLambda) LambdaHandlerFunc {
	return func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		// ctxをコンテキスト領域に格納
		apcontext.Context = ctx

		// AWS Lambda Go API Proxyでginと統合
		// https://github.com/awslabs/aws-lambda-go-api-proxy
		return ginLambda.ProxyWithContext(ctx, request)
	}
}

// TODO:他のトリガのLambdaHandlerFuncの実装
func DelayedLambdaHandler() LambdaHandlerFunc {
	return func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		//　TODO: 実装
		panic("not implement")
	}
}

func ScheduledBatchLambdaHandler() LambdaHandlerFunc {
	return func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		//　TODO: 実装
		panic("not implement")
	}
}

func FlowBatchLambdaHandler() LambdaHandlerFunc {
	return func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		//　TODO: 実装
		panic("not implement")
	}
}
