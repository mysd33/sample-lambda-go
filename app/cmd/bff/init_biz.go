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
	userRepository := repository.NewUserRepositoryForRestAPI(ac.GetHttpClient(), ac.GetLogger(), ac.GetConfig())
	todoRepository := repository.NewTodoRepositoryForRestAPI(ac.GetHttpClient(), ac.GetLogger(), ac.GetConfig())
	// Configからキュー名を取得する
	sampleQueueName := ac.GetConfig().Get("SampleQueueName")
	if sampleQueueName == "" {
		sampleQueueName = "SampleQueue"
	}
	sampleFifoQueueName := ac.GetConfig().Get("SampleFIFOQueueName")
	if sampleFifoQueueName == "" {
		sampleFifoQueueName = "SampleFIFOQueue.fifo"
	}
	asyncMessageRepository := repository.NewAsyncMessageRepository(ac.GetSQSAccessor(), sampleQueueName, sampleFifoQueueName)
	// サービスの作成
	bffService := service.New(ac.GetLogger(), ac.GetConfig(), userRepository, todoRepository, asyncMessageRepository)
	// コントローラの作成
	bffController := controller.New(ac.GetLogger(), ac.GetDynamoDBTransactionManager(), bffService)
	// ハンドラインタセプタの取得
	interceptor := ac.GetInterceptor()

	// エラー確認用サービスの作成
	errorTestService := errservice.New()
	errorTestContoller := errcontroller.New(ac.GetLogger(), errorTestService)

	// ginによるURLマッピング
	// ハンドラインタセプタ経由でコントローラのメソッドを呼び出し
	v1 := r.Group("/bff-api/v1")
	{
		v1.GET("/todo", interceptor.Handle(bffController.FindTodo))
		v1.POST("/users", interceptor.Handle(bffController.RegisterUser))
		v1.POST("/todo", interceptor.Handle(bffController.RegisterTodo))
		v1.POST("/todo-async", interceptor.Handle(bffController.RegisterTodosAsync))

		//エラー確認用
		v1.POST("/error/:errortype", interceptor.Handle(errorTestContoller.Execute))
	}
}
