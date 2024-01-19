/*
api パッケージは、REST APIに関する機能を提供するパッケージです。
*/
package api

import (
	"net/http"

	"example.com/appbase/pkg/constant"
	myerrors "example.com/appbase/pkg/errors"
	"example.com/appbase/pkg/logging"
	"example.com/appbase/pkg/message"
	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
)

// ApiResponseFormatterは、レスポンスデータを作成するインタフェース
type ApiResponseFormatter interface {
	// ReturnResponseBody は、処理結果resultまたはエラーerrに対応するレスポンスボディを返却します。
	ReturnResponseBody(ctx *gin.Context, errorResponse ErrorResponse)
}

// NewApiResponseFormatter は、ApiResponseFormatterを作成します。
func NewApiResponseFormatter(log logging.Logger, messageSource message.MessageSource) ApiResponseFormatter {
	return &defaultApiResponseFormatter{log: log, messageSource: messageSource}
}

type defaultApiResponseFormatter struct {
	log           logging.Logger
	messageSource message.MessageSource
}

// ReturnResponseBody implements ApiResponseFormatter.
func (f *defaultApiResponseFormatter) ReturnResponseBody(ctx *gin.Context, errorResponse ErrorResponse) {
	var (
		validationError *myerrors.ValidationError
		businessErrors  *myerrors.BusinessErrors
		systemError     *myerrors.SystemError
	)
	errs := ctx.Errors

	if len(errs) > 0 {
		//TODO: エラーが複数の場合の処理
		err := errs[0]
		//TODO: ErrorTypeの判定やGin側でのエラーのハンドリングが必要？

		// 各エラー内容に応じた応答メッセージの成形
		if errors.As(err, &validationError) {
			ctx.JSON(errorResponse.ValidationErrorResponse(validationError))
		} else if errors.As(err, &businessErrors) {
			ctx.JSON(errorResponse.BusinessErrorResponse(businessErrors))
		} else if errors.As(err, &systemError) {
			ctx.JSON(errorResponse.SystemErrorResponse(systemError))
		} else {
			ctx.JSON(errorResponse.UnExpectedErrorResponse(err))
		}
	} else {
		result, ok := ctx.Get(constant.CONTROLLER_RESULT)
		if ok {
			ctx.JSON(http.StatusOK, result)
			return
		}
		// resultが取得できなかった場合にエラーログを出力
		f.log.Error(message.E_FW_9003)
	}
}
