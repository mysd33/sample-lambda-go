// controllerのパッケージ
package controller

import (
	"app/internal/app/user/service"

	myerrors "example.com/appbase/pkg/errors"
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
	return &UserControllerImpl{log: log, service: service}
}

// UserControllerImpl は、UserControllerを実装する構造体です。
type UserControllerImpl struct {
	log     logging.Logger
	service service.UserService
}

func (c *UserControllerImpl) Find(ctx *gin.Context) (interface{}, error) {
	// パスパラメータの取得
	userId := ctx.Param("user_id")
	// TODO: 入力チェック

	// RDBトランザクション開始してサービスの実行
	return rdb.ExecuteTransaction(func() (interface{}, error) {
		return c.service.Find(userId)
	})
}

func (c *UserControllerImpl) Regist(ctx *gin.Context) (interface{}, error) {
	// POSTデータをバインド
	var request Request
	if err := ctx.ShouldBindJSON(&request); err != nil {
		// 入力チェックエラーのハンドリング
		return nil, myerrors.NewValidationError(err)
	}

	// RDBトランザクション開始してサービスの実行
	return rdb.ExecuteTransaction(func() (interface{}, error) {
		return c.service.Regist(request.Name)
	})
}
