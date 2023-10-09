package handlerinterceptor

import (
	"errors"
	"reflect"
	"runtime"

	"example.com/appbase/pkg/api"
	myerrors "example.com/appbase/pkg/errors"
	"example.com/appbase/pkg/logging"
	"github.com/gin-gonic/gin"
)

type HandlerInterceptor struct {
	Log logging.Logger
}

type ControllerFunc func(ctx *gin.Context) (interface{}, error)

var businessError *myerrors.BusinessError
var systemError *myerrors.SystemError

func (i HandlerInterceptor) Handle(controllerFunc ControllerFunc) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		fv := reflect.ValueOf(controllerFunc)
		funcName := runtime.FuncForPC(fv.Pointer()).Name()
		i.Log.Info("Controller開始: %s", funcName)
		// Controllerの実行
		result, err := controllerFunc(ctx)
		// 集約エラーハンドリングによるログ出力
		if err != nil {
			if errors.As(err, &businessError) {
				i.Log.Warn(businessError.Error())
			} else if errors.As(err, &systemError) {
				i.Log.Fatal(systemError.Error())
			} else {
				i.Log.Fatal("予期せぬエラー: %s", err.Error())
			}
		} else {
			i.Log.Info("Controller正常終了: %s", funcName)
		}
		api.ReturnResponseBody(ctx, result, err)
	}
}
