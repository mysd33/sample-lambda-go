// repositoryのパッケージ
package repository

import (
	"app/internal/pkg/message"

	"example.com/appbase/pkg/async"
	"example.com/appbase/pkg/errors"
)

type AsyncMessageRepository interface {
	// メッセージを送信する
	Send(msg any) error
	SendToFIFOQueue(msg any, msgGroupId string) error
}

func NewAsyncMessageRepository(sqsTemplate async.SQSTemplate,
	stadardQueueName string,
	fifoQueueName string) AsyncMessageRepository {
	return &defaultAsyncMessageRepository{
		sqsTemplate:        sqsTemplate,
		starndardQueueName: stadardQueueName,
		fifoQueueName:      fifoQueueName,
	}
}

type defaultAsyncMessageRepository struct {
	sqsTemplate        async.SQSTemplate
	starndardQueueName string
	fifoQueueName      string
}

// Send implements AsyncMessageRepository.
func (r *defaultAsyncMessageRepository) Send(msg any) error {
	// 標準キューでの非同期実行依頼
	if err := r.sqsTemplate.SendToStandardQueue(r.starndardQueueName, msg); err != nil {
		return errors.NewSystemError(err, message.E_EX_9001)
	}
	return nil
}

// SendToFIFOQueue implements AsyncMessageRepository.
func (r *defaultAsyncMessageRepository) SendToFIFOQueue(msg any, msgGroupId string) error {
	// FIFOキューでの非同期実行依頼
	if err := r.sqsTemplate.SendToFIFOQueue(r.fifoQueueName, msg, msgGroupId); err != nil {
		return errors.NewSystemError(err, message.E_EX_9001)
	}
	return nil
}
