// repositoryのパッケージ
package repository

import (
	"context"

	"example.com/appbase/pkg/apcontext"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-xray-sdk-go/instrumentation/awsv2"
	"github.com/cockroachdb/errors"
)

type JobRepository interface {
	// TODO: シグニチャは今後構造体に変更
	// メッセージを送信する
	Send(message string) error
}

func NewJobRepository() (JobRepository, error) {
	// TODO: Configからキュー名取得する
	queueName := "SampleQueue"

	// TODO: フレームワークに切り出す
	// TODO: カスタムHTTPClientの作成
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, errors.WithStack(err)
	}
	awsv2.AWSV2Instrumentor(&cfg.APIOptions)
	// TODO:　ローカル実行の考慮
	sqlClient := sqs.NewFromConfig(cfg)
	queueUrlOutput, err := sqlClient.GetQueueUrl(context.TODO(), &sqs.GetQueueUrlInput{
		QueueName: aws.String(queueName),
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &defaultJobRepository{
		queueUrl:  *queueUrlOutput.QueueUrl,
		sqsClient: sqlClient,
	}, nil
}

type defaultJobRepository struct {
	queueUrl  string
	sqsClient *sqs.Client
}

// Send implements JobRepository.
func (jr *defaultJobRepository) Send(message string) error {
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
