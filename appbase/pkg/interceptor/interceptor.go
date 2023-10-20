/*
interceptorパッケージは、Handlerの処理へ挟みこむInterceptor機能を提供するパッケージです
*/
package interceptor

import (
	"errors"
	"reflect"
	"runtime"

	"example.com/appbase/pkg/api"
	myerrors "example.com/appbase/pkg/errors"
	"example.com/appbase/pkg/logging"
	"github.com/gin-gonic/gin"
)

// HandlerInterceptor は、Handlerのインタセプタの構造体です。
type HandlerInterceptor struct {
	log logging.Logger
}

// New は、HandlerInterceptor構造体を作成します。
func New(log logging.Logger) HandlerInterceptor {
	return HandlerInterceptor{log: log}
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

		// TODO: ログの出力形式の修正
		// i.log.Info(code.I_FW_0001, funcName)
		i.log.Info("Controller開始: %s", funcName)

		// Controllerの実行
		result, err := controllerFunc(ctx)
		// 集約エラーハンドリングによるログ出力
		if err != nil {
			if errors.As(err, &validationError) {
				i.log.Debug(validationError.Error())
			} else if errors.As(err, &businessError) {
				i.log.Warn(businessError.ErrorCode, businessError.Args)
			} else if errors.As(err, &systemError) {
				i.log.Error(systemError.ErrorCode, systemError.Args)
			} else {
				i.log.Fatal("予期せぬエラー: %s", err.Error())
			}
		} else {
			// TODO: ログの出力形式の修正
			// i.log.Info(code.I_FW_0002, funcName)
			i.log.Info("Controller正常終了: %s", funcName)
		}
		api.ReturnResponseBody(ctx, result, err)
	}
}
