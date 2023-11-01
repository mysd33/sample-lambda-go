package main

import (
	"app/internal/app/bff/controller"
	"app/internal/app/bff/service"
	"app/internal/pkg/repository"

	"example.com/appbase/pkg/component"
	"github.com/gin-gonic/gin"
)

// 業務の初期化処理
func initBiz(ac component.ApplicationContext, r *gin.Engine) {
	// リポジトリの作成
	userRepository := repository.NewUserRepositoryForRestAPI(ac.GetHttpClient(), ac.GetLogger())
	todoRepository := repository.NewTodoRepositoryForRestAPI(ac.GetHttpClient(), ac.GetLogger())
	// サービスの作成
	bffService := service.New(ac.GetLogger(), ac.GetConfig(), userRepository, todoRepository)
	// コントローラの作成
	bffController := controller.New(ac.GetLogger(), bffService)
	// ハンドラインタセプタの取得
	interceptor := ac.GetInterceptor()

	// ginによるURLマッピング
	// ハンドラインタセプタ経由でコントローラのメソッドを呼び出し
	v1 := r.Group("/bff-api/v1")
	{
		v1.GET("/todo", interceptor.Handle(bffController.FindTodo))
		v1.POST("/users", interceptor.Handle(bffController.RegisterUser))
		v1.POST("/todo", interceptor.Handle(bffController.RegisterTodo))
	}
}
