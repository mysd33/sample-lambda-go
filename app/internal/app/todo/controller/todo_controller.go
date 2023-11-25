// controllerのパッケージ
package controller

import (
	"app/internal/app/todo/service"

	"example.com/appbase/pkg/domain"
	"example.com/appbase/pkg/dynamodb"
	"example.com/appbase/pkg/errors"
	"example.com/appbase/pkg/logging"
	"github.com/gin-gonic/gin"
)

// Request は、REST APIで受け取るリクエストデータの構造体です。
type Request struct {
	// TodoTitle は、Todoのタイトルです。
	TodoTitle string `json:"todo_title" binding:"required"`
}

// TodoController は、Todo業務のControllerインタフェースです。
type TodoController interface {
	// Find は、パスパラメータで指定されたtodo_idのTodoを照会します。
	Find(ctx *gin.Context) (interface{}, error)
	// Register は、リクエストデータで受け取ったTodoを登録します。
	Register(ctx *gin.Context) (interface{}, error)
}

// New は、TodoControllerを作成します。
func New(log logging.Logger,
	transactionManager dynamodb.TransactionManager,
	service service.TodoService,
) TodoController {
	return &todoControllerImpl{log: log, transactionManager: transactionManager, service: service}
}

// todoControllerImpl は、TodoControllerを実装する構造体です。
type todoControllerImpl struct {
	log                logging.Logger
	transactionManager dynamodb.TransactionManager
	service            service.TodoService
}

func (c *todoControllerImpl) Find(ctx *gin.Context) (interface{}, error) {
	// パスパラメータの取得
	todoId := ctx.Param("todo_id")
	// 入力チェック
	if todoId == "" {
		// 入力チェックエラーのハンドリング
		return nil, errors.NewValidationErrorWithMessage("クエリパラメータtodoIdが未指定です")
	}
	// DynamoDBトランザクション管理してサービスの実行
	return c.transactionManager.ExecuteTransaction(func() (interface{}, error) {
		return c.service.Find(todoId)
	})
}

func (c *todoControllerImpl) Register(ctx *gin.Context) (interface{}, error) {
	// POSTデータをバインド
	var request Request
	if err := ctx.ShouldBindJSON(&request); err != nil {
		// 入力チェックエラーのハンドリング
		return nil, errors.NewValidationError(err)
	}
	// クエリパラメータの取得
	transaction := ctx.Query("tx")
	var serviceFunc domain.ServiceFunc

	if transaction != "" {
		serviceFunc = func() (interface{}, error) {
			// トランザクション指定あり
			return c.service.RegisterTx(request.TodoTitle)
		}
	} else {
		serviceFunc = func() (interface{}, error) {
			return c.service.Register(request.TodoTitle)
		}
	}

	// DynamoDBトランザクション管理してサービスの実行
	return c.transactionManager.ExecuteTransaction(serviceFunc)
}
