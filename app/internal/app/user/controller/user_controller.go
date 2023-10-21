// controllerのパッケージ
package controller

import (
	"app/internal/app/user/service"

	"example.com/appbase/pkg/errors"
	"example.com/appbase/pkg/logging"
	"example.com/appbase/pkg/rdb"
	"github.com/gin-gonic/gin"
)

// Request は、REST APIで受け取るリクエストデータの構造体です。
type Request struct {
	Name string `json:"user_name" binding:"required"`
}

// UserController は、ユーザ管理業務のContollerインタフェースです。
type UserController interface {
	// Find は、パスパラメータで指定されたuser_idのユーザ情報を照会します。
	Find(ctx *gin.Context) (interface{}, error)
	// Regist は、リクエストデータで受け取ったユーザ情報を登録します。
	Regist(ctx *gin.Context) (interface{}, error)
}

// New は、UserControllerを作成します。
func New(log logging.Logger, service service.UserService) UserController {
	return &userControllerImpl{log: log, service: service}
}

// userControllerImpl は、UserControllerを実装する構造体です。
type userControllerImpl struct {
	log     logging.Logger
	service service.UserService
}

func (c *userControllerImpl) Find(ctx *gin.Context) (interface{}, error) {
	// パスパラメータの取得
	userId := ctx.Param("user_id")
	// 入力チェック
	if userId == "" {
		// 入力チェックエラーのハンドリング
		return nil, errors.NewValidationErrorWithMessage("クエリパラメータuserIdが未指定です")
	}

	// RDBトランザクション開始してサービスの実行
	return rdb.ExecuteTransaction(func() (interface{}, error) {
		return c.service.Find(userId)
	})
}

func (c *userControllerImpl) Regist(ctx *gin.Context) (interface{}, error) {
	// POSTデータをバインド
	var request Request
	if err := ctx.ShouldBindJSON(&request); err != nil {
		// 入力チェックエラーのハンドリング
		return nil, errors.NewValidationError(err)
	}

	// RDBトランザクション開始してサービスの実行
	return rdb.ExecuteTransaction(func() (interface{}, error) {
		return c.service.Regist(request.Name)
	})
}
