package main

import (
	"app/internal/app/bff/controller"
	"app/internal/app/bff/service"
	"app/internal/pkg/message"
	"app/internal/pkg/repository"

	errcontroller "app/internal/app/errortest/controller"
	errservice "app/internal/app/errortest/service"

	"example.com/appbase/pkg/component"
	"github.com/gin-gonic/gin"
)

// 業務の初期化処理
func initBiz(ac component.ApplicationContext, r *gin.Engine) {
	// メッセージの設定
	ac.GetMessageSource().Add(message.Messages_yaml)
	// リポジトリの作成
	userRepository := repository.NewUserRepositoryForRestAPI(ac.GetHttpClient(), ac.GetLogger())
	todoRepository := repository.NewTodoRepositoryForRestAPI(ac.GetHttpClient(), ac.GetLogger())
	// サービスの作成
	bffService := service.New(ac.GetLogger(), ac.GetConfig(), userRepository, todoRepository)
	// コントローラの作成
	bffController := controller.New(ac.GetLogger(), bffService)
	// ハンドラインタセプタの取得
	interceptor := ac.GetInterceptor()

	// エラー確認用
	errorTestService := errservice.New()
	errorTestContoller := errcontroller.New(ac.GetLogger(), errorTestService)

	// ginによるURLマッピング
	// ハンドラインタセプタ経由でコントローラのメソッドを呼び出し
	v1 := r.Group("/bff-api/v1")
	{
		v1.GET("/todo", interceptor.Handle(bffController.FindTodo))
		v1.POST("/users", interceptor.Handle(bffController.RegisterUser))
		v1.POST("/todo", interceptor.Handle(bffController.RegisterTodo))

		//エラー確認用
		v1.POST("/error/:errortype", interceptor.Handle(errorTestContoller.Execute))
	}
}
