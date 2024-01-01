package common

import (
	mymessage "app/internal/pkg/message"
	"errors"
	"net/http"

	"example.com/appbase/pkg/api"
	myerrors "example.com/appbase/pkg/errors"
	"example.com/appbase/pkg/handler"
	"example.com/appbase/pkg/message"
	"github.com/gin-gonic/gin"
)

// NewCommonErrorResponse は、ErrorResponseを作成します。
func NewCommonErrorResponse(messageSource message.MessageSource) api.ErrorResponse {
	return &commonErrorResponse{
		messageSource: messageSource,
	}
}

// commonErrorResponse は、業務共通のエラーレスポンスを定義する構造体です。
type commonErrorResponse struct {
	messageSource message.MessageSource
}

// ValidationErrorResponse implements api.ErrorResponse.
func (r *commonErrorResponse) ValidationErrorResponse(validationError *myerrors.ValidationError) (int, any) {
	return http.StatusBadRequest, r.errorResponseBody("validationError", validationError.Error())
}

// BusinessErrorResponse implements api.ErrorResponse.
func (r *commonErrorResponse) BusinessErrorResponse(businessError *myerrors.BusinessError) (int, any) {
	return http.StatusBadRequest, r.errorResponseBody(businessError.ErrorCode(),
		r.messageSource.GetMessage(businessError.ErrorCode(), businessError.Args()...))
}

// WarnErrorResponse implements api.ErrorResponse.
func (r *commonErrorResponse) WarnErrorResponse(err error) (int, any) {
	if errors.Is(err, handler.NoRouteError) {
		return http.StatusNotFound, r.errorResponseBody(err.Error(), "")
	} else if errors.Is(err, handler.NoMethodError) {
		return http.StatusMethodNotAllowed, r.errorResponseBody(err.Error(), "")
	} else {
		return http.StatusBadRequest, r.errorResponseBody(err.Error(), "")
	}
}

// SystemErrorResponse implements api.ErrorResponse.
func (r *commonErrorResponse) SystemErrorResponse(systemError *myerrors.SystemError) (int, any) {
	return http.StatusInternalServerError, r.errorResponseBody(systemError.ErrorCode(),
		r.messageSource.GetMessage(mymessage.E_EX_9001))
}

// UnExpectedErrorResponse implements api.ErrorResponse.
func (r *commonErrorResponse) UnExpectedErrorResponse(err error) (int, any) {
	return http.StatusInternalServerError, r.errorResponseBody(mymessage.E_EX_9001,
		r.messageSource.GetMessage(mymessage.E_EX_9999))
}

func (*commonErrorResponse) errorResponseBody(label string, detail string) gin.H {
	//TODO: 要件に応じてエラーレスポンスの形式を修正する。
	return gin.H{"code": label, "detail": detail}
}
