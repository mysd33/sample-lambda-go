// controllerのパッケージ
package controller

import (
	"app/internal/app/bff/service"
	"app/internal/pkg/message"
	"app/internal/pkg/model"

	"example.com/appbase/pkg/domain"
	"example.com/appbase/pkg/errors"
	"example.com/appbase/pkg/logging"
	"example.com/appbase/pkg/transaction"
	"github.com/gin-gonic/gin"
)

// RequestRegisterUser は、REST APIで受け取るリクエストデータの構造体です。
type RequestRegisterUser struct {
	Name string `json:"user_name" binding:"required"`
}

// RequestRegisterTodo は、REST APIで受け取るリクエストデータの構造体です。
type RequestRegisterTodo struct {
	// TodoTitle は、Todoのタイトルです。
	TodoTitle string `json:"todo_title" binding:"required"`
}

// ResponseFindTodo は、REST APIで受け取るレスポンスデータの構造体です。
type ResponseFindTodo struct {
	User *model.User `json:"user"`
	Todo *model.Todo `json:"todo"`
}

type RequestRegisterTodoAsync struct {
	TodoTitles []string `json:"todo_titles" binding:"required"`
}

type ResponseRegisterTodoAsync struct {
	Result string `json:"result"`
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
	userId := ctx.Query("user_id")
	// 入力チェック
	if userId == "" {
		// 入力チェックエラーのハンドリング
		return nil, errors.NewValidationError(message.W_EX_5002, "user_id")
	}
	todoId := ctx.Query("todo_id")
	if todoId == "" {
		// 入力チェックエラーのハンドリング
		return nil, errors.NewValidationError(message.W_EX_5002, "todo_id")
	}
	// サービスの実行（DynamoDBトランザクション管理なし）
	user, todo, err := c.service.FindTodo(userId, todoId)
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
			// 登録失敗の業務エラーにするか、スキップするかはケースバイケース
			return nil, errors.NewBusinessErrorWithCause(err, message.W_EX_8005)
		}
		return nil, err
	}

	return &ResponseRegisterTodoAsync{Result: "ok"}, nil
}
