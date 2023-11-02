package service

import (
	"app/internal/pkg/message"
	"fmt"

	"example.com/appbase/pkg/errors"
	cerrors "github.com/cockroachdb/errors"
)

type ErrorTestService interface {
	Execute(errorType string) error
}

func New() ErrorTestService {
	return &errorTestServiceImpl{}
}

type errorTestServiceImpl struct {
}

// Execute implements ErrorTestService.
func (*errorTestServiceImpl) Execute(errorType string) error {
	// 指定されたエラーの種類によって、Errorを返却する
	switch errorType {
	case "business":
		cause := fmt.Errorf("原因のエラー")
		return errors.NewBusinessErrorWithCause(cause, message.W_EX_8001, "hoge")
	case "system":
		cause := fmt.Errorf("原因のエラー")
		return errors.NewSystemError(cause, message.E_EX_9002, "hoge")
	default:
		//return fmt.Errorf("予期せぬエラー[%s]", errorType)
		return cerrors.Errorf("予期せぬエラー[%s]", errorType)
	}
}
