// controllerのパッケージ
package controller

import (
	"app/internal/app/user/service"
	"app/internal/pkg/message"

	"example.com/appbase/pkg/errors"
	"example.com/appbase/pkg/logging"
	"example.com/appbase/pkg/rdb"
	"github.com/gin-gonic/gin"
)

// RequestFind は、ユーザ照会のREST APIで受け取るリクエストデータの構造体です。
type RequestFind struct {
	// UserId は、ユーザのIDです。
	UserId string `label:"ユーザID(user_id)" uri:"user_id" binding:"required,uuid"`
}

// RequestRegister は、ユーザ登録のREST APIで受け取るリクエストデータの構造体です。
type RequestRegister struct {
	// Name は、ユーザの名前です。
	Name string `label:"ユーザ名(user_name)" json:"user_name" binding:"required"`
}

// UserController は、ユーザ管理業務のContollerインタフェースです。
type UserController interface {
	// Find は、パスパラメータで指定されたuser_idのユーザ情報を照会します。
	Find(ctx *gin.Context) (any, error)
	// Register は、リクエストデータで受け取ったユーザ情報を登録します。
	Register(ctx *gin.Context) (any, error)
}

// New は、UserControllerを作成します。
func New(logger logging.Logger, transactionManager rdb.TransactionManager, service service.UserService) UserController {
	return &userControllerImpl{logger: logger, service: service, transactionManager: transactionManager}
}

// userControllerImpl は、UserControllerを実装する構造体です。
type userControllerImpl struct {
	logger             logging.Logger
	service            service.UserService
	transactionManager rdb.TransactionManager
}

func (c *userControllerImpl) Find(ctx *gin.Context) (any, error) {
	// パスパラメータの取得
	var request RequestFind
	if err := ctx.ShouldBindUri(&request); err != nil {
		// 入力チェックエラーのハンドリング
		return nil, errors.NewValidationErrorWithCause(err, message.W_EX_5001)
	}

	// RDBトランザクション開始してサービスの実行
	return c.transactionManager.ExecuteTransaction(func() (any, error) {
		return c.service.Find(request.UserId)
	})
}

func (c *userControllerImpl) Register(ctx *gin.Context) (any, error) {
	// POSTデータをバインド
	var request RequestRegister
	if err := ctx.ShouldBindJSON(&request); err != nil {
		// 入力チェックエラーのハンドリング
		return nil, errors.NewValidationErrorWithCause(err, message.W_EX_5001)
	}

	// RDBトランザクション開始してサービスの実行
	return c.transactionManager.ExecuteTransaction(func() (any, error) {
		return c.service.Register(request.Name)
	})
}
