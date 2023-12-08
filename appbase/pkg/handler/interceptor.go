package handler

import (
	"errors"
	"reflect"
	"runtime"

	"example.com/appbase/pkg/api"
	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/constant"
	"example.com/appbase/pkg/env"
	myerrors "example.com/appbase/pkg/errors"
	"example.com/appbase/pkg/logging"
	"example.com/appbase/pkg/message"
	"github.com/aws/aws-lambda-go/events"
	"github.com/gin-gonic/gin"
)

type HandlerInterceptor interface {
	// Handleは、同期処理のControllerに割り込み、集約例外ハンドリング等の共通処理を実施します。
	Handle(controllerFunc ControllerFunc) gin.HandlerFunc

	//TODO: APIの在り方検討中
	// HandleAsyncは、非同期処理のControllerに割り込み、集約例外ハンドリング等の共通処理を実施します。
	HandleAsync(asyncControllerFunc AsyncControllerFunc) AsyncControllerFunc
}

// HandlerInterceptor は、Handlerのインタセプタの構造体です。
type defaultHandlerInterceptor struct {
	config               config.Config
	log                  logging.Logger
	apiResponseFormatter api.ApiResponseFormatter
}

// NewHandlerInterceptor は、HandlerInterceptorを作成します。
func NewHandlerInterceptor(config config.Config, log logging.Logger, apiResponseFormatter api.ApiResponseFormatter) HandlerInterceptor {
	ginDebugMode := config.Get(constant.GIN_DEBUG_NAME)
	if env.IsStragingOrProd() && ginDebugMode != "true" {
		// 本番相当の動作環境の場合、ginのモードを本番モードに設定
		gin.SetMode(gin.ReleaseMode)
	}
	return &defaultHandlerInterceptor{config: config, log: log, apiResponseFormatter: apiResponseFormatter}
}

// Handle は、Controlerで実行する関数controllerFuncの前後でインタセプタの処理を実行します。
func (i *defaultHandlerInterceptor) Handle(controllerFunc ControllerFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		fv := reflect.ValueOf(controllerFunc)
		funcName := runtime.FuncForPC(fv.Pointer()).Name()
		i.log.Info(message.I_FW_0001, funcName)

		// Configの最新読み込み
		if err := i.config.Reload(); err != nil {
			//TODO: レスポンス作成を上位のMiddlewareコンポーネントで実施するよう修正
			i.apiResponseFormatter.ReturnResponseBody(ctx, nil, err)
			return
		}

		// Controllerの実行
		result, err := controllerFunc(ctx)
		// 集約エラーハンドリングによるログ出力
		if err != nil {
			i.logError(err)
		} else {
			i.log.Info(message.I_FW_0002, funcName)
		}
		//TODO: レスポンス作成を上位のMiddlewareコンポーネントで実施するよう修正
		i.apiResponseFormatter.ReturnResponseBody(ctx, result, err)
	}
}

// HandleAsync implements HandlerInterceptor.
func (i *defaultHandlerInterceptor) HandleAsync(asyncControllerFunc AsyncControllerFunc) AsyncControllerFunc {
	return func(sqsMessage events.SQSMessage) error {
		fv := reflect.ValueOf(asyncControllerFunc)
		funcName := runtime.FuncForPC(fv.Pointer()).Name()
		i.log.Info(message.I_FW_0001, funcName)

		// Configの最新読み込み
		if err := i.config.Reload(); err != nil {
			return err
		}
		// Controllerの実行
		err := asyncControllerFunc(sqsMessage)
		// 集約エラーハンドリングによるログ出力
		if err != nil {
			i.logError(err)
		} else {
			i.log.Info(message.I_FW_0002, funcName)
		}
		return nil
	}
}

// logError は、エラー情報をログ出力します
func (i *defaultHandlerInterceptor) logError(err error) {
	var (
		validationError *myerrors.ValidationError
		businessError   *myerrors.BusinessError
		systemError     *myerrors.SystemError
	)
	if errors.As(err, &validationError) {
		i.log.Debug(validationError.Error())
	} else if errors.As(err, &businessError) {
		i.log.WarnWithCodableError(businessError)
	} else if errors.As(err, &systemError) {
		i.log.ErrorWithCodableError(systemError)
	} else {
		i.log.ErrorWithUnexpectedError(err)
	}
}
