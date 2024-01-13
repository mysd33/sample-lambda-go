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
	err := ac.GetMessageSource().Add(message.Messages_yaml)
	if err != nil {
		panic(err)
	}
	// リポジトリの作成
	userRepository := repository.NewUserRepositoryForRestAPI(ac.GetHttpClient(), ac.GetLogger(), ac.GetConfig())
	todoRepository := repository.NewTodoRepositoryForRestAPI(ac.GetHttpClient(), ac.GetLogger(), ac.GetConfig())
	tempRepository := repository.NewTempRepository(ac.GetDynamoDBTemplate(), ac.GetDynamoDBAccessor(),
		ac.GetLogger(), ac.GetConfig(), ac.GetIDGenerator())
	// Configからキュー名を取得する
	sampleQueueName := ac.GetConfig().Get("SampleQueueName", "SampleQueue")
	ac.GetLogger().Debug("SampleQueueName:%s", sampleQueueName)
	sampleFifoQueueName := ac.GetConfig().Get("SampleFIFOQueueName", "SampleFIFOQueue.fifo")
	ac.GetLogger().Debug("SampleFIFOQueueName:%s", sampleFifoQueueName)
	asyncMessageRepository := repository.NewAsyncMessageRepository(ac.GetSQSTemplate(), sampleQueueName, sampleFifoQueueName)
	// サービスの作成
	bffService := service.New(ac.GetLogger(), ac.GetConfig(), ac.GetIDGenerator(), userRepository, todoRepository, tempRepository, asyncMessageRepository)
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
