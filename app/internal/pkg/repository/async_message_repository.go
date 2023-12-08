// repositoryのパッケージ
package repository

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

type AsyncMessageRepository interface {
	// TODO: シグニチャは今後構造体に変更
	// メッセージを送信する
	Send(message string) error
}

func NewAsyncMessageRepository(myCfg myConfig.Config) (AsyncMessageRepository, error) {
	// TODO: Configからキュー名取得する
	queueName := "SampleQueue"

	// TODO: フレームワークに切り出す
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
	return &defaultAsyncMessageRepository{
		queueUrl:  *queueUrlOutput.QueueUrl,
		sqsClient: sqlClient,
	}, nil
}

type defaultAsyncMessageRepository struct {
	queueUrl  string
	sqsClient *sqs.Client
}

// Send implements AsyncMessageRepository.
func (jr *defaultAsyncMessageRepository) Send(message string) error {
	// TODO: フレームワークに切り出す
	// TODO: オブジェクトで受け取ってjson化してから送信する処理に変更

	input := &sqs.SendMessageInput{
		MessageBody: aws.String(message),
		QueueUrl:    aws.String(jr.queueUrl),
	}

	// TODO: Ouputの利用
	_, err := jr.sqsClient.SendMessage(apcontext.Context, input)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
