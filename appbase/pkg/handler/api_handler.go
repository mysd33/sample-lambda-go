/*
handler パッケージは、Lambdaのハンドラメソッドに関する機能を提供するパッケージです。
*/
package handler

import (
	"context"

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
	engine.Use(
		gin.Logger(),
		func(ctx *gin.Context) {
			ctx.Next()
			// レスポンスの生成
			h.apiResponseFormatter.ReturnResponseBody(ctx, errorResponse)
		},
		// パニック時のカスタムリカバリ処理
		gin.CustomRecovery(func(c *gin.Context, recover any) {
			// パニックをエラーでラップ
			err := errors.Errorf("recover from: %+v", recover)
			h.log.ErrorWithUnexpectedError(err)
			// エラーをその他のエラー（ginのエラーログ対象外）として、ginのContextに格納
			c.Error(err).SetType(gin.ErrorTypeNu)
		}))

	// 404エラー
	engine.NoRoute(func(ctx *gin.Context) {
		h.log.Debug("%s is not found", ctx.Request.URL.Path)
		// エラーをPublicなエラー（ginのエラーログ対象外）として、ginのContextに格納
		ctx.Error(api.NoRouteError).SetType(gin.ErrorTypePublic)
	})
	// 405エラー
	engine.HandleMethodNotAllowed = true
	engine.NoMethod(func(ctx *gin.Context) {
		h.log.Debug("%s Method %s is not allowed", ctx.Request.Method, ctx.Request.URL.Path)
		// エラーをPublicなエラー（ginのエラーログ対象外）として、ginのContextに格納
		ctx.Error(api.NoMethodError).SetType(gin.ErrorTypePublic)
	})
	return engine
}

// Handleは、APIGatewayトリガのLambdaのハンドラを実行します。
func (h *APILambdaHandler) Handle(ginLambda *ginadapter.GinLambda, errorResponse api.ErrorResponse) APITriggeredLambdaHandlerFunc {
	// Handleは、APIGatewayトリガーのLambdaHandlerFuncです。
	return func(ctx context.Context, request events.APIGatewayProxyRequest) (response events.APIGatewayProxyResponse, err error) {
		defer func() {
			// パニックのリカバリ処理
			if v := recover(); v != nil {
				perr := errors.Errorf("recover from: %+v", v)
				// パニックのスタックトレース情報をログ出力
				h.log.ErrorWithUnexpectedError(perr)
				// エラーレスポンスの生成
				// レスポンスの生成でエラーだと、{"message":"Internal Server Error"} のレスポンスが返却される
				response, err = h.apiResponseFormatter.CreateAPIGatewayProxyResponseForUnexpectedError(perr, errorResponse)
			}
			// ログのフラッシュ
			h.log.Sync()
		}()
		// ctxをコンテキスト領域に格納
		apcontext.Context = ctx

		// リクエストIDをログの付加情報として追加
		h.log.ClearInfo()
		lc := apcontext.GetLambdaContext(ctx)
		h.log.AddInfo("AWS RequestID", lc.AwsRequestID)
		h.log.AddInfo("API Gateway RequestID", request.RequestContext.RequestID)

		// AWS Lambda Go API Proxyでginと統合
		// https://github.com/awslabs/aws-lambda-go-api-proxy
		response, gerr := ginLambda.ProxyWithContext(ctx, request)
		if gerr != nil {
			h.log.ErrorWithUnexpectedError(gerr)
			// エラーレスポンスの生成
			// レスポンスの生成でエラーだと、{"message":"Internal Server Error"} のレスポンスが返却される
			response, err = h.apiResponseFormatter.CreateAPIGatewayProxyResponseForUnexpectedError(gerr, errorResponse)
		}
		return
	}
}
