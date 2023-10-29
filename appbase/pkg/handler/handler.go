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

type LambdaHandlerFunc func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)

func ApiLambdaHandler(ginLambda *ginadapter.GinLambda) LambdaHandlerFunc {
	return func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		// ctxをコンテキスト領域に格納
		apcontext.Context = ctx

		// AWS Lambda Go API Proxyでginと統合
		// https://github.com/awslabs/aws-lambda-go-api-proxy
		return ginLambda.ProxyWithContext(ctx, request)
	}
}
