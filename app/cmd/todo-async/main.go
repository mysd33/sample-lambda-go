package main

import (
	"testing"

	"example.com/appbase/pkg/component"
	"example.com/appbase/pkg/handler"
	"github.com/aws/aws-lambda-go/lambda"
)

// 非同期処理用のControllerの関数をグローバル変数定義
var asyncControllerFunc handler.AsyncControllerFunc

// コードルドスタート時の初期化処理
func init() {
	// 処理テストコード実行に、init関数が動作してしまうのを回避
	if testing.Testing() {
		return
	}
	// ApplicationContextの作成
	ac := component.NewApplicationContext()
	// 業務の初期化処理実行
	asyncControllerFunc = initBiz(ac)
}

func main() {
	lambda.Start(handler.AsyncLambdaHandler(asyncControllerFunc))
}
