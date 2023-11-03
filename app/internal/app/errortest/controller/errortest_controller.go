// controllerのパッケージ
package controller

import (
	"app/internal/app/errortest/service"

	"example.com/appbase/pkg/errors"
	"example.com/appbase/pkg/logging"
	"github.com/gin-gonic/gin"
)

// Request はダミーのリクエスト
type Request struct {
	Name string `json:"name" binding:"required"`
}

// ErrorController は、エラーをテストするためのControllerインタフェースです。
type ErrorTestController interface {
	Execute(ctx *gin.Context) (interface{}, error)
}

// New は、ErrorControllerを作成します。
func New(log logging.Logger, service service.ErrorTestService) ErrorTestController {
	return &errorTestControllerImpl{log: log, service: service}
}

// errorTestControllerImpl は、ErrorTestControllerを実装します。
type errorTestControllerImpl struct {
	log     logging.Logger
	service service.ErrorTestService
}

// Execute implements ErrorTestController.
func (c *errorTestControllerImpl) Execute(ctx *gin.Context) (interface{}, error) {
	errorType := ctx.Param("errortype")

	if errorType == "validation" {
		// 入力チェックエラーのハンドリング
		return nil, errors.NewValidationErrorWithMessage("△△は必須入力です")
	}
	if errorType == "validation2" {
		var request Request
		if err := ctx.ShouldBindJSON(&request); err != nil {
			// 入力チェックエラーのハンドリング
			return nil, errors.NewValidationError(err)
		}
	}

	return nil, c.service.Execute(errorType)
}
