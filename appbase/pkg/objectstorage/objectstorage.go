/*
objectstorage パッケージは、オブジェクトストレージを扱うためのパッケージです。
*/
package objectstorage

import (
	"bytes"
	"context"
	"io"
	"os"

	"example.com/appbase/pkg/apcontext"
	myconfig "example.com/appbase/pkg/config"
	"example.com/appbase/pkg/logging"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-xray-sdk-go/instrumentation/awsv2"
	"github.com/cockroachdb/errors"
)

const (
	S3_LOCAL_ENDPOINT_NAME = "S3_LOCAL_ENDPOINT"
	// TODO: パートサイズのパラメータ化
	partMiBs int64 = 5 // 5MiB
)

// ObjectStorageAccessor は、オブジェクトストレージへアクセスするためのインタフェースです。
type ObjectStorageAccessor interface {
	// Upload は、オブジェクトストレージへbyteスライスのデータをアップロードします。
	Upload(bucketName string, objectKey string, objectBody []byte) error
	// UploadFromReader は、オブジェクトストレージへReaderから読み込んだデータをアップロードします。
	UploadFromReader(bucketName string, objectKey string, reader io.Reader) error
	// UploadLargeObject は、オブジェクトストレージへReaderから読み込んだ大きなデータをマルチパートアップロードします。
	// 5MiBより小さい場合には、このメソッドは使用できません。
	UploadLargeObject(bucketName string, objectKey string, reader io.Reader) error
	// Download は、オブジェクトストレージからデータをbyteスライスのデータでダウンロードします。
	Download(bucketName string, objectKey string) ([]byte, error)
	// DownloadToReader は、オブジェクトストレージからデータをReaderでダウンロードします。
	DownloadToReader(bucketName string, objectKey string) (io.ReadCloser, error)
	// DownloadLargeObject は、オブジェクトストレージから大きなデータをファイルにマルチパートダウンロードします。
	// 5MiBより小さい場合には、このメソッドは使用できません。
	DownloadLargeObject(bucketName string, objectKey string, filePath string) error
}

// NewObjectStorageAccessor は、ObjectStorageAccessorを作成します。
func NewObjectStorageAccessor(myCfg myconfig.Config, log logging.Logger) (ObjectStorageAccessor, error) {
	//TODO: カスタムHTTPClientの作成

	// AWS SDK for Go v2 Migration
	// https://github.com/aws/aws-sdk-go-v2
	// https://aws.github.io/aws-sdk-go-v2/docs/migrating/
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, errors.WithStack(err)
	}
	// Instrumenting AWS SDK v2
	// https://github.com/aws/aws-xray-sdk-go
	awsv2.AWSV2Instrumentor(&cfg.APIOptions)
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		// ローカル実行のためS3のLocal起動先が指定されている場合
		s3Endpoint := myCfg.Get(S3_LOCAL_ENDPOINT_NAME, "")
		if s3Endpoint != "" {
			o.BaseEndpoint = aws.String(s3Endpoint)
			// MinIOの場合は、PathStyleをtrueにする
			o.UsePathStyle = true
			key := myCfg.Get("MINIO_ACCESS_KEY", "minioadmin")
			secret := myCfg.Get("MINIO_SECRET_KEY", "minioadmin")
			o.Credentials = aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(key, secret, ""))
		}
	})

	uploader := manager.NewUploader(client, func(u *manager.Uploader) {
		u.PartSize = partMiBs * 1024 * 1024
	})
	downloader := manager.NewDownloader(client, func(d *manager.Downloader) {
		d.PartSize = partMiBs * 1024 * 1024
	})

	return &defaultObjectStorageAccessor{
		s3Client:   client,
		uploader:   uploader,
		downloader: downloader,
		log:        log,
	}, nil
}

// defaultObjectStorageAccessor は、ObjectStorageAccessorのデフォルト実装です。
type defaultObjectStorageAccessor struct {
	log        logging.Logger
	s3Client   *s3.Client
	uploader   *manager.Uploader
	downloader *manager.Downloader
}

// Upload implements ObjectStorageAccessor.
func (oa *defaultObjectStorageAccessor) Upload(bucketName string, objectKey string, objectBody []byte) error {
	oa.log.Debug("Upload bucketName:%s, objectKey:%s", bucketName, objectKey)
	reader := bytes.NewReader(objectBody)
	return oa.UploadFromReader(bucketName, objectKey, reader)
}

// UploadFromReader implements ObjectStorageAccessor.
func (oa *defaultObjectStorageAccessor) UploadFromReader(bucketName string, objectKey string, reader io.Reader) error {
	oa.log.Debug("UploadFromReader bucketName:%s, objectKey:%s", bucketName, objectKey)
	input := &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
		Body:   reader,
	}
	_, err := oa.s3Client.PutObject(apcontext.Context, input)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// UploadLargeObject implements ObjectStorageAccessor.
func (oa *defaultObjectStorageAccessor) UploadLargeObject(bucketName string, objectKey string, reader io.Reader) error {
	input := &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
		Body:   reader,
	}
	_, err := oa.uploader.Upload(apcontext.Context, input)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// Download implements ObjectStorageAccessor.
func (oa *defaultObjectStorageAccessor) Download(bucketName string, objectKey string) ([]byte, error) {
	body, err := oa.DownloadToReader(bucketName, objectKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer body.Close()
	data, err := io.ReadAll(body)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return data, nil
}

// DownloadToReader implements ObjectStorageAccessor.
func (oa *defaultObjectStorageAccessor) DownloadToReader(bucketName string, objectKey string) (io.ReadCloser, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	}
	output, err := oa.s3Client.GetObject(apcontext.Context, input)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return output.Body, nil
}

// DownloadLargeObject implements ObjectStorageAccessor.
func (oa *defaultObjectStorageAccessor) DownloadLargeObject(bucketName string, objectKey string, filePath string) error {
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	}
	f, err := os.Create(filePath)
	if err != nil {
		return errors.WithStack(err)
	}
	defer f.Close()
	_, err = oa.downloader.Download(apcontext.Context, f, input)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
