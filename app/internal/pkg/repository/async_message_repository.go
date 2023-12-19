// repositoryのパッケージ
package repository

import (
	"app/internal/pkg/message"
	"encoding/json"

	"example.com/appbase/pkg/errors"
	"example.com/appbase/pkg/id"
	"example.com/appbase/pkg/transaction"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type AsyncMessageRepository interface {
	// メッセージを送信する
	Send(msg any) error
	SendToFIFOQueue(msg any, msgGroupId string) error
}

func NewAsyncMessageRepository(sqsAccessor transaction.TransactionalSQSAccessor, stadardQueueName string, fifoQueueName string) AsyncMessageRepository {
	return &defaultAsyncMessageRepository{
		sqsAccessor:        sqsAccessor,
		starndardQueueName: stadardQueueName,
		fifoQueueName:      fifoQueueName,
	}
}

type defaultAsyncMessageRepository struct {
	sqsAccessor        transaction.TransactionalSQSAccessor
	starndardQueueName string
	fifoQueueName      string
}

// Send implements AsyncMessageRepository.
func (r *defaultAsyncMessageRepository) Send(msg any) error {
	// 構造体をjson文字列としてメッセージ送信
	byteMessage, err := json.Marshal(msg)
	if err != nil {
		return errors.NewSystemError(err, message.E_EX_9001)
	}
	input := &sqs.SendMessageInput{
		MessageBody: aws.String(string(byteMessage)),
	}
	// トランザクション管理して非同期実行依頼メッセージを追加
	err = r.sqsAccessor.AppendTransactMessage(r.starndardQueueName, input)
	if err != nil {
		return errors.NewSystemError(err, message.E_EX_9001)
	}

	return nil
}

// SendToFIFOQueue implements AsyncMessageRepository.
func (r *defaultAsyncMessageRepository) SendToFIFOQueue(msg any, msgGroupId string) error {
	// 構造体をjson文字列としてメッセージ送信
	byteMessage, err := json.Marshal(msg)
	if err != nil {
		return errors.NewSystemError(err, message.E_EX_9001)
	}
	// メッセージ重複排除IDの作成
	msgDeduplicationId := id.GenerateId()
	input := &sqs.SendMessageInput{
		MessageBody:            aws.String(string(byteMessage)),
		MessageGroupId:         aws.String(msgGroupId),
		MessageDeduplicationId: aws.String(msgDeduplicationId),
	}
	// トランザクション管理して非同期実行依頼メッセージを追加
	err = r.sqsAccessor.AppendTransactMessage(r.fifoQueueName, input)
	if err != nil {
		return errors.NewSystemError(err, message.E_EX_9001)
	}

	return nil
}
