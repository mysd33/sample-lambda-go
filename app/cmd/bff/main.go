package main

import (
	"app/cmd/common"
	"testing"

	"example.com/appbase/pkg/component"
	"example.com/appbase/pkg/handler"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
)

var lambdaHandler handler.APITriggeredLambdaHandlerFunc

// コードルドスタート時の初期化処理
func init() {
	// 処理テストコード実行に、init関数が動作してしまうのを回避
	if testing.Testing() {
		return
	}

	// ApplicationContextの作成
	ac := component.NewApplicationContext()
	// Ginのエンジンを作成
	apiLambdaHandler := ac.GetAPILambdaHandler()
	r := apiLambdaHandler.GetDefaultGinEngine(common.NewCommonErrorResponse(ac.GetMessageSource()))
	// 業務の初期化処理実行
	initBiz(ac, r)
	// ハンドラ関数の作成
	ginLambda := ginadapter.New(r)
	lambdaHandler = apiLambdaHandler.Handle(ginLambda)
}

// Main関数
func main() {
	// API用Lambdaハンドラ関数で開始
	lambda.Start(lambdaHandler)
}
