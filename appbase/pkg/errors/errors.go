/*
erros パッケージは、エラー情報を扱うパッケージです。
*/
package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"example.com/appbase/pkg/env"
	myvalidator "example.com/appbase/pkg/validator"
	cerrors "github.com/cockroachdb/errors"
	"github.com/go-playground/validator/v10"
)

// CodableErrorは、エラーコード定義付きのエラーインタフェースです。
type CodableError interface {
	error
	ErrorCode() string
	Args() []any
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
func NewValidationErrorWithMessage(format string, args ...any) *ValidationError {
	return &ValidationError{
		// cockloachdb/errorのスタックトレース付きのcauseエラー作成
		cause: cerrors.Errorf(format, args...),
	}
}

// Error は、エラーを返却します。
func (e *ValidationError) Error() string {
	//Causeのツリーをたどってgo-playground/validatorのエラーを取得
	var gPValidationErrors validator.ValidationErrors
	if errors.As(e.cause, &gPValidationErrors) {
		if myvalidator.Translator != nil {
			//TODO: バリデーションエラーメッセージの整形（暫定でそのままJSON文字列）
			//エラーメッセージの日本語化
			translated := gPValidationErrors.Translate(myvalidator.Translator)
			bytes, _ := json.Marshal(translated)
			return fmt.Sprintf("入力エラー:%s", string(bytes))
		}
	}

	return fmt.Sprintf("入力エラー:%s", e.cause.Error())
}

// Unwrap は、原因となるエラーにUnwrapします。
func (e *ValidationError) Unwrap() error {
	return e.cause
}

// BusinessError 業務エラーの構造体です。
type BusinessError struct {
	cause     error
	errorCode string
	args      []any
}

// NewBusinessError は、BusinessError構造体を作成します。
func NewBusinessError(errorCode string, args ...any) *BusinessError {
	// スタックトレース出力のため、cockloachdb/errorのスタックトレース付きのcauseエラー作成
	cause := cerrors.NewWithDepthf(1, "code:%s, error:%v", errorCode, args)
	return &BusinessError{cause: cause, errorCode: errorCode, args: args}
}

// NewBusinessError は、原因となるエラー（cause）をラップし、
// メッセージIDにもなるエラーコード（errorCode）とメッセージの置換文字列(args）を渡し
// BusinessError構造体を作成します。
func NewBusinessErrorWithCause(cause error, errorCode string, args ...any) *BusinessError {
	// 誤ったエラーのラップを確認
	requiredNotBusinessAndSystemError(cause)
	// causeはスタックトレース付与
	return &BusinessError{cause: cerrors.WithStackDepth(cause, 1), errorCode: errorCode, args: args}
}

// Error は、エラーを返却します。errorインタフェースを実装します。
func (e *BusinessError) Error() string {
	return fmt.Sprintf("業務エラー[%s], cause:%+v", e.errorCode, e.cause)
}

// Format は、%+vを正しく動作させるため、fmt.Formatterのインタフェースを実装します。
// https://github.com/cockroachdb/errors#Making-v-work-with-your-type
func (e *BusinessError) Format(s fmt.State, verb rune) { cerrors.FormatError(e, s, verb) }

// Unwrap は、原因となるエラーにUnwrapします。
// https://github.com/cockroachdb/errors#building-your-own-error-types
func (e *BusinessError) Unwrap() error {
	return e.cause
}

// ErrorCode は、エラーコード（メッセージID）を返します。
func (e *BusinessError) ErrorCode() string {
	return e.errorCode
}

// Argsは、エラーメッセージの置換文字列(args）を返します
func (e *BusinessError) Args() []any {
	return e.args
}

// SystemError は、システムエラーの構造体
type SystemError struct {
	cause     error
	errorCode string
	args      []any
}

// NewSystemError は、原因となるエラー（cause）をラップし、
// メッセージIDにもなるエラーコード（errorCode）とメッセージの置換文字列(args）を渡し
// SystemError構造体を作成します。
func NewSystemError(cause error, errorCode string, args ...any) *SystemError {
	// 誤ったエラーのラップを確認
	requiredNotBusinessAndSystemError(cause)
	// causeはスタックトレース付与
	return &SystemError{cause: cerrors.WithStackDepth(cause, 1), errorCode: errorCode, args: args}
}

// Error は、エラーを返却します。errorインタフェースを実装します。
func (e *SystemError) Error() string {
	return fmt.Sprintf("システムエラー[%s], cause:%+v", e.errorCode, e.cause)
}

// Format は、%+vを正しく動作させるため、fmt.Formatterのインタフェースを実装します。
// https://github.com/cockroachdb/errors#Making-v-work-with-your-type
func (e *SystemError) Format(s fmt.State, verb rune) { cerrors.FormatError(e, s, verb) }

// Unwrap は、原因となるエラーにUnwrapします。
// https://github.com/cockroachdb/errors#building-your-own-error-types
func (e *SystemError) Unwrap() error {
	return e.cause
}

// ErrorCode は、エラーコード（メッセージID）を返します。
func (e *SystemError) ErrorCode() string {
	return e.errorCode
}

// Argsは、エラーメッセージの置換文字列(args）を返します
func (e *SystemError) Args() []any {
	return e.args
}

// CauseがBusinessError、SystemErrorでないことを確認
func requiredNotBusinessAndSystemError(cause error) {
	if cause == nil {
		return
	}
	var be *BusinessError
	var se *SystemError
	// causeがBusinessError、SystemErrorの場合は、
	// コーディングミスで二重でラップしてしまっている判断して、開発中は異常終了させている
	if errors.As(cause, &be) {
		if !env.IsProd() {
			// 異常終了
			panic(fmt.Sprintf("誤ってBusinessErrorを二重でラップしています:%+v", be))
		}
		log.Printf("誤ってBusinessErrorを二重でラップしています:%+v", be)
	} else if errors.As(cause, &se) {
		if !env.IsProd() {
			// 異常終了
			panic(fmt.Sprintf("誤ってSystemErrorを二重でラップしています:%+v", se))
		}
		log.Printf("誤ってSystemErrorを二重でラップしています:%+v", se)
	}
}
