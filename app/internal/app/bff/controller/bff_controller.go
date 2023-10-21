// controllerのパッケージ
package controller

import (
	"app/internal/app/bff/service"
	"app/internal/pkg/entity"

	"example.com/appbase/pkg/dynamodb"
	"example.com/appbase/pkg/errors"
	"example.com/appbase/pkg/logging"
	"example.com/appbase/pkg/rdb"
	"github.com/gin-gonic/gin"
)

// Response は、REST APIで受け取るレスポンスデータの構造体です。
type Response struct {
	User *entity.User `json:"user"`
	Todo *entity.Todo `json:"todo"`
}

// BffController は、Bff業務のControllerインタフェースです。
type BffController interface {
	// Find は、クエリパラメータで指定されたtodo_idとuser_idのTodoを照会します。
	Find(ctx *gin.Context) (interface{}, error)
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

// Find implements BffController.
func (c *bffControllerImpl) Find(ctx *gin.Context) (interface{}, error) {
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
	// TODO: トランザクションは削除予定
	return rdb.ExecuteTransaction(func() (interface{}, error) {
		return dynamodb.ExecuteTransaction(func() (interface{}, error) {
			user, todo, err := c.service.Find(userId, todoId)
			if err != nil {
				return nil, err
			}
			return &Response{User: user, Todo: todo}, nil
		})
	})

}
