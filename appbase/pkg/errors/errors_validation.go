/*
erros パッケージは、エラー情報を扱うパッケージです。
*/
package errors

import (
	"fmt"

	myvalidator "example.com/appbase/pkg/validator"
	"github.com/cockroachdb/errors"
	"github.com/go-playground/validator/v10"
)

// ValidationError は、入力エラーの構造体です。
type ValidationError struct {
	cause     error
	errorCode string
	args      []any
}

// NewValidationError は、ッセージIDにもなるエラーコード（errorCode）とメッセージの置換文字列(args）を渡し
// ValidationError構造体を作成します。
func NewValidationError(errorCode string, args ...any) *ValidationError {
	// スタックトレース出力のため、cockloachdb/errorのダミーのcauseエラー作成
	cause := errors.NewWithDepthf(1, "code:%s, args:%v", errorCode, args)
	return &ValidationError{cause: cause, errorCode: errorCode, args: args}
}

// NewValidationErrorWithCause は、原因となるエラー（cause）をラップし、
// メッセージIDにもなるエラーコード（errorCode）とメッセージの置換文字列(args）を渡しValidationError構造体を作成します。
func NewValidationErrorWithCause(cause error, errorCode string, args ...any) *ValidationError {
	if cause == nil {
		// nilの場合、ダミーのエラーを作成
		cause = errors.NewWithDepthf(1, "code:%s, args:%v", errorCode, args)
	} else {
		cause = errors.WithStackDepth(cause, 1)
	}

	// 誤ったエラーのラップを確認
	requiredNotCodableError(cause)
	return &ValidationError{cause: cause, errorCode: errorCode, args: args}
}

// ErrorDetails は、エラー詳細を返します。
func (e *ValidationError) ErrorDetails() map[string]string {
	// Causeのツリーをたどってgo-playground/validatorのエラーを取得
	var gPValidationErrors validator.ValidationErrors
	if errors.As(e.cause, &gPValidationErrors) {
		if myvalidator.Translator != nil {
			//TODO: バリデーションエラーメッセージの整形（現状、そのままmapで出力）
			//エラーメッセージの日本語化
			return gPValidationErrors.Translate(myvalidator.Translator)
		}
	}
	// ValidationErrorsがない場合は、空のmapを返却
	return make(map[string]string)
}

// Error は、エラーを返却します。
func (e *ValidationError) Error() string {
	return fmt.Sprintf("入力エラー[%s] details:%v", e.errorCode, e.ErrorDetails())
}

// Unwrap は、原因となるエラーにUnwrapします。
func (e *ValidationError) Unwrap() error {
	return e.cause
}

// Args implements CodableError.
func (e *ValidationError) Args() []any {
	return e.args
}

// ErrorCode implements CodableError.
func (e *ValidationError) ErrorCode() string {
	return e.errorCode
}

// Format は、%+vを正しく動作させるため、fmt.Formatterのインタフェースを実装します。
// https://github.com/cockroachdb/errors#Making-v-work-with-your-type
func (e *ValidationError) Format(s fmt.State, verb rune) { errors.FormatError(e, s, verb) }
