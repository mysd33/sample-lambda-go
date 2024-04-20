/*
erros パッケージは、エラー情報を扱うパッケージです。
*/
package errors

import "github.com/cockroachdb/errors"

// Isは、Go標準のerrors.Is関数と同じで、
// errのラップ階層中のエラーがtargatで指定されたエラーと一致するかどうかを返します。
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// Asは、Go標準のerrors.As関数と同じで
// errのラップ階層中のエラーがtargetで指定されたエラーの型と一致するかどうかを返します。
// また、一致した場合は、targetに一致したエラーを設定します。
func As(err error, target any) bool {
	return errors.As(err, target)
}
