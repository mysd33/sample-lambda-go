/*
handler パッケージは、Lambdaのハンドラメソッドに関する機能を提供するパッケージです。
*/
package handler

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// GetDefaultGinEngine は、ginのEngineを取得します。
func GetDefaultGinEngine() *gin.Engine {
	r := gin.Default()
	// 404エラー
	r.NoRoute(func(ctx *gin.Context) {
		ctx.JSON(404, gin.H{"code": "NOT_FOUND", "detail": fmt.Sprintf("%s is not found", ctx.Request.URL.Path)})
	})
	// 405エラー
	r.HandleMethodNotAllowed = true
	r.NoMethod(func(ctx *gin.Context) {
		ctx.JSON(405, gin.H{"code": "METHOD_NOT_ALLOWED", "detail": fmt.Sprintf("%s Method %s is not allowed", ctx.Request.Method, ctx.Request.URL.Path)})
	})
	return r
}
