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

// Request は、REST APIで受け取るリクエストデータの構造体です。
type Request struct {
	Name string `json:"user_name" binding:"required"`
}

// UserController は、ユーザ管理業務のContollerインタフェースです。
type UserController interface {
	// Find は、パスパラメータで指定されたuser_idのユーザ情報を照会します。
	Find(ctx *gin.Context) (any, error)
	// Register は、リクエストデータで受け取ったユーザ情報を登録します。
	Register(ctx *gin.Context) (any, error)
}

// New は、UserControllerを作成します。
func New(log logging.Logger, transactionManager rdb.TransactionManager, service service.UserService) UserController {
	return &userControllerImpl{log: log, service: service, transactionManager: transactionManager}
}

// userControllerImpl は、UserControllerを実装する構造体です。
type userControllerImpl struct {
	log                logging.Logger
	service            service.UserService
	transactionManager rdb.TransactionManager
}

func (c *userControllerImpl) Find(ctx *gin.Context) (any, error) {
	// パスパラメータの取得
	userId := ctx.Param("user_id")
	// 入力チェック
	if userId == "" {
		// 入力チェックエラーのハンドリング
		return nil, errors.NewValidationError(message.W_EX_5002, "user_id")
	}

	// RDBトランザクション開始してサービスの実行
	return c.transactionManager.ExecuteTransaction(func() (any, error) {
		return c.service.Find(userId)
	})
}

func (c *userControllerImpl) Register(ctx *gin.Context) (any, error) {
	// POSTデータをバインド
	var request Request
	if err := ctx.ShouldBindJSON(&request); err != nil {
		// 入力チェックエラーのハンドリング
		return nil, errors.NewValidationErrorWithCause(err, message.W_EX_5001)
	}

	// RDBトランザクション開始してサービスの実行
	return c.transactionManager.ExecuteTransaction(func() (any, error) {
		return c.service.Register(request.Name)
	})
}
