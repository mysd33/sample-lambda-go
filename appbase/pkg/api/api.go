/*
api パッケージは、REST APIに関する機能を提供するパッケージです。
*/
package api

import (
	"encoding/json"
	"net/http"

	"example.com/appbase/pkg/constant"
	myerrors "example.com/appbase/pkg/errors"
	"example.com/appbase/pkg/logging"
	"example.com/appbase/pkg/message"
	"github.com/aws/aws-lambda-go/events"
	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
)

var (
	NoRouteError  = errors.New("NOT FOUND")
	NoMethodError = errors.New("METHOD NOT ALLOWED")
)

// ApiResponseFormatterは、レスポンスデータを作成するインタフェース
type ApiResponseFormatter interface {
	// ReturnResponseBody は、ginのContextに対して処理結果resultまたはエラーerrに対応するレスポンスボディを設定します。
	ReturnResponseBody(ctx *gin.Context, errorResponse ErrorResponse)
	// CreateAPIGatewayProxyResponseForUnexpectedError は、予期せぬエラーによるAPIGatewayProxyResponseを作成します。
	CreateAPIGatewayProxyResponseForUnexpectedError(err error, errorResponse ErrorResponse) (events.APIGatewayProxyResponse, error)
}

// NewApiResponseFormatter は、ApiResponseFormatterを作成します。
func NewApiResponseFormatter(log logging.Logger, messageSource message.MessageSource) ApiResponseFormatter {
	return &defaultApiResponseFormatter{log: log, messageSource: messageSource}
}

type defaultApiResponseFormatter struct {
	log           logging.Logger
	messageSource message.MessageSource
}

// ReturnResponseBody implements ApiResponseFormatter.
func (f *defaultApiResponseFormatter) ReturnResponseBody(ctx *gin.Context, errorResponse ErrorResponse) {
	var (
		validationError *myerrors.ValidationError
		businessErrors  *myerrors.BusinessErrors
		systemError     *myerrors.SystemError
	)
	errs := ctx.Errors

	if len(errs) > 0 {
		// エラーの場合
		err := convertSingleError(errs)

		// 各エラー内容に応じたレスポンスを作成
		if errors.As(err, &validationError) {
			ctx.JSON(errorResponse.ValidationErrorResponse(validationError))
		} else if errors.As(err, &businessErrors) {
			ctx.JSON(errorResponse.BusinessErrorResponse(businessErrors))
		} else if errors.As(err, &systemError) {
			ctx.JSON(errorResponse.SystemErrorResponse(systemError))
		} else if errors.Is(err, NoRouteError) {
			ctx.JSON(errorResponse.WarnErrorResponse(NoRouteError))
		} else if errors.Is(err, NoMethodError) {
			ctx.JSON(errorResponse.WarnErrorResponse(NoMethodError))
		} else {
			ctx.JSON(errorResponse.UnexpectedErrorResponse(err))
		}
	} else {
		// 正常終了の場合
		result, ok := ctx.Get(constant.CONTROLLER_RESULT)
		if ok {
			// 処理結果の構造体をもとにレスポンスを作成
			ctx.JSON(http.StatusOK, result)
			return
		}
		// resultが取得できなかった場合には予期せぬエラーとしてログを出力し、エラーを返却する
		err := errors.New("result is not found")
		f.log.ErrorWithUnexpectedError(err)
		// 予期せぬエラー扱いのレスポンスを返却
		ctx.JSON(errorResponse.UnexpectedErrorResponse(err))
	}
}

// CreateAPIGatewayProxyResponseForUnexpectedError implements ApiResponseFormatter.
func (f *defaultApiResponseFormatter) CreateAPIGatewayProxyResponseForUnexpectedError(err error, errorResponse ErrorResponse) (events.APIGatewayProxyResponse, error) {
	statusCode, body := errorResponse.UnexpectedErrorResponse(err)
	bbody, jerr := json.Marshal(body)
	if jerr != nil {
		return events.APIGatewayProxyResponse{}, jerr
	}
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Body:       string(bbody),
	}, nil
}

// convertSingleError は、エラーが複数あった場合にも1つのエラーに変換します。
func convertSingleError(errs []*gin.Error) error {
	if len(errs) == 1 {
		return errs[0]
	}
	// 複数エラーがある場合は、エラーを結合
	var converted []error
	for _, err := range errs {
		converted = append(converted, err.Err)
	}
	return errors.Join(converted...)
}
