/*
objectstorage パッケージは、オブジェクトストレージを扱うためのパッケージです。
*/
package objectstorage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"os"

	"example.com/appbase/pkg/apcontext"
	myconfig "example.com/appbase/pkg/config"
	"example.com/appbase/pkg/logging"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
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
	// ListObjectesは、フォルダ（プレフィックス）配下のオブジェクトストレージのオブジェクト一覧を取得します。
	ListObjects(bucketName string, folderName string) ([]types.Object, error)
	// ExistsObject は、オブジェクトストレージにオブジェクトが存在するか確認します。
	ExistsObject(bucketName string, objectKey string) (bool, error)
	// Upload は、オブジェクトストレージへbyteスライスのデータをアップロードします。
	Upload(bucketName string, objectKey string, objectBody []byte) error
	// UploadFromReader は、オブジェクトストレージへReaderから読み込んだデータをアップロードします。
	// readerは、クローズは、呼び出し元にて行う必要があります。
	UploadFromReader(bucketName string, objectKey string, reader io.Reader) error
	// UploadLargeObject は、オブジェクトストレージへReaderから読み込んだ大きなデータをマルチパートアップロードします。
	// 5MiBより小さい場合には、このメソッドは使用できません。
	// readerは、クローズは、呼び出し元にて行う必要があります。
	UploadLargeObject(bucketName string, objectKey string, reader io.Reader) error
	// Download は、オブジェクトストレージからデータをbyteスライスのデータでダウンロードします。
	Download(bucketName string, objectKey string) ([]byte, error)
	// DownloadToReader は、オブジェクトストレージからデータをReaderでダウンロードします。
	DownloadToReader(bucketName string, objectKey string) (io.ReadCloser, error)
	// DownloadLargeObject は、オブジェクトストレージから大きなデータをマルチパートダウンロードして指定のパスのファイルに保存します。
	// 5MiBより小さい場合には、このメソッドは使用できません。
	DownloadLargeObject(bucketName string, objectKey string, filePath string) error
	// Delele は、オブジェクトストレージからデータを削除します。
	Delele(bucketName string, objectKey string) error
	// CopyToFolder は、オブジェクトストレージのオブジェクトを指定フォルダにコピーします。
	CopyToFolder(bucketName string, objectKey string, folderName string) error
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

// ListObjects implements ObjectStorageAccessor.
func (a *defaultObjectStorageAccessor) ListObjects(bucketName string, folderName string) ([]types.Object, error) {
	a.log.Debug("ListObjects bucketName:%s", bucketName)
	input := &s3.ListObjectsV2Input{
		Bucket:  aws.String(bucketName),
		Prefix:  aws.String(folderName),
		MaxKeys: aws.Int32(math.MaxInt32),
	}
	output, err := a.s3Client.ListObjectsV2(apcontext.Context, input)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return output.Contents, nil
}

// ExistsObject implements ObjectStorageAccessor.
func (a *defaultObjectStorageAccessor) ExistsObject(bucketName string, objectKey string) (bool, error) {
	a.log.Debug("ExistsObject bucketName:%s, objectKey:%s", bucketName, objectKey)
	input := &s3.HeadObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	}
	_, err := a.s3Client.HeadObject(apcontext.Context, input)
	if err != nil {
		var notFound *types.NotFound
		if errors.As(err, &notFound) {
			a.log.Debug("Object not found. bucketName:%s, objectKey:%s", bucketName, objectKey)
			return false, nil
		}
		return false, errors.WithStack(err)
	}
	return true, nil
}

// Upload implements ObjectStorageAccessor.
func (a *defaultObjectStorageAccessor) Upload(bucketName string, objectKey string, objectBody []byte) error {
	a.log.Debug("Upload bucketName:%s, objectKey:%s", bucketName, objectKey)
	reader := bytes.NewReader(objectBody)
	return a.UploadFromReader(bucketName, objectKey, reader)
}

// UploadFromReader implements ObjectStorageAccessor.
func (a *defaultObjectStorageAccessor) UploadFromReader(bucketName string, objectKey string, reader io.Reader) error {
	a.log.Debug("UploadFromReader bucketName:%s, objectKey:%s", bucketName, objectKey)
	input := &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
		Body:   reader,
	}
	_, err := a.s3Client.PutObject(apcontext.Context, input)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// UploadLargeObject implements ObjectStorageAccessor.
func (a *defaultObjectStorageAccessor) UploadLargeObject(bucketName string, objectKey string, reader io.Reader) error {
	a.log.Debug("UploadLargeObject bucketName:%s, objectKey:%s", bucketName, objectKey)
	input := &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
		Body:   reader,
	}
	_, err := a.uploader.Upload(apcontext.Context, input)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// Download implements ObjectStorageAccessor.
func (a *defaultObjectStorageAccessor) Download(bucketName string, objectKey string) ([]byte, error) {
	a.log.Debug("Download bucketName:%s, objectKey:%s", bucketName, objectKey)
	body, err := a.DownloadToReader(bucketName, objectKey)
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
func (a *defaultObjectStorageAccessor) DownloadToReader(bucketName string, objectKey string) (io.ReadCloser, error) {
	a.log.Debug("DownloadToReader bucketName:%s, objectKey:%s", bucketName, objectKey)
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	}
	output, err := a.s3Client.GetObject(apcontext.Context, input)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return output.Body, nil
}

// DownloadLargeObject implements ObjectStorageAccessor.
func (a *defaultObjectStorageAccessor) DownloadLargeObject(bucketName string, objectKey string, filePath string) error {
	a.log.Debug("DownloadLargeObject bucketName:%s, objectKey:%s, filePath", bucketName, objectKey, filePath)
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	}
	f, err := os.Create(filePath)
	if err != nil {
		return errors.WithStack(err)
	}
	defer f.Close()
	_, err = a.downloader.Download(apcontext.Context, f, input)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// Delele implements ObjectStorageAccessor.
func (a *defaultObjectStorageAccessor) Delele(bucketName string, objectKey string) error {
	a.log.Debug("Delete bucketName:%s, objectKey:%s", bucketName, objectKey)
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	}
	_, err := a.s3Client.DeleteObject(apcontext.Context, input)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// CopyToFolder implements ObjectStorageAccessor.
func (a *defaultObjectStorageAccessor) CopyToFolder(bucketName string, objectKey string, folderName string) error {
	a.log.Debug("CopyToFolder bucketName:%s, objectKey:%s, folderName:%s", bucketName, objectKey, folderName)
	input := &s3.CopyObjectInput{
		Bucket:     aws.String(bucketName),
		CopySource: aws.String(fmt.Sprintf("%v/%v", bucketName, objectKey)),
		Key:        aws.String(fmt.Sprintf("%v/%v", folderName, objectKey)),
	}
	_, err := a.s3Client.CopyObject(apcontext.Context, input)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
