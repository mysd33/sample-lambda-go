/*
async パッケージは、非同期実行依頼の機能を提供します。
*/
package async

import "github.com/aws/aws-sdk-go-v2/service/sqs"

// SQSAccessor は、AWS SDKを使ったSQSアクセスの実装をラップしカプセル化するインタフェースです。
type SQSAccessor interface {
	// SendMessageSdk は、AWS SDKによるSendMessageをラップします。
	SendMessageSdk(queueName string, input *sqs.SendMessageInput) (*sqs.SendMessageOutput, error)
}
