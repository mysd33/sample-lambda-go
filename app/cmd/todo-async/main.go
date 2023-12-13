package main

import (
	"testing"

	"example.com/appbase/pkg/component"
	"example.com/appbase/pkg/handler"
	"github.com/aws/aws-lambda-go/lambda"
)

// 非同期処理用のHandler関数をグローバル変数定義
var lambdaHandler handler.SQSTriggeredLambdaHandlerFunc

// コードルドスタート時の初期化処理
func init() {
	// 処理テストコード実行に、init関数が動作してしまうのを回避
	if testing.Testing() {
		return
	}
	// ApplicationContextの作成
	ac := component.NewApplicationContext()
	asyncLambdaHandler := ac.GetAsyncLambdaHandler()
	// 業務の初期化処理実行
	asyncControllerFunc := initBiz(ac)
	// ハンドラ関数の作成
	lambdaHandler = asyncLambdaHandler.Handle(asyncControllerFunc)
}

func main() {
	lambda.Start(lambdaHandler)
}
