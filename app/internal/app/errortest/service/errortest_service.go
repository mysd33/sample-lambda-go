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
		// 業務エラー（Causeありの場合）
		cause := fmt.Errorf("原因のエラーA")
		return errors.NewBusinessErrorWithCause(cause, message.W_EX_8001, "hoge")
	case "business2":
		// 業務エラー（Causeなしの場合）
		return errors.NewBusinessError(message.W_EX_8001, "fuga")
	case "business3":
		// 業務エラー（Causeがnilの場合）
		return errors.NewBusinessErrorWithCause(nil, message.W_EX_8001, "foo")
	case "business4":
		// 複数業務エラーの場合
		err1 := errors.NewBusinessError(message.W_EX_8001, "fuga")
		// 付加情報も付与できる
		err1.AddInfo("info1", "hoge")
		err1.AddInfo("info2", "foo")
		err2 := errors.NewBusinessError(message.W_EX_8001, "piyo")
		err2.AddInfo("info1", "bar")
		return errors.NewBusinessErrors(err1, err2)

	case "system":
		// システムエラー（Causeアリの場合）
		cause := fmt.Errorf("原因のエラーB")
		return errors.NewSystemError(cause, message.E_EX_9002, "foo")
	case "system2":
		return errors.NewSystemError(nil, message.E_EX_9002, "bar")
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
