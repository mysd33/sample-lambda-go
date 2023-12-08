// controllerのパッケージ
package controller

import (
	"app/internal/app/bff/service"
	"app/internal/pkg/entity"

	"example.com/appbase/pkg/errors"
	"example.com/appbase/pkg/logging"
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
	User *entity.User `json:"user"`
	Todo *entity.Todo `json:"todo"`
}

type RequestRegisterTodoAsync struct {
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
func New(log logging.Logger, service service.BffService) BffController {
	return &bffControllerImpl{log: log, service: service}
}

// bffControllerImpl は、BffControllerを実装する構造体です。
type bffControllerImpl struct {
	log     logging.Logger
	service service.BffService
}

// FindTodo implements BffController.
func (c *bffControllerImpl) FindTodo(ctx *gin.Context) (any, error) {
	// クエリパラメータの取得
	userId := ctx.Query("user_id")
	// 入力チェック
	if userId == "" {
		// 入力チェックエラーのハンドリング
		return nil, errors.NewValidationErrorWithMessage("クエリパラメータuserIdが未指定です")
	}
	todoId := ctx.Query("todo_id")
	if todoId == "" {
		// 入力チェックエラーのハンドリング
		return nil, errors.NewValidationErrorWithMessage("クエリパラメータtodoIdが未指定です")
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
		return nil, errors.NewValidationError(err)
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
		return nil, errors.NewValidationError(err)
	}

	// サービスの実行
	return c.service.RegisterTodo(request.TodoTitle)
}

// RegisterTodosAsync implements BffController.
func (c *bffControllerImpl) RegisterTodosAsync(ctx *gin.Context) (any, error) {
	// TODO: 入力情報の受付
	err := c.service.RegisterTodosAsync([]string{"dummy1", "dummy2"})
	if err != nil {
		return nil, err
	}
	// TODO: トランザクション管理して実行
	return &RequestRegisterTodoAsync{Result: "ok"}, nil
}
