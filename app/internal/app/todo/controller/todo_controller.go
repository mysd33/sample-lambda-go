package controller

import (
	"app/internal/app/todo/service"

	"example.com/appbase/pkg/dynamodb"
	"example.com/appbase/pkg/logging"
	"github.com/gin-gonic/gin"
)

// リクエストデータ
type Request struct {
	TodoTitle string `json:"todo_title"`
}

type TodoController interface {
	Find(ctx *gin.Context) (interface{}, error)
	Regist(ctx *gin.Context) (interface{}, error)
}

func New(log logging.Logger, service *service.TodoService) TodoController {
	return &TodoControllerImpl{log: log, service: service}
}

type TodoControllerImpl struct {
	log     logging.Logger
	service *service.TodoService
}

func (c *TodoControllerImpl) Find(ctx *gin.Context) (interface{}, error) {
	// パスパラメータの取得
	todoId := ctx.Param("todo_id")
	// TODO: 入力チェック

	// DynamoDBトランザクション管理してサービスの実行
	return dynamodb.HandleTransaction(func() (interface{}, error) {
		return (*c.service).Find(todoId)
	})
}

func (c *TodoControllerImpl) Regist(ctx *gin.Context) (interface{}, error) {
	// POSTデータをバインド
	var request Request
	ctx.BindJSON(&request)
	// TODO: 入力チェック

	// DynamoDBトランザクション管理してサービスの実行
	return dynamodb.HandleTransaction(func() (interface{}, error) {
		return (*c.service).Regist(request.TodoTitle)
	})
}
