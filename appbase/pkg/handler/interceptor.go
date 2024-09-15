package handler

import (
	"context"
	"reflect"
	"runtime"
	"time"

	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/constant"
	"example.com/appbase/pkg/env"
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
	logger logging.Logger
}

// NewHandlerInterceptor は、HandlerInterceptorを作成します。
func NewHandlerInterceptor(config config.Config, logger logging.Logger) HandlerInterceptor {
	ginDebugMode := config.GetBool(GIN_DEBUG_NAME, false)
	if env.IsStragingOrProd() && ginDebugMode {
		// 本番相当の動作環境の場合、ginのモードを本番モードに設定
		gin.SetMode(gin.ReleaseMode)
	}
	return &defaultHandlerInterceptor{config: config, logger: logger}
}

// Handle は、Controlerで実行する関数controllerFuncの前後でインタセプタの処理を実行します。
func (i *defaultHandlerInterceptor) Handle(controllerFunc ControllerFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		fv := reflect.ValueOf(controllerFunc)
		funcName := runtime.FuncForPC(fv.Pointer()).Name()
		i.logger.Info(message.I_FW_0001, funcName)
		startTime := time.Now()
		defer func() {
			i.logger.Info(message.I_FW_0002, funcName, time.Since(startTime))
		}()

		// Configの最新読み込み
		if err := i.config.Reload(); err != nil {
			// エラーログの出力
			logging.LogError(i.logger, err)
			// エラーをその他のエラー（ginのエラーログ対象外）としてginのContextに格納
			ctx.Error(err).SetType(gin.ErrorTypeNu)
			return
		}
		// Controllerの実行
		result, err := controllerFunc(ctx)
		if err != nil {
			// 集約エラーハンドリングによるログ出力
			logging.LogError(i.logger, err)
			// エラーをPublicなエラー（ginのエラーログ対象外）としてginのContextに格納
			ctx.Error(err).SetType(gin.ErrorTypePublic)
			return
		}

		// 処理結果をContextに格納
		ctx.Set(constant.CONTROLLER_RESULT, result)
	}
}

// HandleAsync implements HandlerInterceptor.
func (i *defaultHandlerInterceptor) HandleAsync(asyncControllerFunc AsyncControllerFunc) AsyncControllerFunc {
	return func(sqsMessage events.SQSMessage) error {
		fv := reflect.ValueOf(asyncControllerFunc)
		funcName := runtime.FuncForPC(fv.Pointer()).Name()
		i.logger.Info(message.I_FW_0001, funcName)
		startTime := time.Now()
		defer func() {
			i.logger.Info(message.I_FW_0002, funcName, time.Since(startTime))
		}()

		// Configの最新読み込み
		if err := i.config.Reload(); err != nil {
			logging.LogError(i.logger, err)
			return err
		}
		// Controllerの実行
		err := asyncControllerFunc(sqsMessage)
		// 集約エラーハンドリングによるログ出力
		if err != nil {
			logging.LogError(i.logger, err)
			return err
		}
		return nil
	}
}

// HandleSimple implements HandlerInterceptor.
func (i *defaultHandlerInterceptor) HandleSimple(simpleControllerFunc SimpleControllerFunc) SimpleControllerFunc {
	return func(ctx context.Context, event any) (any, error) {
		fv := reflect.ValueOf(simpleControllerFunc)
		funcName := runtime.FuncForPC(fv.Pointer()).Name()
		i.logger.Info(message.I_FW_0001, funcName)
		startTime := time.Now()
		defer func() {
			i.logger.Info(message.I_FW_0002, funcName, time.Since(startTime))
		}()

		// Configの最新読み込み
		if err := i.config.Reload(); err != nil {
			logging.LogError(i.logger, err)
			return nil, err
		}
		// Controllerの実行
		result, err := simpleControllerFunc(ctx, event)
		// 集約エラーハンドリングによるログ出力
		if err != nil {
			logging.LogError(i.logger, err)
			return nil, err
		}
		return result, nil
	}
}
