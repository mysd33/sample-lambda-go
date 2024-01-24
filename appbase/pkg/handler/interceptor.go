package handler

import (
	"errors"
	"reflect"
	"runtime"

	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/constant"
	"example.com/appbase/pkg/env"
	myerrors "example.com/appbase/pkg/errors"
	"example.com/appbase/pkg/logging"
	"example.com/appbase/pkg/message"
	"github.com/aws/aws-lambda-go/events"
	"github.com/gin-gonic/gin"
)

const (
	GIN_DEBUG_NAME = "GIN_DEBUG"
)

type HandlerInterceptor interface {
	// Handleは、同期処理のControllerに割り込み、集約例外ハンドリング等の共通処理を実施します。
	Handle(controllerFunc ControllerFunc) gin.HandlerFunc
	// HandleAsyncは、非同期処理のControllerに割り込み、集約例外ハンドリング等の共通処理を実施します。
	HandleAsync(asyncControllerFunc AsyncControllerFunc) AsyncControllerFunc
	// HandleSimpleは、その他のトリガのControllerに割り込み、集約例外ハンドリング等の共通処理を実施します。
	HandleSimple(simpleControllerFunc SimpleControllerFunc) SimpleControllerFunc
}

// HandlerInterceptor は、Handlerのインタセプタの構造体です。
type defaultHandlerInterceptor struct {
	config config.Config
	log    logging.Logger
}

// NewHandlerInterceptor は、HandlerInterceptorを作成します。
func NewHandlerInterceptor(config config.Config, log logging.Logger) HandlerInterceptor {
	ginDebugMode := config.GetBool(GIN_DEBUG_NAME, false)
	if env.IsStragingOrProd() && ginDebugMode {
		// 本番相当の動作環境の場合、ginのモードを本番モードに設定
		gin.SetMode(gin.ReleaseMode)
	}
	return &defaultHandlerInterceptor{config: config, log: log}
}

// Handle は、Controlerで実行する関数controllerFuncの前後でインタセプタの処理を実行します。
func (i *defaultHandlerInterceptor) Handle(controllerFunc ControllerFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		fv := reflect.ValueOf(controllerFunc)
		funcName := runtime.FuncForPC(fv.Pointer()).Name()
		i.log.Info(message.I_FW_0001, funcName)
		// Configの最新読み込み
		if err := i.config.Reload(); err != nil {
			// エラーをContextに格納
			ctx.Error(err)
			return
		}
		// Controllerの実行
		result, err := controllerFunc(ctx)
		if err != nil {
			// 集約エラーハンドリングによるログ出力
			i.logError(err)
			// エラーをContextに格納
			ctx.Error(err)
			return
		}

		// 処理結果をContextに格納
		ctx.Set(constant.CONTROLLER_RESULT, result)
		i.log.Info(message.I_FW_0002, funcName)
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
			return err
		}
		i.log.Info(message.I_FW_0002, funcName)
		return nil
	}
}

// HandleSimple implements HandlerInterceptor.
func (i *defaultHandlerInterceptor) HandleSimple(simpleControllerFunc SimpleControllerFunc) SimpleControllerFunc {
	return func(event any) error {
		fv := reflect.ValueOf(simpleControllerFunc)
		funcName := runtime.FuncForPC(fv.Pointer()).Name()
		i.log.Info(message.I_FW_0001, funcName)

		// Configの最新読み込み
		if err := i.config.Reload(); err != nil {
			return err
		}
		// Controllerの実行
		err := simpleControllerFunc(event)
		// 集約エラーハンドリングによるログ出力
		if err != nil {
			i.logError(err)
			return err
		}
		i.log.Info(message.I_FW_0002, funcName)
		return nil
	}
}

// logError は、エラー情報をログ出力します
func (i *defaultHandlerInterceptor) logError(err error) {
	var (
		validationError *myerrors.ValidationError
		businessErrors  *myerrors.BusinessErrors
		systemError     *myerrors.SystemError
	)
	if errors.As(err, &validationError) {
		i.log.WarnWithCodableError(validationError)
	} else if errors.As(err, &businessErrors) {
		i.log.WarnWithMultiCodableError(businessErrors)
	} else if errors.As(err, &systemError) {
		i.log.ErrorWithCodableError(systemError)
	} else {
		i.log.ErrorWithUnexpectedError(err)
	}
}
