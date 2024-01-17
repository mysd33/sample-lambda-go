/*
api パッケージは、REST APIに関する機能を提供するパッケージです。
*/
package api

import "example.com/appbase/pkg/errors"

// ErrorResponse は、エラーレスポンスを作成するインタフェースです。
type ErrorResponse interface {
	// ValidationErrorResponse は、バリデーションエラーに対応するレスポンスを返却します。
	ValidationErrorResponse(validationError *errors.ValidationError) (int, any)
	// BusinessErrorResponse は、業務エラーに対応するレスポンスを返却します。
	BusinessErrorResponse(businessErrors *errors.BusinessErrors) (int, any)
	// WarnErrorResponse は、警告エラーに対応するレスポンスを返却します。
	WarnErrorResponse(err error) (int, any)
	// SystemErrorResponse は、システムエラーに対応するレスポンスを返却します。
	SystemErrorResponse(systemError *errors.SystemError) (int, any)
	// UnExpectedErrorResponse は、予期せぬエラーに対応するレスポンスを返却します。
	UnExpectedErrorResponse(err error) (int, any)
}
