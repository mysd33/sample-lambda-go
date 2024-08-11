/*
erros パッケージは、エラー情報を扱うパッケージです。
*/
package errors

import (
	"fmt"

	"github.com/cockroachdb/errors"
)

// OtherError は、入力エラー、業務エラー、システムエラー以外でのソフトウェアフレームワーク起因のエラーを表す構造体
type OtherError struct {
	cause     error
	errorCode string
	args      []any
}

func NewOtherError(cause error, errorCode string, args ...any) *OtherError {
	if cause == nil {
		// nilの場合、ダミーのエラーを作成
		cause = errors.NewWithDepthf(1, "code:%s, args:%v", errorCode, args)
	} else {
		cause = errors.WithStackDepth(cause, 1)
	}
	// 誤ったエラーのラップを確認
	requiredNotCodableError(cause)
	return &OtherError{cause: cause, errorCode: errorCode, args: args}
}

// Error は、エラーを返却します。errorインタフェースを実装します。
func (e *OtherError) Error() string {
	return fmt.Sprintf("その他のエラー[%s]", e.errorCode)
}

// Format は、%+vを正しく動作させるため、fmt.Formatterのインタフェースを実装します。
// https://github.com/cockroachdb/errors#Making-v-work-with-your-type
func (e *OtherError) Format(s fmt.State, verb rune) {
	errors.FormatError(e, s, verb)
}

// Unwrap は、原因となるエラーにUnwrapします。
// https://github.com/cockroachdb/errors#building-your-own-error-types
func (e *OtherError) Unwrap() error {
	return e.cause
}

// ErrorCode implements CodableError.
func (e *OtherError) ErrorCode() string {
	return e.errorCode
}

// Args implements CodableError.
func (e *OtherError) Args() []any {
	return e.args
}
