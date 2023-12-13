/*
handler パッケージは、Lambdaのハンドラメソッドに関する機能を提供するパッケージです。
*/
package handler

import (
	"context"
	"fmt"

	"example.com/appbase/pkg/apcontext"
	"example.com/appbase/pkg/api"
	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/logging"
	"github.com/aws/aws-lambda-go/events"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
)

// APITriggeredLambdaHandlerFunc は、APIGatewayトリガのLambdaのハンドラを表す関数です。
type APITriggeredLambdaHandlerFunc func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)

// APILambdaHandler は、APIGatewayトリガのLambdaのハンドラを管理する構造体です。
type APILambdaHandler struct {
	config               config.Config
	log                  logging.Logger
	apiResponseFormatter api.ApiResponseFormatter
}

// NewAPILambdaHandler は、APILambdaHandlerを作成します。
func NewAPILambdaHandler(config config.Config,
	log logging.Logger,
	apiResponseFormatter api.ApiResponseFormatter) *APILambdaHandler {
	return &APILambdaHandler{
		config:               config,
		log:                  log,
		apiResponseFormatter: apiResponseFormatter,
	}
}

// GetDefaultGinEngine は、ginのEngineを取得します。
func (h *APILambdaHandler) GetDefaultGinEngine() *gin.Engine {
	// ginをLoggerとCustomerRecoverのミドルウェアがアタッチされた状態で作成
	engine := gin.New()
	engine.Use(gin.Logger(),
		func(ctx *gin.Context) {
			ctx.Next()
			// レスポンスの生成
			h.apiResponseFormatter.ReturnResponseBody(ctx)
		},
		// パニック時のカスタムリカバリ処理
		gin.CustomRecovery(func(c *gin.Context, recover any) {
			err := fmt.Errorf("recover from: %+v", recover)
			h.log.ErrorWithUnexpectedError(err)
			// エラーをContextに格納
			c.Error(err)
		}))

	// 404エラー
	engine.NoRoute(func(ctx *gin.Context) {
		h.log.Debug("404エラー")
		ctx.JSON(404, gin.H{"code": "NOT_FOUND", "detail": fmt.Sprintf("%s is not found", ctx.Request.URL.Path)})
	})
	// 405エラー
	engine.HandleMethodNotAllowed = true
	engine.NoMethod(func(ctx *gin.Context) {
		h.log.Debug("405エラー")
		ctx.JSON(405, gin.H{"code": "METHOD_NOT_ALLOWED", "detail": fmt.Sprintf("%s Method %s is not allowed", ctx.Request.Method, ctx.Request.URL.Path)})
	})
	return engine
}

// Handleは、APIGatewayトリガのLambdaのハンドラを実行します。
func (h *APILambdaHandler) Handle(ginLambda *ginadapter.GinLambda) APITriggeredLambdaHandlerFunc {
	// Handleは、APIGatewayトリガーのLambdaHandlerFuncです。
	return func(ctx context.Context, request events.APIGatewayProxyRequest) (response events.APIGatewayProxyResponse, err error) {
		// パニックのリカバリ処理
		defer func() {
			if v := recover(); v != nil {
				err = fmt.Errorf("recover from: %+v", v)
				h.log.ErrorWithUnexpectedError(err)
			}
		}()
		// ctxをコンテキスト領域に格納
		apcontext.Context = ctx
		// AWS Lambda Go API Proxyでginと統合
		// https://github.com/awslabs/aws-lambda-go-api-proxy
		response, err = ginLambda.ProxyWithContext(ctx, request)
		return
	}
}
