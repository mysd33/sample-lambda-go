package main

import (
	"app/internal/app/todo/controller"
	"app/internal/app/todo/service"
	"app/internal/pkg/repository"
	"context"

	"example.com/appbase/pkg/apcontext"
	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/interceptor"
	"example.com/appbase/pkg/logging"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
)

// ginadapter.GinLambdaをグローバルスコープで宣言
var ginLambda *ginadapter.GinLambda

// コードルドスタート時の初期化処理
func init() {
	log := logging.NewLogger()
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("初期化処理エラー:%s", err.Error())
		panic(err.Error())
	}
	// リポジトリの作成
	todoRepository, err := repository.NewTodoRepository(log)
	if err != nil {
		log.Fatal("初期化処理エラー:%s", err.Error())
		panic(err.Error())
	}
	// サービスの作成
	todoService := service.New(log, cfg, todoRepository)
	// コントローラの作成
	todoController := controller.New(log, todoService)
	// ハンドラインタセプタの作成
	interceptor := interceptor.New(log)

	// ginによるURLマッピング
	r := gin.Default()
	// ハンドラインタセプタ経由でコントローラのメソッドを呼び出し
	v1 := r.Group("/todo-api/v1")
	{
		v1.GET("/todo/:todo_id", interceptor.Handle(todoController.Find))
		v1.POST("/todo", interceptor.Handle(todoController.Regist))
	}
	ginLambda = ginadapter.New(r)
}

// Lambdaのハンドラメソッド
func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// ctxをコンテキスト領域に格納
	apcontext.Context = ctx

	// AWS Lambda Go API Proxyでginと統合
	// https://github.com/awslabs/aws-lambda-go-api-proxy
	return ginLambda.ProxyWithContext(ctx, request)
}

// Main関数
func main() {
	lambda.Start(handler)
}
