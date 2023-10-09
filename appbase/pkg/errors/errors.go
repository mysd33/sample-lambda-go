package errors

import (
	"fmt"
)

// TODO: 入力エラー構造体
type ValidationError struct {
	// TODO:
}

func NewValidationError() *ValidationError {
	return &ValidationError{
		// TODO:
	}
}

func (e *ValidationError) Error() string {
	// TODO:
	return "入力エラー"
}

// TODO: 業務エラー構造体
type BusinessError struct {
	ErrorCode string
	Cause     error
}

func NewBusinessError(errorCode string, cause error) *BusinessError {
	return &BusinessError{ErrorCode: errorCode, Cause: cause}
}

func (e *BusinessError) Error() string {
	return fmt.Sprintf("業務エラー[%s]:%s", e.ErrorCode, e.Cause.Error())
}

func (e *BusinessError) UnWrap() error {
	return e.Cause
}

// TODO: システムエラー構造体
type SystemError struct {
	ErrorCode string
	Cause     error
}

func NewSystemError(errorCode string, cause error) *SystemError {
	return &SystemError{ErrorCode: errorCode, Cause: cause}
}

func (e *SystemError) Error() string {
	return fmt.Sprintf("システムエラー[%s]:%s", e.ErrorCode, e.Cause.Error())
}

func (e *SystemError) UnWrap() error {
	return e.Cause
}
