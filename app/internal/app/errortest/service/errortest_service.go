// serviceのパッケージ
package service

import (
	"app/internal/pkg/message"
	"fmt"

	"example.com/appbase/pkg/errors"
)

// ErrorTestService は、エラーをテストするためのServiceインタフェースです。
type ErrorTestService interface {
	// Executeは、errorTypeで指定されたダミーのエラーを返却します。
	Execute(errorType string) error
}

// New は、ErrorTestServiceを作成します。
func New() ErrorTestService {
	return &errorTestServiceImpl{}
}

// errorTestServiceImpl は、ErrorTestServiceを実装する構造体です。
type errorTestServiceImpl struct {
}

// Execute implements ErrorTestService.
func (*errorTestServiceImpl) Execute(errorType string) error {
	// 指定されたエラーの種類によって、Errorを返却する
	switch errorType {
	case "business":
		// Causeありの場合
		cause := fmt.Errorf("原因のエラーA")
		return errors.NewBusinessErrorWithCause(cause, message.W_EX_8001, "hoge")
	case "business2":
		// Causeなしの場合
		return errors.NewBusinessError(message.W_EX_8001, "fuga")
	case "system":
		//システムエラー
		cause := fmt.Errorf("原因のエラーB")
		return errors.NewSystemError(cause, message.E_EX_9002, "foo")
	case "panic":
		panic("パニック発生")
	default:
		//予期せぬエラー（実際には、AP側でのSystemErrorのラップし忘れ）
		//スタックトレースなし
		return fmt.Errorf("予期せぬエラー[%s]", errorType)
		//スタックトレースあり
		//return cerrors.Errorf("予期せぬエラー[%s]", errorType)
	}
}
