// repositoryのパッケージ
package repository

import (
	"app/internal/pkg/message"

	"example.com/appbase/pkg/async"
	"example.com/appbase/pkg/errors"
	"example.com/appbase/pkg/id"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type AsyncMessageRepository interface {
	// TODO: シグニチャは今後構造体に変更
	// メッセージを送信する
	Send(msg string) error
	SendToFIFOQueue(msg string, msgGroupId string) error
}

func NewAsyncMessageRepository(sqsAccessor async.SQSAccessor, stadardQueueName string, fifoQueueName string) AsyncMessageRepository {
	return &defaultAsyncMessageRepository{
		sqsAccessor:        sqsAccessor,
		starndardQueueName: stadardQueueName,
		fifoQueueName:      fifoQueueName,
	}
}

type defaultAsyncMessageRepository struct {
	sqsAccessor        async.SQSAccessor
	starndardQueueName string
	fifoQueueName      string
}

// Send implements AsyncMessageRepository.
func (r *defaultAsyncMessageRepository) Send(msg string) error {
	// TODO: 構造体で受け取ってjson化してから送信する処理に変更
	input := &sqs.SendMessageInput{
		MessageBody: aws.String(msg),
	}
	// TODO: Outputの扱い
	_, err := r.sqsAccessor.SendMessageSdk(r.starndardQueueName, input)
	if err != nil {
		return errors.NewSystemError(err, message.E_EX_9001)
	}

	return nil
}

// SendToFIFOQueue implements AsyncMessageRepository.
func (r *defaultAsyncMessageRepository) SendToFIFOQueue(msg string, msgGroupId string) error {
	// TODO: 構造体で受け取ってjson化してから送信する処理に変更
	// メッセージ重複排除IDの作成
	msgDeduplicationId := id.GenerateId()
	input := &sqs.SendMessageInput{
		MessageBody:            aws.String(msg),
		MessageGroupId:         aws.String(msgGroupId),
		MessageDeduplicationId: aws.String(msgDeduplicationId),
	}
	// TODO: Outputの扱い
	_, err := r.sqsAccessor.SendMessageSdk(r.fifoQueueName, input)
	if err != nil {
		return errors.NewSystemError(err, message.E_EX_9001)
	}

	return nil
}
