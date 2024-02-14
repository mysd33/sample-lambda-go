/*
transaction パッケージは、トランザクション管理に関する機能を提供するパッケージです。
*/
package transaction

import (
	"context"
	"encoding/json"

	"example.com/appbase/pkg/async"
	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/id"
	"example.com/appbase/pkg/logging"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/cockroachdb/errors"
)

const (
	STANDARD_QUEUE_DELAY_SECONDS = "STANDARD_QUEUE_DELAY_SECONDS"
)

// NewSQSTemplateは、SQSTemplateを作成します。
func NewSQSTemplate(log logging.Logger, config config.Config, id id.IDGenerator,
	sqsAccessor TransactionalSQSAccessor) async.SQSTemplate {
	return &defaultTransactionalSQSTemplate{
		log:         log,
		config:      config,
		id:          id,
		sqsAccessor: sqsAccessor,
	}
}

// defaultTransactionalSQSTemplateは、SQSTemplateの実装です。
type defaultTransactionalSQSTemplate struct {
	log         logging.Logger
	config      config.Config
	id          id.IDGenerator
	sqsAccessor TransactionalSQSAccessor
}

// SendToStandardQueue implements async.SQSTemplate.
func (t *defaultTransactionalSQSTemplate) SendToStandardQueue(queueName string, msg any) error {
	input, err := t.newSendMessageInputToStandardQueue(msg)
	if err != nil {
		return err
	}
	// トランザクション管理して非同期実行依頼メッセージを追加
	return t.sqsAccessor.AppendTransactMessage(queueName, input)
}

// SendToStandardQueueWithContext implements async.SQSTemplate.
func (t *defaultTransactionalSQSTemplate) SendToStandardQueueWithContext(ctx context.Context, queueName string, msg any) error {
	input, err := t.newSendMessageInputToStandardQueue(msg)
	if err != nil {
		return err
	}
	// トランザクション管理して非同期実行依頼メッセージを追加
	return t.sqsAccessor.AppendTransactMessageWithContext(ctx, queueName, input)
}

// newSendMessageInputToStandardQueueは、標準キューへ送信するためのSendMessageInputを作成します。
func (t *defaultTransactionalSQSTemplate) newSendMessageInputToStandardQueue(msg any) (*sqs.SendMessageInput, error) {
	// 構造体をjson文字列としてメッセージ送信
	byteMessage, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var input *sqs.SendMessageInput
	// DelaySecondsが設定されていれば上書き
	delaySeconds, found := t.config.GetIntWithContains(STANDARD_QUEUE_DELAY_SECONDS)
	if !found {
		input = &sqs.SendMessageInput{
			MessageBody: aws.String(string(byteMessage)),
		}
	} else {
		input = &sqs.SendMessageInput{
			MessageBody:  aws.String(string(byteMessage)),
			DelaySeconds: int32(delaySeconds),
		}
	}
	return input, nil
}

// SendToFIFOQueue implements async.SQSTemplate.
func (t *defaultTransactionalSQSTemplate) SendToFIFOQueue(queueName string, msg any, msgGroupId string) error {
	input, err := t.newSendMessageInputToFIFOQueue(msg, msgGroupId)
	if err != nil {
		return err
	}
	// トランザクション管理して非同期実行依頼メッセージを追加
	return t.sqsAccessor.AppendTransactMessage(queueName, input)
}

// SendToFIFOQueueWithContext implements async.SQSTemplate.
func (t *defaultTransactionalSQSTemplate) SendToFIFOQueueWithContext(ctx context.Context, queueName string, msg any, msgGroupId string) error {
	input, err := t.newSendMessageInputToFIFOQueue(msg, msgGroupId)
	if err != nil {
		return err
	}
	// トランザクション管理して非同期実行依頼メッセージを追加
	return t.sqsAccessor.AppendTransactMessageWithContext(ctx, queueName, input)
}

func (t *defaultTransactionalSQSTemplate) newSendMessageInputToFIFOQueue(msg any, msgGroupId string) (*sqs.SendMessageInput, error) {
	// 構造体をjson文字列としてメッセージ送信
	byteMessage, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	// メッセージ重複排除IDの作成
	msgDeduplicationId := t.id.GenerateUUID()
	input := &sqs.SendMessageInput{
		MessageBody:            aws.String(string(byteMessage)),
		MessageGroupId:         aws.String(msgGroupId),
		MessageDeduplicationId: aws.String(msgDeduplicationId),
	}
	return input, nil
}
