/*
handler パッケージは、Lambdaのハンドラメソッドに関する機能を提供するパッケージです。
*/
package handler

import (
	"context"
	"encoding/json"

	"example.com/appbase/pkg/apcontext"
	"example.com/appbase/pkg/api"
	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/logging"
	"example.com/appbase/pkg/message"
	"github.com/aws/aws-lambda-go/events"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/cockroachdb/errors"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// APITriggeredLambdaHandlerFunc は、APIGatewayトリガのLambdaのハンドラを表す関数です。
type APITriggeredLambdaHandlerFunc func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)

// APILambdaHandler は、APIGatewayトリガのLambdaのハンドラを管理する構造体です。
type APILambdaHandler struct {
	config               config.Config
	logger               logging.Logger
	messageSource        message.MessageSource
	apiResponseFormatter api.ApiResponseFormatter
}

// NewAPILambdaHandler は、APILambdaHandlerを作成します。
func NewAPILambdaHandler(config config.Config,
	logger logging.Logger,
	messageSource message.MessageSource,
	apiResponseFormatter api.ApiResponseFormatter) *APILambdaHandler {
	return &APILambdaHandler{
		config:               config,
		logger:               logger,
		messageSource:        messageSource,
		apiResponseFormatter: apiResponseFormatter,
	}
}

// GetDefaultGinEngine は、ginのEngineを取得します。
func (h *APILambdaHandler) GetDefaultGinEngine(errorResponse api.ErrorResponse, corsConfig *cors.Config) *gin.Engine {
	// ginをLoggerとCustomerRecoverのミドルウェアがアタッチされた状態で作成
	engine := gin.New()

	var middlewares []gin.HandlerFunc
	// ロガーのミドルウェアを追加
	middlewares = append(middlewares, gin.Logger())
	// CORSのミドルウェアを追加
	if corsConfig != nil {
		middlewares = append(middlewares, cors.New(*corsConfig))
	}
	// レスポンス生成のミドルウェアを追加
	middlewares = append(middlewares,
		func(ctx *gin.Context) {
			ctx.Next()
			// レスポンスの生成
			h.apiResponseFormatter.ReturnResponseBody(ctx, errorResponse)
		})
	// panic時のカスタムリカバリのミドルウェアを追加
	middlewares = append(middlewares,
		// パニック時のカスタムリカバリ処理
		gin.CustomRecovery(func(c *gin.Context, recover any) {
			// パニックをエラーでラップ
			err := errors.Errorf("recover from: %+v", recover)
			h.logger.ErrorWithUnexpectedError(err)
			// エラーをその他のエラー（ginのエラーログ対象外）として、ginのContextに格納
			c.Error(err).SetType(gin.ErrorTypeNu)
		}))

	// ミドルウェアをアタッチ
	engine.Use(middlewares...)

	// 404エラー
	engine.NoRoute(func(ctx *gin.Context) {
		h.logger.Debug("%s is not found", ctx.Request.URL.Path)
		// エラーをPublicなエラー（ginのエラーログ対象外）として、ginのContextに格納
		ctx.Error(api.NoRouteError).SetType(gin.ErrorTypePublic)
	})
	// 405エラー
	engine.HandleMethodNotAllowed = true
	engine.NoMethod(func(ctx *gin.Context) {
		h.logger.Debug("%s Method %s is not allowed", ctx.Request.Method, ctx.Request.URL.Path)
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
				h.logger.ErrorWithUnexpectedError(perr)
				// 想定外のエラーレスポンスの生成
				response = h.createUnexpectedErrorResponse(perr, errorResponse)
				// errがnilでないと{"message":"Internal Server Error"} のレスポンスが返却されるため、nilのまま
			}
			// ログのフラッシュ
			h.logger.Sync()
		}()
		// ctxをコンテキスト領域に格納
		apcontext.Context = ctx

		// リクエストIDをログの付加情報として追加
		h.logger.ClearInfo()
		lc := apcontext.GetLambdaContext(ctx)
		h.logger.AddInfo("AWS RequestID", lc.AwsRequestID)
		h.logger.AddInfo("API Gateway RequestID", request.RequestContext.RequestID)

		// AWS Lambda Go API Proxyでginと統合
		// https://github.com/awslabs/aws-lambda-go-api-proxy
		response, gerr := ginLambda.ProxyWithContext(ctx, request)
		if gerr != nil {
			h.logger.ErrorWithUnexpectedError(gerr)
			// 想定外のエラーレスポンスの生成
			response = h.createUnexpectedErrorResponse(gerr, errorResponse)
			// errがnilでないと{"message":"Internal Server Error"} のレスポンスが返却されるため、nilのまま
		}
		return
	}
}

// createUnexpectedErrorResponse は、予期せぬエラーによるAPIGatewayProxyResponseを作成します。
func (h *APILambdaHandler) createUnexpectedErrorResponse(err error, errorResponse api.ErrorResponse) events.APIGatewayProxyResponse {
	statusCode, body := errorResponse.UnexpectedErrorResponse(err)
	bbody, jerr := json.Marshal(body)
	if jerr != nil {
		h.logger.ErrorWithUnexpectedError(jerr)
		return events.APIGatewayProxyResponse{
			StatusCode: statusCode,
		}
	}
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Body:       string(bbody),
	}
}
