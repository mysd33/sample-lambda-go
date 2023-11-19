package main

import (
	"testing"

	"example.com/appbase/pkg/component"
	"example.com/appbase/pkg/handler"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
)

// ginadapter.GinLambdaをグローバルスコープで宣言
var ginLambda *ginadapter.GinLambda

// コードルドスタート時の初期化処理
func init() {
	// 処理テストコード実行に、init関数が動作してしまうのを回避
	if testing.Testing() {
		return
	}
	// ApplicationContextの作成
	ac := component.NewApplicationContext()
	// 業務の初期化処理実行
	r := gin.Default()
	initBiz(ac, r)
	ginadapter.New(r)
	ginLambda = ginadapter.New(r)
}

// Main関数
func main() {
	// API用Lambdaハンドラ関数で開始
	lambda.Start(handler.ApiLambdaHandler(ginLambda))
}
