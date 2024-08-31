/*
erros パッケージは、エラー情報を扱うパッケージです。
*/
package errors

import (
	"fmt"
	"log"

	"example.com/appbase/pkg/env"
	"github.com/cockroachdb/errors"
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

// Causeが、ValidationError、BusinessError、SystemErrorでないことを確認
func requiredNotCodableError(cause error) {
	if cause == nil {
		return
	}
	var ve *ValidationError
	var be *BusinessError
	var bes *BusinessErrors
	var se *SystemError
	var oe *OtherError
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
	} else if errors.As(cause, &bes) {
		if !env.IsProd() {
			// 異常終了
			panic(fmt.Sprintf("誤ってBusinessErrorsを二重でラップしています:%+v", bes))
		}
		log.Printf("誤ってBusinessErrorsを二重でラップしています:%+v", bes)
	} else if errors.As(cause, &se) {
		if !env.IsProd() {
			// 異常終了
			panic(fmt.Sprintf("誤ってSystemErrorを二重でラップしています:%+v", se))
		}
		log.Printf("誤ってSystemErrorを二重でラップしています:%+v", se)
	} else if errors.As(cause, &oe) {
		if !env.IsProd() {
			// 異常終了
			panic(fmt.Sprintf("誤ってOtherErrorを二重でラップしています:%+v", oe))
		}
		log.Printf("誤ってOtherErrorを二重でラップしています:%+v", oe)
	}
}
