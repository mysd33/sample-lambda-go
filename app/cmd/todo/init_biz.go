package main

import (
	"app/internal/app/todo/controller"
	"app/internal/app/todo/service"
	"app/internal/pkg/message"
	"app/internal/pkg/repository"

	"example.com/appbase/pkg/component"

	"github.com/gin-gonic/gin"
)

// 業務の初期化処理
func initBiz(ac component.ApplicationContext, r *gin.Engine) {
	// メッセージの設定
	ac.GetMessageSource().Add(message.Messages_yaml)
	// リポジトリの作成
	todoRepository := repository.NewTodoRepositoryForDynamoDB(ac.GetDynamoDBAccessor(), ac.GetLogger(), ac.GetConfig())
	// サービスの作成
	todoService := service.New(ac.GetLogger(), ac.GetConfig(), todoRepository)
	// コントローラの作成
	todoController := controller.New(ac.GetLogger(), ac.GetDynamoDBTransactionManager(), todoService)
	// ハンドラインタセプタの作成
	interceptor := ac.GetInterceptor()

	// ginによるURLマッピング
	// ハンドラインタセプタ経由でコントローラのメソッドを呼び出し
	v1 := r.Group("/todo-api/v1")
	{
		v1.GET("/todo/:todo_id", interceptor.Handle(todoController.Find))
		v1.POST("/todo", interceptor.Handle(todoController.Register))
	}
}
