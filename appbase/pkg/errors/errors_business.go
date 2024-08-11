/*
erros パッケージは、エラー情報を扱うパッケージです。
*/
package errors

import (
	"fmt"
	"unsafe"

	"github.com/cockroachdb/errors"
)

// BusinessErrors 業務エラーを複数保持する構造体です。
type BusinessErrors struct {
	errs []*BusinessError
	info any
}

// NewBusinessErrors は、BusinessErrors構造体を作成します。
func NewBusinessErrors(errors ...*BusinessError) *BusinessErrors {
	return &BusinessErrors{errs: errors}
}

// WithInfo は、付加情報を設定します。
func (e *BusinessErrors) WithInfo(info any) *BusinessErrors {
	e.info = info
	return e
}

// Info は、付加情報を返します。
func (e *BusinessErrors) Info() any {
	return e.info
}

// BusinessErrors は、保持している業務エラーを返します。
func (e *BusinessErrors) BusinessErrors() []*BusinessError {
	return e.errs
}

// Errors implements MultiCodableError.
func (e *BusinessErrors) CodableErrors() []CodableError {
	errs := make([]CodableError, len(e.errs))
	for i, err := range e.errs {
		errs[i] = err
	}
	return errs
}

// Error は、エラーを返却します。errorインタフェースを実装します。
func (e *BusinessErrors) Error() string {
	if len(e.errs) == 0 {
		return ""
	}
	if len(e.errs) == 1 {
		return e.errs[0].Error()
	}
	b := []byte(e.errs[0].Error())
	for _, err := range e.errs[1:] {
		b = append(b, '\n')
		b = append(b, err.Error()...)
	}
	return unsafe.String(&b[0], len(b))
}

// Unwrap は、原因となるエラーにUnwrapします。
// https://pkg.go.dev/errors
func (e *BusinessErrors) Unwrap() []error {
	errs := make([]error, len(e.errs))
	for i, err := range e.errs {
		errs[i] = err
	}
	return errs
}

// Format は、%+vを正しく動作させるため、fmt.Formatterのインタフェースを実装します。
// https://github.com/cockroachdb/errors#Making-v-work-with-your-type
func (e *BusinessErrors) Format(s fmt.State, verb rune) { errors.FormatError(e, s, verb) }

// BusinessError 業務エラーの構造体です。
type BusinessError struct {
	cause     error
	errorCode string
	args      []any
	infos     map[string]any
}

// NewBusinessError は、メッセージIDにもなるエラーコード（errorCode）とメッセージの置換文字列(args）を渡し
// BusinessError構造体を作成します。
func NewBusinessError(errorCode string, args ...any) *BusinessError {
	// ダミーのエラーを作成
	cause := errors.NewWithDepthf(1, "code:%s, args:%v", errorCode, args)
	// スタックトレース出力のため、cockloachdb/errorのスタックトレース付きのcauseエラー作成
	return &BusinessError{cause: cause, errorCode: errorCode, args: args}
}

// NewBusinessError は、原因となるエラー（cause）をラップし、
// メッセージIDにもなるエラーコード（errorCode）とメッセージの置換文字列(args）を渡し
// BusinessError構造体を作成します。
func NewBusinessErrorWithCause(cause error, errorCode string, args ...any) *BusinessError {
	if cause == nil {
		// nilの場合、ダミーのエラーを作成
		cause = errors.NewWithDepthf(1, "code:%s, args:%v", errorCode, args)
	} else {
		cause = errors.WithStackDepth(cause, 1)
	}
	// 誤ったエラーのラップを確認
	requiredNotCodableError(cause)
	// causeはスタックトレース付与
	return &BusinessError{cause: cause, errorCode: errorCode, args: args}
}

// Error は、エラーを返却します。errorインタフェースを実装します。
func (e *BusinessError) Error() string {
	return fmt.Sprintf("業務エラー[%s]", e.errorCode)
}

// Format は、%+vを正しく動作させるため、fmt.Formatterのインタフェースを実装します。
// https://github.com/cockroachdb/errors#Making-v-work-with-your-type
func (e *BusinessError) Format(s fmt.State, verb rune) { errors.FormatError(e, s, verb) }

// Unwrap は、原因となるエラーにUnwrapします。
// https://github.com/cockroachdb/errors#building-your-own-error-types
func (e *BusinessError) Unwrap() error {
	return e.cause
}

// ErrorCode implements CodableError.
func (e *BusinessError) ErrorCode() string {
	return e.errorCode
}

// Args implements CodableError.
func (e *BusinessError) Args() []any {
	return e.args
}

// AddInfo は、付加情報を追加します。
func (e *BusinessError) AddInfo(key string, value any) {
	if e.infos == nil {
		e.infos = make(map[string]any)
	}
	e.infos[key] = value
}

// Info は、指定したkeyに対応する付加情報を返します。
func (e *BusinessError) Info(key string) any {
	return e.infos[key]
}

// As は、errors.As関数で使用されるインタフェースを実装します。
// BusinessErrorは、BusinessErrorsに変換可能です。
// targetがBusinessErrorsの場合は、BusinessErrorsに自身を追加したものをtargetに設定します。
func (e *BusinessError) As(target any) bool {
	if t, ok := target.(**BusinessErrors); ok {
		*t = NewBusinessErrors(e)
		return true
	}
	if t, ok := target.(**BusinessError); ok {
		*t = e
		return true
	}
	return false
}
