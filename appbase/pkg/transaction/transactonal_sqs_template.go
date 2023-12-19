/*
transaction パッケージは、トランザクション管理に関する機能を提供するパッケージです。
*/
package transaction

import (
	"encoding/json"

	"example.com/appbase/pkg/async"
	"example.com/appbase/pkg/id"
	"example.com/appbase/pkg/logging"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/cockroachdb/errors"
)

// NewSQSTemplateは、SQSTemplateを作成します。
func NewSQSTemplate(log logging.Logger, sqsAccessor TransactionalSQSAccessor) async.SQSTemplate {
	return &defaultTransactionalSQSTemplate{
		log:         log,
		sqsAccessor: sqsAccessor,
	}
}

// defaultTransactionalSQSTemplateは、SQSTemplateの実装です。
type defaultTransactionalSQSTemplate struct {
	log         logging.Logger
	sqsAccessor TransactionalSQSAccessor
}

// SendToStandardQueue implements async.SQSTemplate.
func (t *defaultTransactionalSQSTemplate) SendToStandardQueue(queueName string, msg any) error {

	// 構造体をjson文字列としてメッセージ送信
	byteMessage, err := json.Marshal(msg)
	if err != nil {
		return errors.WithStack(err)
	}
	input := &sqs.SendMessageInput{
		MessageBody: aws.String(string(byteMessage)),
	}
	// トランザクション管理して非同期実行依頼メッセージを追加
	err = t.sqsAccessor.AppendTransactMessage(queueName, input)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// SendToFIFOQueue implements async.SQSTemplate.
func (t *defaultTransactionalSQSTemplate) SendToFIFOQueue(queueName string, msg any, msgGroupId string) error {
	// 構造体をjson文字列としてメッセージ送信
	byteMessage, err := json.Marshal(msg)
	if err != nil {
		return errors.WithStack(err)
	}
	// メッセージ重複排除IDの作成
	msgDeduplicationId := id.GenerateId()
	input := &sqs.SendMessageInput{
		MessageBody:            aws.String(string(byteMessage)),
		MessageGroupId:         aws.String(msgGroupId),
		MessageDeduplicationId: aws.String(msgDeduplicationId),
	}
	// トランザクション管理して非同期実行依頼メッセージを追加
	err = t.sqsAccessor.AppendTransactMessage(queueName, input)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
