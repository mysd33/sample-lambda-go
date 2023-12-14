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
	ReturnResponseBody(ctx *gin.Context)
}

// NewApiResponseFormatter は、ApiResponseFormatterを作成します。
func NewApiResponseFormatter(messageSource message.MessageSource) ApiResponseFormatter {
	return &defaultApiResponseFormatter{messageSource: messageSource}
}

type defaultApiResponseFormatter struct {
	messageSource message.MessageSource
}

// ReturnResponseBody implements ApiResponseFormatter.
func (f *defaultApiResponseFormatter) ReturnResponseBody(ctx *gin.Context) {
	var (
		validationError *myerrors.ValidationError
		businessError   *myerrors.BusinessError
		systemError     *myerrors.SystemError
	)
	errs := ctx.Errors

	if len(errs) > 0 {
		//TODO: エラーが複数の場合の処理
		err := errs[0]
		//TODO: ErrorTypeの判定やGin側でのエラーのハンドリングが必要？

		// 各エラー内容に応じた応答メッセージの成形
		if errors.As(err, &validationError) {
			ctx.JSON(http.StatusBadRequest, ErrorResponseBody("validationError", validationError.Error()))
		} else if errors.As(err, &businessError) {
			ctx.JSON(http.StatusBadRequest, ErrorResponseBody(businessError.ErrorCode(), f.messageSource.GetMessage(businessError.ErrorCode(), businessError.Args()...)))
		} else if errors.As(err, &systemError) {
			ctx.JSON(http.StatusInternalServerError, ErrorResponseBody(systemError.ErrorCode(), f.messageSource.GetMessage(message.E_FW_9001)))
		} else {
			ctx.JSON(http.StatusInternalServerError, ErrorResponseBody(message.E_FW_9999, f.messageSource.GetMessage(message.E_FW_9999)))
		}
	} else {
		//TODO: 定数化
		result, ok := ctx.Get("result")
		if ok {
			ctx.JSON(http.StatusOK, result)
		}
		//TODO: resultがない場合
	}
}

func ErrorResponseBody(label string, detail string) gin.H {
	//TODO: 要件に応じてエラーレスポンスの形式を修正する。
	return gin.H{"code": label, "detail": detail}
}
