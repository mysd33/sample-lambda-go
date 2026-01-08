// controllerのパッケージ
package controller

import (
	"app/internal/app/bff/service"
	"app/internal/pkg/message"
	"app/internal/pkg/model"
	"app/internal/pkg/repository"

	"example.com/appbase/pkg/domain"
	"example.com/appbase/pkg/errors"
	"example.com/appbase/pkg/logging"
	"example.com/appbase/pkg/transaction"
	"github.com/gin-gonic/gin"
)

// RequestFindTodo は、TODO検索のREST APIで受け取るリクエストデータの構造体です。
type RequestFindTodo struct {
	UserId string `label:"ユーザID(user_id)" form:"user_id" binding:"required"`
	TodoId string `label:"Todo ID(todo_id)" form:"todo_id" binding:"required"`
}

// ResponseFindTodo は、TODO検索のREST APIで受け取るレスポンスデータの構造体です。
type ResponseFindTodo struct {
	User *model.User `json:"user"`
	Todo *model.Todo `json:"todo"`
}

// RequestRegisterUser は、ユーザ登録のREST APIで受け取るリクエストデータの構造体です。
type RequestRegisterUser struct {
	Name string `label:"ユーザ名(user_name)" json:"user_name" binding:"required"`
}

// RequestRegisterTodo は、TODO登録のREST APIで受け取るリクエストデータの構造体です。
type RequestRegisterTodo struct {
	// TodoTitle は、Todoのタイトルです。
	TodoTitle string `label:"Todoタイトル(todo_title)" json:"todo_title" binding:"required"`
}

// RequestRegisterTodoAsync は、TODO一括登録のREST APIで受け取るリクエストデータの構造体です。
type RequestRegisterTodoAsync struct {
	TodoTitles []string `json:"todo_titles" binding:"required"`
}

// ResponseRegisterTodoAsync は、TODO一括登録のREST APIで受け取るレスポンスデータの構造体です。
type ResponseRegisterTodoAsync struct {
	Result string `json:"result"`
}

// RequestFindBook は、書籍検索のREST APIで受け取るリクエストデータの構造体です。
type RequestFindBook struct {
	// Title は、書籍のタイトルです。
	Title string `label:"タイトル" form:"title"`
	// Author は、書籍の著者です。
	Author string `label:"著者" form:"author"`
	// Publisher は、書籍の出版社です。
	Publisher string `label:"出版社" form:"publisher"`
}

// RequestRegisterBook は、書籍登録のREST APIで受け取るリクエストデータの構造体です。
type RequestRegisterBook struct {
	// Title は、書籍のタイトルです。
	Title string `label:"タイトル" json:"title" binding:"required"`
	// Author は、書籍の著者です。
	Author string `label:"著者" json:"author" binding:"required"`
	// Publisher は、書籍の出版社です。
	Publisher string `label:"出版社" json:"publisher"`
	// PublishedDate は、書籍の発売日です。
	PublishedDate string `label:"発売日" json:"published_date" binding:"datetime=2006-01-02"`
	// ISBNは、書籍のISBNです。
	ISBN string `label:"ISBN" json:"isbn"`
}

// BffController は、Bff業務のControllerインタフェースです。
type BffController interface {
	// FindTodo は、クエリパラメータで指定されたtodo_idとuser_idのTodoを照会します。
	FindTodo(ctx *gin.Context) (any, error)
	// RegisterUser は、リクエストデータで受け取ったユーザ情報を登録します。
	RegisterUser(ctx *gin.Context) (any, error)
	// RegisterTodo は、リクエストデータで受け取ったTodoを登録します。
	RegisterTodo(ctx *gin.Context) (any, error)
	// RegisterTodoAsync は、リクエストデータで受け取ったTodoのリストを非同期で登録します。
	RegisterTodosAsync(ctx *gin.Context) (any, error)
	// FindBooksByCriteria は、クエリパラメータで指定された検索条件に合致する書籍を検索します。
	FindBooksByCriteria(ctx *gin.Context) (any, error)
	// RegisterBook は、リクエストデータで受け取った書籍を登録します。
	RegisterBook(ctx *gin.Context) (any, error)
}

// New は、BffControllerを作成します。
func New(logger logging.Logger, transactionManager transaction.TransactionManager, service service.BffService) BffController {
	return &bffControllerImpl{logger: logger, transactionManager: transactionManager, service: service}
}

// bffControllerImpl は、BffControllerを実装する構造体です。
type bffControllerImpl struct {
	logger             logging.Logger
	transactionManager transaction.TransactionManager
	service            service.BffService
}

// FindTodo implements BffController.
func (c *bffControllerImpl) FindTodo(ctx *gin.Context) (any, error) {
	// クエリパラメータの取得
	var request RequestFindTodo
	if err := ctx.ShouldBindQuery(&request); err != nil {
		// 入力チェックエラーのハンドリング
		return nil, errors.NewValidationErrorWithCause(err, message.W_EX_5001)
	}

	// サービスの実行（DynamoDBトランザクション管理なし）
	user, todo, err := c.service.FindTodo(request.UserId, request.TodoId)
	if err != nil {
		return nil, err
	}
	return &ResponseFindTodo{User: user, Todo: todo}, nil

}

