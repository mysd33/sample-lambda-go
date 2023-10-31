package main

import (
	"app/internal/app/todo/controller"
	"app/internal/app/todo/service"
	"app/internal/pkg/message"
	"app/internal/pkg/repository"

	"example.com/appbase/pkg/component"

	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
)

// 業務の初期化処理
func initBiz(ac component.ApplicationContext) *ginadapter.GinLambda {
	// メッセージの設定
	ac.GetMessageSource().Add(message.Messages_yaml)
	// リポジトリの作成
	todoRepository := repository.NewTodoRepositoryForDynamoDB(ac.GetDynamoDBAccessor(), ac.GetLogger())
	// サービスの作成
	todoService := service.New(ac.GetLogger(), ac.GetConfig(), todoRepository)
	// コントローラの作成
	todoController := controller.New(ac.GetLogger(), todoService)
	// ハンドラインタセプタの作成
	interceptor := ac.GetInterceptor()

	// ginによるURLマッピング
	r := gin.Default()
	// ハンドラインタセプタ経由でコントローラのメソッドを呼び出し
	v1 := r.Group("/todo-api/v1")
	{
		v1.GET("/todo/:todo_id", interceptor.Handle(todoController.Find))
		v1.POST("/todo", interceptor.Handle(todoController.Register))
	}
	return ginadapter.New(r)
}
