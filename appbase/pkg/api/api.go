/*
api パッケージは、REST APIに関する機能を提供するパッケージです。
*/
package api

import (
	"errors"
	"net/http"

	myerrors "example.com/appbase/pkg/errors"
	"github.com/gin-gonic/gin"
)

// ReturnResponseBody は、処理結果resultまたはエラーerrに対応するレスポンスボディを返却します。
func ReturnResponseBody(ctx *gin.Context, result interface{}, err error) {
	var (
		validationError *myerrors.ValidationError
		businessError   *myerrors.BusinessError
		systemError     *myerrors.SystemError
	)
	if err != nil {
		// TODO: 各エラー内容に応じた応答メッセージの成形
		if errors.As(err, &validationError) {
			ctx.JSON(http.StatusBadRequest, errorResponseBody("validationError", validationError.Cause.Error()))
		} else if errors.As(err, &businessError) {
			ctx.JSON(http.StatusBadRequest, errorResponseBody(businessError.ErrorCode, businessError.Cause.Error()))
		} else if errors.As(err, &systemError) {
			ctx.JSON(http.StatusInternalServerError, errorResponseBody(systemError.ErrorCode, "システムエラーが発生しました"))
		} else {
			ctx.JSON(http.StatusInternalServerError, errorResponseBody("e.ex.9999", "システムエラーが発生しました"))
		}
	} else {
		ctx.JSON(http.StatusOK, result)
	}
}

func errorResponseBody(label string, detail string) gin.H {
	return gin.H{"code": label, "detail": detail}
}
