// repositoryのパッケージ
package repository

import (
	"app/internal/pkg/message"

	"example.com/appbase/pkg/async"
	"example.com/appbase/pkg/errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type AsyncMessageRepository interface {
	// TODO: シグニチャは今後構造体に変更
	// メッセージを送信する
	Send(msg string) error
}

func NewAsyncMessageRepository(sqsAccessor async.SQSAccessor) AsyncMessageRepository {
	return &defaultAsyncMessageRepository{
		sqsAccessor: sqsAccessor,
	}
}

type defaultAsyncMessageRepository struct {
	sqsAccessor async.SQSAccessor
}

// Send implements AsyncMessageRepository.
func (jr *defaultAsyncMessageRepository) Send(msg string) error {
	// TODO: 構造体で受け取ってjson化してから送信する処理に変更
	input := &sqs.SendMessageInput{
		MessageBody: aws.String(msg),
	}
	_, err := jr.sqsAccessor.SendMessageSdk(input)
	if err != nil {
		return errors.NewSystemError(err, message.E_EX_9001)
	}

	return nil
}
