package controller

import (
	"app/internal/app/user/service"

	"example.com/appbase/pkg/logging"
	"example.com/appbase/pkg/rdb"
	"github.com/gin-gonic/gin"
)

// リクエストデータ
type Request struct {
	Name string `json:"user_name"`
}

type UserController interface {
	Find(ctx *gin.Context) (interface{}, error)
	Regist(ctx *gin.Context) (interface{}, error)
}

func New(log logging.Logger, service *service.UserService) UserController {
	return &UserControllerImpl{log: log, service: service}
}

type UserControllerImpl struct {
	log     logging.Logger
	service *service.UserService
}

func (c *UserControllerImpl) Find(ctx *gin.Context) (interface{}, error) {
	// パスパラメータの取得
	userId := ctx.Param("user_id")
	// TODO: 入力チェック

	// RDBトランザクション開始してサービスの実行
	return rdb.HandleTransaction(func() (interface{}, error) {
		return (*c.service).Find(userId)
	})
}

func (c *UserControllerImpl) Regist(ctx *gin.Context) (interface{}, error) {
	// POSTデータをバインド
	var request Request
	ctx.BindJSON(&request)
	// TODO: 入力チェック

	// RDBトランザクション開始してサービスの実行
	return rdb.HandleTransaction(func() (interface{}, error) {
		return (*c.service).Regist(request.Name)
	})
}
