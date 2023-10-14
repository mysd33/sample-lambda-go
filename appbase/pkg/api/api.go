/*
api パッケージは、REST APIに関する機能を提供するパッケージです。
*/
package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ReturnResponseBody は、処理結果resultまたはエラーerrに対応するレスポンスボディを返却します。
func ReturnResponseBody(ctx *gin.Context, result interface{}, err error) {
	if err != nil {
		//TODO: エラー時の応答メッセージ
		ctx.JSON(http.StatusBadRequest, errorResponseBody(err.Error()))
	}
	ctx.JSON(http.StatusOK, result)
}

func errorResponseBody(msg string) string {
	return fmt.Sprintf("{\"message\":\"%s\"}", msg)
}
