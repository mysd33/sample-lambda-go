/*
erros パッケージは、エラー情報を扱うパッケージです。
*/
package errors

import (
	"fmt"

	cerrors "github.com/cockroachdb/errors"
)

// CodableErrorは、エラーコード定義付きのエラーインタフェースです。
type CodableError interface {
	error
	ErrorCode() string
	Args() []interface{}
}

// ValidationError は、入力エラーの構造体です。
type ValidationError struct {
	// TODO:
	cause error
}

// NewValidationError は、原因となるエラー（cause）をラップし、ValidationError構造体を作成します。
func NewValidationError(cause error) *ValidationError {
	return &ValidationError{cause: cerrors.WithStack(cause)}
}

// NewValidationErrorWithMessage は、メッセージをもとにBusinessError構造体を作成します。
func NewValidationErrorWithMessage(format string, args ...interface{}) *ValidationError {
	return &ValidationError{
		// cockloachdb/errorのスタックトレース付きのcauseエラー作成
		cause: cerrors.Errorf(format, args...),
	}
}

// Error は、エラーを返却します。
func (e *ValidationError) Error() string {
	// TODO:削除
	//log.Printf("入力エラーの型:%+v", reflect.TypeOf(errors.Unwrap(e.cause)))

	// Causeのツリーをたどってgo-playground/validatorのエラーを取得
	//var gPValidationErrors validator.ValidationErrors
	//if errors.As(e.cause, &gPValidationErrors) {
	//	for _, v := range gPValidationErrors {
	//		field := v.Field()
	//tag :=
	//	}
	//}
	return fmt.Sprintf("入力エラー:%s", e.cause.Error())
}

// UnWrap は、原因となるエラーにUnWrapします。
func (e *ValidationError) UnWrap() error {
	return e.cause
}

// BusinessError 業務エラーの構造体です。
type BusinessError struct {
	cause     error
	errorCode string
	args      []interface{}
}

// NewBusinessError は、BusinessError構造体を作成します。
func NewBusinessError(errorCode string, args ...interface{}) *BusinessError {
	// スタックトレース出力のため、cockloachdb/errorのスタックトレース付きのcauseエラー作成
	cause := cerrors.Errorf("code:%s, error:%v", errorCode, args)
	return &BusinessError{cause: cause, errorCode: errorCode, args: args}
}

// NewBusinessError は、原因となるエラー（cause）をラップし、
// メッセージIDにもなるエラーコード（errorCode）とメッセージの置換文字列(args）を渡し
// BusinessError構造体を作成します。
func NewBusinessErrorWithCause(cause error, errorCode string, args ...interface{}) *BusinessError {
	// causeはスタックトレース付与
	return &BusinessError{cause: cerrors.WithStack(cause), errorCode: errorCode, args: args}
}

// Error は、エラーを返却します。
func (e *BusinessError) Error() string {
	return fmt.Sprintf("業務エラー[%s], cause:%+v", e.errorCode, e.cause)
}

// UnWrap は、原因となるエラーにUnWrapします。
func (e *BusinessError) UnWrap() error {
	return e.cause
}

// ErrorCode は、エラーコード（メッセージID）を返します。
func (e *BusinessError) ErrorCode() string {
	return e.errorCode
}

// Argsは、エラーメッセージの置換文字列(args）を返します
func (e *BusinessError) Args() []interface{} {
	return e.args
}

// SystemError は、システムエラーの構造体
type SystemError struct {
	cause     error
	errorCode string
	args      []interface{}
}

// NewSystemError は、原因となるエラー（cause）をラップし、
// メッセージIDにもなるエラーコード（errorCode）とメッセージの置換文字列(args）を渡し
// SystemError構造体を作成します。
func NewSystemError(cause error, errorCode string, args ...interface{}) *SystemError {
	// causeはスタックトレース付与
	return &SystemError{cause: cerrors.WithStack(cause), errorCode: errorCode, args: args}
}

// Error は、エラーを返却します。
func (e *SystemError) Error() string {
	return fmt.Sprintf("システムエラー[%s], cause:%+v", e.errorCode, e.cause)
}

// UnWrap は、原因となるエラーにUnWrapします。
func (e *SystemError) UnWrap() error {
	return e.cause
}

// ErrorCode は、エラーコード（メッセージID）を返します。
func (e *SystemError) ErrorCode() string {
	return e.errorCode
}

// Argsは、エラーメッセージの置換文字列(args）を返します
func (e *SystemError) Args() []interface{} {
	return e.args
}
