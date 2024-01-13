// controllerのパッケージ
package controller

import (
	"app/internal/app/errortest/service"
	"app/internal/pkg/message"

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
	Execute(ctx *gin.Context) (any, error)
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
func (c *errorTestControllerImpl) Execute(ctx *gin.Context) (any, error) {
	errorType := ctx.Param("errortype")

	if errorType == "validation" {
		// 入力チェックエラーのハンドリング
		return nil, errors.NewValidationError(message.W_EX_5002, "△△")
	}
	if errorType == "validation2" {
		var request Request
		if err := ctx.ShouldBindJSON(&request); err != nil {
			// 入力チェックエラーのハンドリング
			return nil, errors.NewValidationErrorWithCause(err, message.W_EX_5001)
		}
	}

	return nil, c.service.Execute(errorType)
}