// RegisterUser implements BffController.
func (c *bffControllerImpl) RegisterUser(ctx *gin.Context) (any, error) {
	// POSTデータをバインド
	var request RequestRegisterUser
	if err := ctx.ShouldBindJSON(&request); err != nil {
		// 入力チェックエラーのハンドリング
		return nil, errors.NewValidationErrorWithCause(err, message.W_EX_5001)
	}

	// サービスの実行
	return c.service.RegisterUser(request.Name)
}

// RegisterTodo implements BffController.
func (c *bffControllerImpl) RegisterTodo(ctx *gin.Context) (any, error) {
	// POSTデータをバインド
	var request RequestRegisterTodo
	if err := ctx.ShouldBindJSON(&request); err != nil {
		// 入力チェックエラーのハンドリング
		return nil, errors.NewValidationErrorWithCause(err, message.W_EX_5001)
	}

	// サービスの実行
	return c.service.RegisterTodo(request.TodoTitle)
}

// RegisterTodosAsync implements BffController.
func (c *bffControllerImpl) RegisterTodosAsync(ctx *gin.Context) (any, error) {
	// POSTデータをバインド
	var request RequestRegisterTodoAsync
	if err := ctx.ShouldBindJSON(&request); err != nil {
		// 入力チェックエラーのハンドリング
		return nil, errors.NewValidationErrorWithCause(err, message.W_EX_5001)
	}
	todoTitles := request.TodoTitles

	// クエリパラメータfifoの取得
	fifo := ctx.Query("fifo")
	c.logger.Debug("fifo=%s", fifo)
	// クエリパラメータdbtxの取得
	dbtx := ctx.Query("dbtx")
	c.logger.Debug("dbtx=%s", dbtx)

	var serviceFunc domain.ServiceFunc
	if fifo == "" {
		serviceFunc = func() (any, error) {
			return nil, c.service.RegisterTodosAsync(todoTitles, dbtx)
		}
	} else {
		serviceFunc = func() (any, error) {
			return nil, c.service.RegisterTodosAsyncByFIFO(todoTitles, dbtx)
		}
	}
	// トランザクション管理してサービス実行
	_, err := c.transactionManager.ExecuteTransaction(serviceFunc)
	if err != nil {
		var bizErrs *errors.BusinessErrors
		// 業務エラーの場合にハンドリングしたい場合は、BusinessErrorsのみAsで判定すればよい
		// BusinessError(単一の業務エラー)の場合もBusinessErrorsとして判定できるようになっている
		if errors.As(err, &bizErrs) {
			// 付加情報が付与できる
			bizErrs.WithInfo("label1")
		} else if transaction.IsTransactionConditionalCheckFailed(err) {
			// 登録失敗の業務エラーにするか、スキップするかはケースバイケース
			return nil, errors.NewBusinessErrorWithCause(err, message.W_EX_8005)
		} else if transaction.IsTransactionConflict(err) {
			// 登録失敗の業務エラーにするか、システムエラーにするかはケースバイケース
			return nil, errors.NewSystemError(err, message.E_EX_9006)
		}
		/* 2つの理由コードが混在するケースでも業務エラーにする配慮する場合はこちらを使用
		} else if transaction.IsTransactionConditionalCheckFailedOrTransactionConflict(err) {
			// 登録失敗の業務エラーにするか、システムエラーにするかはケースバイケース
			return nil, errors.NewBusinessErrorWithCause(err, message.W_EX_8005)
		}*/
		return nil, err
	}

	return &ResponseRegisterTodoAsync{Result: "ok"}, nil
}

// FindBooksByCriteria implements BffController.
func (c *bffControllerImpl) FindBooksByCriteria(ctx *gin.Context) (any, error) {
	var request RequestFindBook
	if err := ctx.ShouldBindQuery(&request); err != nil {
		return nil, errors.NewValidationErrorWithCause(err, message.W_EX_5001)
	}
	bookCriteria := &repository.BookCriteria{
		Title:     request.Title,
		Author:    request.Author,
		Publisher: request.Publisher,
	}
	return c.service.FindBooksByCriteria(bookCriteria)
}

// RegisterBook implements BffController.
func (c *bffControllerImpl) RegisterBook(ctx *gin.Context) (any, error) {
	var request RequestRegisterBook
	if err := ctx.ShouldBindJSON(&request); err != nil {
		return nil, errors.NewValidationErrorWithCause(err, message.W_EX_5001)
	}
	book := &model.Book{
		Title:         request.Title,
		Author:        request.Author,
		Publisher:     request.Publisher,
		PublishedDate: request.PublishedDate,
		ISBN:          request.ISBN,
	}
	return c.service.RegisterBook(book)
}
