package handler

import (
	"errors"
	"reflect"
	"runtime"

	"example.com/appbase/pkg/api"
	myerrors "example.com/appbase/pkg/errors"
	"example.com/appbase/pkg/logging"
	"example.com/appbase/pkg/message"
	"github.com/gin-gonic/gin"
)

// HandlerInterceptor は、Handlerのインタセプタの構造体です。
type HandlerInterceptor struct {
	log                  logging.Logger
	apiResponseFormatter api.ApiResponseFormatter
}

// New は、HandlerInterceptor構造体を作成します。
func NewHandlerInterceptor(log logging.Logger, apiResponseFormatter api.ApiResponseFormatter) HandlerInterceptor {
	return HandlerInterceptor{log: log, apiResponseFormatter: apiResponseFormatter}
}

// ControllerFunc Controlerで実行する関数です。
type ControllerFunc func(ctx *gin.Context) (interface{}, error)

// Handle は、Controlerで実行する関数controllerFuncの前後でインタセプタの処理を実行します。
func (i HandlerInterceptor) Handle(controllerFunc ControllerFunc) gin.HandlerFunc {
	var (
		validationError *myerrors.ValidationError
		businessError   *myerrors.BusinessError
		systemError     *myerrors.SystemError
	)
	return func(ctx *gin.Context) {
		fv := reflect.ValueOf(controllerFunc)
		funcName := runtime.FuncForPC(fv.Pointer()).Name()

		i.log.Info(message.I_FW_0001, funcName)

		// Controllerの実行
		result, err := controllerFunc(ctx)
		// 集約エラーハンドリングによるログ出力
		if err != nil {
			if errors.As(err, &validationError) {
				i.log.Debug(validationError.Error())
			} else if errors.As(err, &businessError) {
				i.log.WarnWithCodableError(businessError)
			} else if errors.As(err, &systemError) {
				i.log.ErrorWithCodableError(systemError)
			} else {
				i.log.FatalWithCodableError(myerrors.NewSystemError(err, message.E_FW_9999))
			}
		} else {
			i.log.Info(message.I_FW_0002, funcName)
		}
		i.apiResponseFormatter.ReturnResponseBody(ctx, result, err)
	}
}
