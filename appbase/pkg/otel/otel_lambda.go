/*
otel パッケージは、ADOT（AWS Distro for OpenTelemetry）を扱うためのパッケージです。
*/
package otel

import (
	"context"
	"fmt"
	"os"

	"example.com/appbase/pkg/env"
	"github.com/aws-observability/aws-otel-go/exporters/xrayudp"
	"github.com/aws/aws-lambda-go/lambda"
	lambdadetector "go.opentelemetry.io/contrib/detectors/aws/lambda"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-lambda-go/otellambda"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-lambda-go/otellambda/xrayconfig"
	"go.opentelemetry.io/contrib/propagators/aws/xray"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// StartLambda は、main関数で使用される関数です。ADOTに対応しLambda関数のトレースを開始します。
func StartLambda(lambdaHandler any) {
	// ローカル実行のときにはADOTが動作しないように起動
	if env.IsLocalOrLocalTest() {
		// Lambdaハンドラ関数を開始
		lambda.Start(lambdaHandler)
		return
	}

	// クラウド環境での実行の場合は、ADOTに対応
	// https://docs.aws.amazon.com/xray/latest/devguide/manual-instrumentation-go.html#lambda-instrumentation
	ctx := context.Background()

	// ResourceDetectorの作成
	detector := lambdadetector.NewResourceDetector()
	lambdaResource, err := detector.Detect(context.Background())
	if err != nil {
		fmt.Printf("Lambdaリソースの検出に失敗しました: %+v", err)
		// フォールバックして、通常のLambdaハンドラ関数で開始
		lambda.Start(lambdaHandler)
		return
	}
	// サービス名をLambda関数名に設定するための属性を作成し、Resourceに追加
	var attributes = []attribute.KeyValue{
		{Key: semconv.ServiceNameKey, Value: attribute.StringValue(os.Getenv("AWS_LAMBDA_FUNCTION_NAME"))},
	}
	customResource := resource.NewWithAttributes(semconv.SchemaURL, attributes...)
	mergedResource, _ := resource.Merge(lambdaResource, customResource)

	// TracerProviderの作成
	xrayUdpExporter, _ := xrayudp.NewSpanExporter(ctx)
	tp := trace.NewTracerProvider(
		trace.WithSpanProcessor(trace.NewSimpleSpanProcessor(xrayUdpExporter)),
		trace.WithResource(mergedResource),
	)

	defer func(ctx context.Context) {
		// main関数の終了時にTracerProviderをシャットダウンさせる
		err := tp.Shutdown(ctx)
		if err != nil {
			fmt.Printf("TracerProviderのシャットダウンに失敗しました: %+v", err)
		}
	}(ctx)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(xray.Propagator{})

	// Lambdaハンドラ関数を開始
	lambda.Start(otellambda.InstrumentHandler(lambdaHandler, xrayconfig.WithRecommendedOptions(tp)...))
}
