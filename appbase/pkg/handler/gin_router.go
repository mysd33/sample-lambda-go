/*
handler パッケージは、Lambdaのハンドラメソッドに関する機能を提供するパッケージです。
*/
package handler

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func GetDefaultGinEngine() *gin.Engine {
	r := gin.Default()
	// 404エラー
	r.NoRoute(func(ctx *gin.Context) {
		ctx.JSON(404, gin.H{"code": "NOT_FOUND", "detail": fmt.Sprintf("%s is not found", ctx.Request.URL.RawPath)})
	})
	return r
}
