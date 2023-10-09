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

type HandlerInterceptor struct {
	log logging.Logger
}

func New(log logging.Logger) HandlerInterceptor {
	return HandlerInterceptor{log: log}
}

type ControllerFunc func(ctx *gin.Context) (interface{}, error)

var businessError *myerrors.BusinessError
var systemError *myerrors.SystemError

func (i HandlerInterceptor) Handle(controllerFunc ControllerFunc) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		fv := reflect.ValueOf(controllerFunc)
		funcName := runtime.FuncForPC(fv.Pointer()).Name()
		i.log.Info("Controller開始: %s", funcName)
		// Controllerの実行
		result, err := controllerFunc(ctx)
		// 集約エラーハンドリングによるログ出力
		if err != nil {
			if errors.As(err, &businessError) {
				i.log.Warn(businessError.Error())
			} else if errors.As(err, &systemError) {
				i.log.Fatal(systemError.Error())
			} else {
				i.log.Fatal("予期せぬエラー: %s", err.Error())
			}
		} else {
			i.log.Info("Controller正常終了: %s", funcName)
		}
		api.ReturnResponseBody(ctx, result, err)
	}
}
