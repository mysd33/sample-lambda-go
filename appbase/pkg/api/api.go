/*
api パッケージは、REST APIに関する機能を提供するパッケージです。
*/
package api

import (
	"errors"
	"net/http"

	myerrors "example.com/appbase/pkg/errors"
	"example.com/appbase/pkg/message"
	"github.com/gin-gonic/gin"
)

// ApiResponseFormatterは、レスポンスデータを作成するインタフェース
type ApiResponseFormatter interface {
	// ReturnResponseBody は、処理結果resultまたはエラーerrに対応するレスポンスボディを返却します。
	ReturnResponseBody(ctx *gin.Context, result any, err error)
}

// NewApiResponseFormatter は、ApiResponseFormatterを作成します。
func NewApiResponseFormatter(messageSource message.MessageSource) ApiResponseFormatter {
	return &defaultApiResponseFormatter{messageSource: messageSource}
}

type defaultApiResponseFormatter struct {
	messageSource message.MessageSource
}

// ReturnResponseBody implements ApiResponseFormatter.
func (f *defaultApiResponseFormatter) ReturnResponseBody(ctx *gin.Context, result any, err error) {
	var (
		validationError *myerrors.ValidationError
		businessError   *myerrors.BusinessError
		systemError     *myerrors.SystemError
	)
	if err != nil {
		// 各エラー内容に応じた応答メッセージの成形
		if errors.As(err, &validationError) {
			ctx.JSON(http.StatusBadRequest, errorResponseBody("validationError", validationError.Error()))
		} else if errors.As(err, &businessError) {
			ctx.JSON(http.StatusBadRequest, errorResponseBody(businessError.ErrorCode(), f.messageSource.GetMessage(businessError.ErrorCode(), businessError.Args()...)))
		} else if errors.As(err, &systemError) {
			ctx.JSON(http.StatusInternalServerError, errorResponseBody(systemError.ErrorCode(), f.messageSource.GetMessage(message.E_FW_9001)))
		} else {
			ctx.JSON(http.StatusInternalServerError, errorResponseBody(message.E_FW_9999, f.messageSource.GetMessage(message.E_FW_9999)))
		}
	} else {
		ctx.JSON(http.StatusOK, result)
	}
}

func errorResponseBody(label string, detail string) gin.H {
	return gin.H{"code": label, "detail": detail}
}
