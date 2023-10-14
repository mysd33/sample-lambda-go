package main

import (
	"context"

	"example.com/appbase/pkg/apcontext"
	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/interceptor"
	"example.com/appbase/pkg/logging"

	"app/internal/app/user/controller"
	"app/internal/app/user/service"
	"app/internal/pkg/repository"

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
	// リポジトリの作成（DynamoDBの場合）
	// userRepository, err := repository.NewUserRepositoryForDynamoDB()
	// if err != nil {
	//	log.Fatal("初期化処理エラー:%s", err.Error())
	//	panic(err.Error())
	// }
	// リポジトリの作成（RDBの場合）
	userRepository := repository.NewUserRepositoryForRDB()
	// サービスの作成
	userService := service.New(log, cfg, &userRepository)
	// コントローラの作成
	userController := controller.New(log, &userService)
	// ハンドラインタセプタの作成
	interceptor := interceptor.New(log)

	// ginによるURLマッピング定義
	r := gin.Default()
	// ハンドラインタセプタ経由でコントローラのメソッドを呼び出し
	r.GET("/users/:user_id", interceptor.Handle(userController.Find))
	r.POST("/users", interceptor.Handle(userController.Regist))
	ginLambda = ginadapter.New(r)
}

// ハンドラメソッド
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
