package main

import (
	"app/internal/app/books/controller"
	"app/internal/app/books/service"
	"app/internal/pkg/message"
	"app/internal/pkg/repository"

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
	// TODO: スタブをDocumentDBアクセスの実装に置き換える
	// リポジトリの作成
	bookRepository := repository.NewBookRepository()
	// サービスの作成
	bookService := service.New(ac.GetLogger(), ac.GetConfig(), bookRepository)
	// コントローラの作成
	bookController := controller.New(ac.GetLogger(), bookService)
	// ハンドラインタセプタの作成
	interceptor := ac.GetInterceptor()

	// ginによるURLマッピング
	// ハンドラインタセプタ経由でコントローラのメソッドを呼び出し
	v1 := r.Group("/books-api/v1")
	{
		v1.GET("/books", interceptor.Handle(bookController.FindByCriteria))
		v1.POST("/books", interceptor.Handle(bookController.Register))
	}
}
