/*
handler パッケージは、Lambdaのハンドラメソッドに関する機能を提供するパッケージです。
*/
package handler

import (
	"context"
	"fmt"
	"net/http"

	"example.com/appbase/pkg/apcontext"
	"example.com/appbase/pkg/api"
	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/logging"
	"example.com/appbase/pkg/message"
	"github.com/aws/aws-lambda-go/events"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
)

var (
	NoRouteError  = errors.New("NOT FOUND")
	NoMethodError = errors.New("METHOD NOT ALLOWED")
)

// APITriggeredLambdaHandlerFunc は、APIGatewayトリガのLambdaのハンドラを表す関数です。
type APITriggeredLambdaHandlerFunc func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)

// APILambdaHandler は、APIGatewayトリガのLambdaのハンドラを管理する構造体です。
type APILambdaHandler struct {
	config               config.Config
	log                  logging.Logger
	messageSource        message.MessageSource
	apiResponseFormatter api.ApiResponseFormatter
}

// NewAPILambdaHandler は、APILambdaHandlerを作成します。
func NewAPILambdaHandler(config config.Config,
	log logging.Logger,
	messageSource message.MessageSource,
	apiResponseFormatter api.ApiResponseFormatter) *APILambdaHandler {
	return &APILambdaHandler{
		config:               config,
		log:                  log,
		messageSource:        messageSource,
		apiResponseFormatter: apiResponseFormatter,
	}
}

// GetDefaultGinEngine は、ginのEngineを取得します。
func (h *APILambdaHandler) GetDefaultGinEngine(errorResponse api.ErrorResponse) *gin.Engine {
	// ginをLoggerとCustomerRecoverのミドルウェアがアタッチされた状態で作成
	engine := gin.New()
	engine.Use(gin.Logger(),
		func(ctx *gin.Context) {
			ctx.Next()
			// レスポンスの生成
			h.apiResponseFormatter.ReturnResponseBody(ctx, errorResponse)
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
		h.log.Debug("%s is not found", ctx.Request.URL.Path)
		ctx.JSON(errorResponse.WarnErrorResponse(NoRouteError))
	})
	// 405エラー
	engine.HandleMethodNotAllowed = true
	engine.NoMethod(func(ctx *gin.Context) {
		h.log.Debug("%s Method %s is not allowed", ctx.Request.Method, ctx.Request.URL.Path)
		ctx.JSON(errorResponse.WarnErrorResponse(NoMethodError))
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
				// パニックのスタックトレース情報をログ出力
				h.log.ErrorWithUnexpectedError(fmt.Errorf("recover from: %+v", v))

				response = h.createErrorResponse()
				// errはnilにままにして{"message":"Internal Server Error"} のレスポンスが返却されないようにする
			}
		}()
		// ctxをコンテキスト領域に格納
		apcontext.Context = ctx
		// AWS Lambda Go API Proxyでginと統合
		// https://github.com/awslabs/aws-lambda-go-api-proxy
		response, err = ginLambda.ProxyWithContext(ctx, request)
		if err != nil {
			h.log.ErrorWithUnexpectedError(err)
			response = h.createErrorResponse()
			// errはnilに戻して{"message":"Internal Server Error"} のレスポンスが返却されないようにする
			err = nil
		}
		return
	}
}

// TODO: エラーレスポンスの形式
func (h *APILambdaHandler) createErrorResponse() events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusInternalServerError,
		Body:       fmt.Sprintf("{\"code\": %s, \"detail\": %s}", message.E_FW_9999, h.messageSource.GetMessage(message.E_FW_9999)),
	}
}
