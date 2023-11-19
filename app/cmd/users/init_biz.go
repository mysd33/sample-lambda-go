package main

import (
	"app/internal/app/user/controller"
	"app/internal/app/user/service"
	"app/internal/pkg/message"
	"app/internal/pkg/repository"

	"example.com/appbase/pkg/component"
	"github.com/gin-gonic/gin"
)

// 業務の初期化処理
func initBiz(ac component.ApplicationContext, r *gin.Engine) {
	// メッセージの設定
	ac.GetMessageSource().Add(message.Messages_yaml)
	// リポジトリの作成（DynamoDBの場合）
	//userRepository := repository.NewUserRepositoryForDynamoDB(ac.GetDynamoDBAccessor(), ac.GetLogger(), ac.GetConfig())
	// リポジトリの作成（RDBの場合）
	userRepository := repository.NewUserRepositoryForRDB()
	// サービスの作成
	userService := service.New(ac.GetLogger(), ac.GetConfig(), userRepository)
	// コントローラの作成
	userController := controller.New(ac.GetLogger(), userService)
	// ハンドラインタセプタの取得
	interceptor := ac.GetInterceptor()

	// ginによるURLマッピング定義
	// ハンドラインタセプタ経由でコントローラのメソッドを呼び出し
	v1 := r.Group("/users-api/v1")
	{
		v1.GET("/users/:user_id", interceptor.Handle(userController.Find))
		v1.POST("/users", interceptor.Handle(userController.Register))
	}
}
