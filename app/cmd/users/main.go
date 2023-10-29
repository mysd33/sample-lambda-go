package main

import (
	"example.com/appbase/pkg/component"
	"example.com/appbase/pkg/handler"

	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
)

// ginadapter.GinLambdaをグローバルスコープで宣言
var ginLambda *ginadapter.GinLambda

// コードルドスタート時の初期化処理
func init() {
	// ApplicationContextの作成
	ac := component.NewApplicationContext()
	// 業務の初期化処理実行
	ginLambda = initBiz(ac)
}

// Main関数
func main() {
	// API用Lambdaハンドラ関数で開始
	lambda.Start(handler.ApiLambdaHandler(ginLambda))
}
