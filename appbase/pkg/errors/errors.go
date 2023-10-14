/*
erros パッケージは、エラー情報を扱うパッケージです。
*/
package errors

import (
	"fmt"
)

// ValidationError は、入力エラーの構造体です。
type ValidationError struct {
	// TODO:
}

// NewValidationError は、ValidationError構造体を作成します。
func NewValidationError() *ValidationError {
	return &ValidationError{
		// TODO:
	}
}

// Error は、エラーを返却します。
func (e *ValidationError) Error() string {
	// TODO:
	return "入力エラー"
}

// BusinessError 業務エラーの構造体です。
type BusinessError struct {
	ErrorCode string
	Cause     error
}

// NewBusinessError は、BusinessError構造体を作成します。
func NewBusinessError(errorCode string, cause error) *BusinessError {
	return &BusinessError{ErrorCode: errorCode, Cause: cause}
}

// Error は、エラーを返却します。
func (e *BusinessError) Error() string {
	return fmt.Sprintf("業務エラー[%s]:%s", e.ErrorCode, e.Cause.Error())
}

// UnWrap は、原因となるエラーにUnWrapします。
func (e *BusinessError) UnWrap() error {
	return e.Cause
}

// SystemError は、システムエラーの構造体
type SystemError struct {
	ErrorCode string
	Cause     error
}

// NewSystemError は、SystemError構造体を作成します。
func NewSystemError(errorCode string, cause error) *SystemError {
	return &SystemError{ErrorCode: errorCode, Cause: cause}
}

// Error は、エラーを返却します。
func (e *SystemError) Error() string {
	return fmt.Sprintf("システムエラー[%s]:%s", e.ErrorCode, e.Cause.Error())
}

// UnWrap は、原因となるエラーにUnWrapします。
func (e *SystemError) UnWrap() error {
	return e.Cause
}
