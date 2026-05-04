package main

import (
	"context"
	"fmt"
	"testing"

	"example.com/appbase/pkg/component"
	"example.com/appbase/pkg/env"
	"example.com/appbase/pkg/handler"
	"github.com/aws/aws-lambda-go/lambda"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-lambda-go/otellambda"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-lambda-go/otellambda/xrayconfig"
	"go.opentelemetry.io/contrib/propagators/aws/xray"
	"go.opentelemetry.io/otel"
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
	// ローカル実行のときにはADOTが動作しないようにOS環境変数で設定切り替え
	if env.IsLocalOrLocalTest() {
		// API用Lambdaハンドラ関数で開始
		lambda.Start(lambdaHandler)
		return
	}

	// クラウド環境での実行の場合は、ADOTに対応
	// https://docs.aws.amazon.com/xray/latest/devguide/manual-instrumentation-go.html#lambda-instrumentation
	// TODO: ソフトウェアフレームワーク側での部品化を検討。
	ctx := context.Background()
	tp, err := xrayconfig.NewTracerProvider(ctx)
	if err != nil {
		fmt.Printf("error creating tracer provider: %v", err)
	}
	defer func(ctx context.Context) {
		// main関数の終了時にTracerProviderをシャットダウンする。
		err := tp.Shutdown(ctx)
		if err != nil {
			fmt.Printf("error shutting down tracer provider: %v", err)
		}
	}(ctx)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(xray.Propagator{})

	lambda.Start(otellambda.InstrumentHandler(lambdaHandler, xrayconfig.WithRecommendedOptions(tp)...))
}
