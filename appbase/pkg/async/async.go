/*
async パッケージは、非同期実行依頼の機能を提供します。
*/
package async

import (
	"context"

	"example.com/appbase/pkg/apcontext"
	"example.com/appbase/pkg/awssdk"
	myConfig "example.com/appbase/pkg/config"
	"example.com/appbase/pkg/logging"
	"example.com/appbase/pkg/message"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-xray-sdk-go/instrumentation/awsv2"
	"github.com/cockroachdb/errors"
)

const (
	SQS_LOCAL_ENDPOINT_NAME = "SQS_LOCAL_ENDPOINT"
)

// SQSTemplate は、SQSにメッセージを送信するための高次のインタフェースです。
type SQSTemplate interface {
	// SendToStandardQueue は、標準のキューにメッセージを送信します。
	SendToStandardQueue(queueName string, msg any) error
	// SendToStandardQueueWithContext は、goroutine向けに渡されたContextを利用して、標準のキューにメッセージを送信します。
	SendToStandardQueueWithContext(ctx context.Context, queueName string, msg any) error
	// SendToFIFOQueue は、FIFOキューにメッセージを送信します。
	// SQS側の設定でコンテンツに基づく重複排除を設定すると重複排除IDは自動で設定されます。
	SendToFIFOQueue(queueName string, msg any, msgGroupId string) error
	// SendToFIFOQueueRandomDedupId は、ランダム値による重複排除IDを生成し指定してFIFOキューにメッセージを送信します。
	// SQS側の設定でコンテンツに基づく重複排除を設定しない場合などに使用します。
	SendToFIFOQueueRandomDedupId(queueName string, msg any, msgGroupId string) error
	// SendToFIFOQueueWithContext は、goroutine向けに渡されたContextを利用して、FIFOキューにメッセージを送信します。
	// SQS側の設定でコンテンツに基づく重複排除を設定すると重複排除IDは自動で設定されます。
	SendToFIFOQueueWithContext(ctx context.Context, queueName string, msg any, msgGroupId string) error
	// SendToFIFOQueueRandomDedupIdWithContext は、goroutine向けに渡されたContextを利用して、
	// ランダム値による重複排除IDを生成し指定してFIFOキューにメッセージを送信します。
	// SQS側の設定でコンテンツに基づく重複排除を設定しない場合などに使用します。
	SendToFIFOQueueRandomDedupIdWithContext(ctx context.Context, queueName string, msg any, msgGroupId string) error
}

// SQSAccessor は、AWS SDKを使ったSQSアクセスの実装をラップしカプセル化する低次のインタフェースです。
type SQSAccessor interface {
	// SendMessageSdk は、AWS SDKによるSendMessageをラップします。
	SendMessageSdk(queueName string, input *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
	// SendMessageSdkWithContext は、AWS SDKによるSendMessageをラップします。goroutine向けに、渡されたContextを利用して実行します。
	SendMessageSdkWithContext(ctx context.Context, queueName string, input *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
}

// NewSQSAccessor は、SQSAccessorを作成します。
func NewSQSAccessor(logger logging.Logger, myCfg myConfig.Config) (SQSAccessor, error) {
	// カスタムHTTPClientの作成
	sdkHTTPClient := awssdk.NewHTTPClient(myCfg)
	// ClientLogModeの取得
	clientLogMode, found := awssdk.GetClientLogMode(myCfg)
	// AWS SDK for Go v2 Migration
	// https://github.com/aws/aws-sdk-go-v2
	// https://aws.github.io/aws-sdk-go-v2/docs/migrating/
	var cfg aws.Config
	var err error
	if found {
		cfg, err = config.LoadDefaultConfig(context.TODO(), config.WithHTTPClient(sdkHTTPClient), config.WithClientLogMode(clientLogMode))
	} else {
		cfg, err = config.LoadDefaultConfig(context.TODO(), config.WithHTTPClient(sdkHTTPClient))
	}
	if err != nil {
		return nil, errors.WithStack(err)
	}
	awsv2.AWSV2Instrumentor(&cfg.APIOptions)
	sqlClient := sqs.NewFromConfig(cfg, func(o *sqs.Options) {
		// ローカル実行のためDynamoDB Local起動先が指定されている場合
		sqsEndpoint := myCfg.Get(SQS_LOCAL_ENDPOINT_NAME, "")
		if sqsEndpoint != "" {
			o.BaseEndpoint = aws.String(sqsEndpoint)
		}
	})
	return &defaultSQSAccessor{
		config:    myCfg,
		logger:    logger,
		sqsClient: sqlClient,
		queueUrls: make(map[string]string),
	}, nil
}

// defaultSQSAccessor は、SQSAccessorを実装する構造体です。
type defaultSQSAccessor struct {
	config    myConfig.Config
	logger    logging.Logger
	sqsClient *sqs.Client
	queueUrls map[string]string
}

// SendMessageSdk implements SQSAccessor.
func (sa *defaultSQSAccessor) SendMessageSdk(queueName string, input *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
	return sa.SendMessageSdkWithContext(apcontext.Context, queueName, input, optFns...)
}

// SendMessageSdkWithContext implements SQSAccessor.
func (sa *defaultSQSAccessor) SendMessageSdkWithContext(ctx context.Context, queueName string, input *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
	if ctx == nil {
		ctx = apcontext.Context
	}
	// QueueのURLの取得・設定
	queueUrl, ok := sa.queueUrls[queueName]
	if ok {
		// キャッシュがある場合は、キャッシュから取得
		sa.logger.Debug("QueueURLキャッシュ:%s", queueUrl)
		input.QueueUrl = &queueUrl
	} else {
		// キャッシュがない場合は、APIで取得
		queueUrlOutput, err := sa.sqsClient.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
			QueueName: aws.String(queueName),
		}, optFns...)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		sa.logger.Debug("GetQueueURL:%s", *queueUrlOutput.QueueUrl)
		// 送信先の設定
		input.QueueUrl = queueUrlOutput.QueueUrl
		// キャッシュへ格納
		sa.queueUrls[queueName] = *queueUrlOutput.QueueUrl
	}
	// ログ出力
	if input.MessageGroupId != nil {
		if input.MessageDeduplicationId != nil {
			sa.logger.Info(message.I_FW_0006, queueName, *input.MessageGroupId, *input.MessageDeduplicationId)
			sa.logger.Debug("MessageGroupId=%s, MessageDeduplicationId=%s, Message=%s", *input.MessageGroupId, *input.MessageDeduplicationId, *input.MessageBody)
		} else {
			sa.logger.Info(message.I_FW_0006, queueName, *input.MessageGroupId, "N/A")
			sa.logger.Debug("MessageGroupId=%s, Message=%s", *input.MessageGroupId, *input.MessageBody)
		}
	} else {
		sa.logger.Info(message.I_FW_0006, queueName, "N/A", "N/A")
		sa.logger.Debug("Message=%s", *input.MessageBody)
	}

	//　SQSへメッセージ送信する
	output, err := sa.sqsClient.SendMessage(ctx, input, optFns...)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// ログ出力
	if input.MessageGroupId != nil {
		if input.MessageDeduplicationId != nil {
			sa.logger.Info(message.I_FW_0007, queueName, *output.MessageId, *input.MessageGroupId, *input.MessageDeduplicationId)
		} else {
			sa.logger.Info(message.I_FW_0007, queueName, *output.MessageId, *input.MessageGroupId, "N/A")
		}
	} else {
		sa.logger.Info(message.I_FW_0007, queueName, *output.MessageId, "N/A", "N/A")
	}
	return output, nil
}
