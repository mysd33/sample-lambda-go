/*
erros パッケージは、エラー情報を扱うパッケージです。
*/
package errors

import (
	"fmt"

	"github.com/cockroachdb/errors"
)

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
	if cause == nil {
		// nilの場合、ダミーのエラーを作成
		cause = errors.NewWithDepthf(1, "code:%s, args:%v", errorCode, args)
	} else {
		cause = errors.WithStackDepth(cause, 1)
	}
	// 誤ったエラーのラップを確認
	requiredNotCodableError(cause)
	// causeはスタックトレース付与
	return &SystemError{cause: cause, errorCode: errorCode, args: args}
}

// Error は、エラーを返却します。errorインタフェースを実装します。
func (e *SystemError) Error() string {
	return fmt.Sprintf("システムエラー[%s]", e.errorCode)
}

// Format は、%+vを正しく動作させるため、fmt.Formatterのインタフェースを実装します。
// https://github.com/cockroachdb/errors#Making-v-work-with-your-type
func (e *SystemError) Format(s fmt.State, verb rune) { errors.FormatError(e, s, verb) }

// Unwrap は、原因となるエラーにUnwrapします。
// https://github.com/cockroachdb/errors#building-your-own-error-types
func (e *SystemError) Unwrap() error {
	return e.cause
}

// ErrorCode implements CodableError.
func (e *SystemError) ErrorCode() string {
	return e.errorCode
}

// Args implements CodableError.
func (e *SystemError) Args() []any {
	return e.args
}