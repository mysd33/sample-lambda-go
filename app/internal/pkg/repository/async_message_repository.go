// repositoryのパッケージ
package repository

import (
	"app/internal/pkg/entity"
	"app/internal/pkg/message"

	"example.com/appbase/pkg/async"
	"example.com/appbase/pkg/errors"
)

type AsyncMessageRepository interface {
	// メッセージを（標準キューに）送信する
	Send(msg *entity.AsyncMessage) error
	// メッセージをFIFOに送信する
	SendToFIFOQueue(msg *entity.AsyncMessage, msgGroupId string) error
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
func (r *defaultAsyncMessageRepository) Send(msg *entity.AsyncMessage) error {
	// 標準キューでの非同期実行依頼
	if err := r.sqsTemplate.SendToStandardQueue(r.starndardQueueName, msg); err != nil {
		return errors.NewSystemError(err, message.E_EX_9001)
	}
	return nil
}

// SendToFIFOQueue implements AsyncMessageRepository.
func (r *defaultAsyncMessageRepository) SendToFIFOQueue(msg *entity.AsyncMessage, msgGroupId string) error {
	// FIFOキューでの非同期実行依頼
	if err := r.sqsTemplate.SendToFIFOQueue(r.fifoQueueName, msg, msgGroupId); err != nil {
		return errors.NewSystemError(err, message.E_EX_9001)
	}
	return nil
}
