/*
erros パッケージは、エラー情報を扱うパッケージです。
*/
package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"unsafe"

	"example.com/appbase/pkg/env"
	myvalidator "example.com/appbase/pkg/validator"
	cerrors "github.com/cockroachdb/errors"
	"github.com/go-playground/validator/v10"
)

// CodableErrorは、エラーコード定義付きのエラーインタフェースです。
type CodableError interface {
	error
	// ErrorCode は、エラーコード（メッセージID）を返します。
	ErrorCode() string
	// Argsは、エラーメッセージの置換文字列(args）を返します
	Args() []any
}

// MultiCodableError は、複数のCodableErrorを保持するインタフェースです。
type MultiCodableError interface {
	error
	CodableErrors() []CodableError
}

// ValidationError は、入力エラーの構造体です。
type ValidationError struct {
	cause     error
	errorCode string
	args      []any
}

// NewValidationError は、ッセージIDにもなるエラーコード（errorCode）とメッセージの置換文字列(args）を渡し
// ValidationError構造体を作成します。
func NewValidationError(errorCode string, args ...any) *ValidationError {
	// スタックトレース出力のため、cockloachdb/errorのスタックトレース付きのcauseエラー作成
	cause := cerrors.NewWithDepthf(1, "code:%s, error:%v", errorCode, args)
	return &ValidationError{cause: cause, errorCode: errorCode, args: args}
}

// NewValidationErrorWithCause は、原因となるエラー（cause）をラップし、
// メッセージIDにもなるエラーコード（errorCode）とメッセージの置換文字列(args）を渡しValidationError構造体を作成します。
func NewValidationErrorWithCause(cause error, errorCode string, args ...any) *ValidationError {
	// 誤ったエラーのラップを確認
	requiredNotCodableError(cause)
	return &ValidationError{cause: cerrors.WithStackDepth(cause, 1), errorCode: errorCode, args: args}
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

// Args implements CodableError.
func (e *ValidationError) Args() []any {
	return e.args
}

// ErrorCode implements CodableError.
func (e *ValidationError) ErrorCode() string {
	return e.errorCode
}

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
func (e *BusinessErrors) Format(s fmt.State, verb rune) { cerrors.FormatError(e, s, verb) }

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
	// スタックトレース出力のため、cockloachdb/errorのスタックトレース付きのcauseエラー作成
	cause := cerrors.NewWithDepthf(1, "code:%s, error:%v", errorCode, args)
	return &BusinessError{cause: cause, errorCode: errorCode, args: args}
}

// NewBusinessError は、原因となるエラー（cause）をラップし、
// メッセージIDにもなるエラーコード（errorCode）とメッセージの置換文字列(args）を渡し
// BusinessError構造体を作成します。
func NewBusinessErrorWithCause(cause error, errorCode string, args ...any) *BusinessError {
	// 誤ったエラーのラップを確認
	requiredNotCodableError(cause)
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
	// TODO: causeがnilでないことのチェック
	// 誤ったエラーのラップを確認
	requiredNotCodableError(cause)
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

// ErrorCode implements CodableError.
func (e *SystemError) ErrorCode() string {
	return e.errorCode
}

// Args implements CodableError.
func (e *SystemError) Args() []any {
	return e.args
}

// Causeが、ValidationError、BusinessError、SystemErrorでないことを確認
func requiredNotCodableError(cause error) {
	if cause == nil {
		return
	}
	var ve *ValidationError
	var be *BusinessError
	var se *SystemError
	// causeが、ValidationError、BusinessError、SystemErrorの場合は、
	// コーディングミスで二重でラップしてしまっている判断して、開発中は異常終了させている
	if errors.As(cause, &ve) {
		if !env.IsProd() {
			// 異常終了
			panic(fmt.Sprintf("誤ってValidationErrorを二重でラップしています:%+v", ve))
		}
		log.Printf("誤ってValidationErrorを二重でラップしています:%+v", ve)
	} else if errors.As(cause, &be) {
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
