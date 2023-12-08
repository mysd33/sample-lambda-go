/*
async パッケージは、非同期実行依頼の機能を提供します。
*/
package async

import (
	"context"

	"example.com/appbase/pkg/apcontext"
	myConfig "example.com/appbase/pkg/config"
	"example.com/appbase/pkg/constant"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-xray-sdk-go/instrumentation/awsv2"
	"github.com/cockroachdb/errors"
)

//TODO: API検討中

// SQSAccessor は、AWS SDKを使ったSQSアクセスの実装をラップしカプセル化するインタフェースです。
type SQSAccessor interface {
	// SendMessageSdk は、AWS SDKによるSendMessageをラップします。
	SendMessageSdk(input *sqs.SendMessageInput) (*sqs.SendMessageOutput, error)
}

func NewSQSAccessor(myCfg myConfig.Config, queueName string) (SQSAccessor, error) {
	// TODO: カスタムHTTPClientの作成
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, errors.WithStack(err)
	}
	awsv2.AWSV2Instrumentor(&cfg.APIOptions)
	sqlClient := sqs.NewFromConfig(cfg, func(o *sqs.Options) {
		// ローカル実行のためDynamoDB Local起動先が指定されている場合
		sqsEndpoint := myCfg.Get(constant.SQS_LOCAL_ENDPOINT_NAME)
		if sqsEndpoint != "" {
			o.BaseEndpoint = aws.String(sqsEndpoint)
		}
	})
	queueUrlOutput, err := sqlClient.GetQueueUrl(context.TODO(), &sqs.GetQueueUrlInput{
		QueueName: aws.String(queueName),
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &defaultSQSAccessor{
		config:    myCfg,
		sqsClient: sqlClient,
		queueUrl:  *queueUrlOutput.QueueUrl,
	}, nil
}

type defaultSQSAccessor struct {
	config    myConfig.Config
	sqsClient *sqs.Client
	queueUrl  string
}

// TODO: DBとのデータ整合性を担保

// SendMessageSdk implements SQSAccessor.
func (sa *defaultSQSAccessor) SendMessageSdk(input *sqs.SendMessageInput) (*sqs.SendMessageOutput, error) {
	// TODO: Ouputの利用
	input.QueueUrl = &sa.queueUrl
	output, err := sa.sqsClient.SendMessage(apcontext.Context, input)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return output, nil
}
