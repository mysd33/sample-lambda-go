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
	Name string `label:"名前" json:"name" binding:"required"`
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
		// 入力チェックエラーのハンドリング（Causeなしの場合）
		return nil, errors.NewValidationError(message.W_EX_5002, "△△")
	}
	if errorType == "validation2" {
		var request Request
		if err := ctx.ShouldBindJSON(&request); err != nil {
			// 入力チェックエラーのハンドリング（Causeありの場合）
			return nil, errors.NewValidationErrorWithCause(err, message.W_EX_5001)
		}
	}
	if errorType == "validation3" {
		// 入力チェックエラーのハンドリング(Causeがnilの場合)
		return nil, errors.NewValidationErrorWithCause(nil, message.W_EX_5001)
	}

	err := c.service.Execute(errorType)
	if err != nil {
		// 業務エラーの場合にハンドリングしたい場合は、BusinessErrorsのみAsで判定すればよい
		// BusinessError(単一の業務エラー)の場合もBusinessErrorsとして判定できるようになっている
		var bizErrs *errors.BusinessErrors
		if errors.As(err, &bizErrs) {
			// 付加情報が付与できる
			bizErrs.WithInfo("label1")
		}

		return nil, err
	}
	return nil, nil
}
