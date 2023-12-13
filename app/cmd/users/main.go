package main

import (
	"testing"

	"example.com/appbase/pkg/component"
	"example.com/appbase/pkg/handler"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
)

// ginadapter.GinLambdaをグローバルスコープで宣言
var ginLambda *ginadapter.GinLambda
var apiLambdaHandler *handler.APILambdaHandler

// コードルドスタート時の初期化処理
func init() {
	// 処理テストコード実行に、init関数が動作してしまうのを回避
	if testing.Testing() {
		return
	}

	// ApplicationContextの作成
	ac := component.NewApplicationContext()
	// 業務の初期化処理実行
	apiLambdaHandler = ac.GetAPILambdaHandler()
	r := apiLambdaHandler.GetDefaultGinEngine()
	initBiz(ac, r)
	ginLambda = ginadapter.New(r)
}

// Main関数
func main() {
	// API用Lambdaハンドラ関数で開始
	lambda.Start(apiLambdaHandler.Handle(ginLambda))
}
