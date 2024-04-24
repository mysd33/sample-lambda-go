package common

import (
	mymessage "app/internal/pkg/message"
	"net/http"

	"example.com/appbase/pkg/api"
	"example.com/appbase/pkg/errors"
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
func (r *commonErrorResponse) ValidationErrorResponse(validationError *errors.ValidationError) (int, any) {
	detail := make(map[string]any)
	detail["message"] = r.messageSource.GetMessage(validationError.ErrorCode(), validationError.Args()...)
	vErrDetails := make([]string, 0, len(validationError.ErrorDetails()))
	for _, v := range validationError.ErrorDetails() {
		vErrDetails = append(vErrDetails, v)
	}
	detail["errorDetails"] = vErrDetails
	return http.StatusBadRequest, r.errorResponseBody(validationError.ErrorCode(), detail)
}

// BusinessErrorResponse implements api.ErrorResponse.
func (r *commonErrorResponse) BusinessErrorResponse(businessErrors *errors.BusinessErrors) (int, any) {
	bizErrMsgs := make([]map[string]string, 0, len(businessErrors.BusinessErrors()))
	for _, businessError := range businessErrors.BusinessErrors() {
		bizErrMsg := make(map[string]string)
		bizErrMsg["code"] = businessError.ErrorCode()
		bizErrMsg["message"] = r.messageSource.GetMessage(businessError.ErrorCode(), businessError.Args()...)
		info1 := businessError.Info("info1")
		if info1Str, ok := info1.(string); ok {
			bizErrMsg["info1"] = info1Str
		}
		info2 := businessError.Info("info2")
		if info2Str, ok := info2.(string); ok {
			bizErrMsg["info2"] = info2Str
		}
		bizErrMsgs = append(bizErrMsgs, bizErrMsg)
	}
	var label string
	info := businessErrors.Info()
	if infoStr, ok := info.(string); ok {
		label = infoStr
	} else {
		label = "businessError"
	}
	return http.StatusBadRequest, r.errorResponseBody(label, bizErrMsgs)
}

// WarnErrorResponse implements api.ErrorResponse.
func (r *commonErrorResponse) WarnErrorResponse(err error) (int, any) {
	if errors.Is(err, api.NoRouteError) {
		return http.StatusNotFound, r.errorResponseBody(err.Error(), "")
	} else if errors.Is(err, api.NoMethodError) {
		return http.StatusMethodNotAllowed, r.errorResponseBody(err.Error(), "")
	} else {
		return http.StatusBadRequest, r.errorResponseBody(err.Error(), "")
	}
}

// SystemErrorResponse implements api.ErrorResponse.
func (r *commonErrorResponse) SystemErrorResponse(systemError *errors.SystemError) (int, any) {
	return http.StatusInternalServerError, r.errorResponseBody(systemError.ErrorCode(),
		r.messageSource.GetMessage(mymessage.E_EX_9001))
}

// UnexpectedErrorResponse implements api.ErrorResponse.
func (r *commonErrorResponse) UnexpectedErrorResponse(err error) (int, any) {
	return http.StatusInternalServerError, r.errorResponseBody(mymessage.E_EX_9999,
		r.messageSource.GetMessage(mymessage.E_EX_9999))
}

func (*commonErrorResponse) errorResponseBody(label string, detail any) gin.H {
	//TODO: 要件に応じてエラーレスポンスの形式を修正する。
	return gin.H{"code": label, "detail": detail}
}
