// controllerのパッケージ
package controller

import (
	"app/internal/app/todo/service"

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
	// Regist は、リクエストデータで受け取ったTodoを登録します。
	Regist(ctx *gin.Context) (interface{}, error)
}

// New は、TodoControllerを作成します。
func New(log logging.Logger, service service.TodoService) TodoController {
	return &todoControllerImpl{log: log, service: service}
}

// todoControllerImpl は、TodoControllerを実装する構造体です。
type todoControllerImpl struct {
	log     logging.Logger
	service service.TodoService
}

func (c *todoControllerImpl) Find(ctx *gin.Context) (interface{}, error) {
	// パスパラメータの取得
	todoId := ctx.Param("todo_id")
	// TODO: 入力チェック

	// DynamoDBトランザクション管理してサービスの実行
	return dynamodb.ExecuteTransaction(func() (interface{}, error) {
		return c.service.Find(todoId)
	})
}

func (c *todoControllerImpl) Regist(ctx *gin.Context) (interface{}, error) {
	// POSTデータをバインド
	var request Request
	if err := ctx.ShouldBindJSON(&request); err != nil {
		// 入力チェックエラーのハンドリング
		return nil, errors.NewValidationError(err)
	}

	// DynamoDBトランザクション管理してサービスの実行
	return dynamodb.ExecuteTransaction(func() (interface{}, error) {
		return c.service.Regist(request.TodoTitle)
	})
}
