package main

import (
	"app/internal/app/todo-async/controller"
	"app/internal/app/todo-async/service"
	"app/internal/pkg/message"
	"app/internal/pkg/repository"

	"example.com/appbase/pkg/component"
	"example.com/appbase/pkg/handler"
)

// 業務の初期化処理
func initBiz(ac component.ApplicationContext) handler.AsyncControllerFunc {
	// メッセージの設定
	ac.GetMessageSource().Add(message.Messages_yaml)
	// リポジトリの作成
	todoRepository := repository.NewTodoRepositoryForDynamoDB(ac.GetDynamoDBAccessor(), ac.GetLogger(), ac.GetConfig())
	// サービスの作成
	todoAsyncService := service.New(ac.GetLogger(), ac.GetConfig(), todoRepository)
	// コントローラの作成
	controller := controller.New(ac.GetLogger(), ac.GetDynamoDBTransactionManager(), todoAsyncService)
	// ハンドラインタセプタの取得
	interceptor := ac.GetInterceptor()
	// ハンドラインタセプタ経由でコントローラのメソッドを呼び出し
	return interceptor.HandleAsync(controller.RegisterAll)

}
