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
	"strings"

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
	S3_LOCAL_ENDPOINT_NAME         = "S3_LOCAL_ENDPOINT"
	S3_MINIO_ACCESS_KEY_NAME       = "S3_MINIO_ACCESS_KEY"
	S3_MINIO_SECRET_KEY_NAME       = "S3_MINIO_SECRET_KEY"
	S3_UPLOAD_PART_SIZE_MiB_NAME   = "S3_UPLOAD_PART_SIZE_MiB"
	S3_DOWNLOAD_PART_SIZE_MiB_NAME = "S3_DOWNLOAD_PART_SIZE_MiB"
	S3_UPLOAD_CONCURRENCY          = "S3_UPLOAD_CONCURRENCY"
	S3_DOWNLOAD_CONCURRENCY        = "S3_DOWNLOAD_CONCURRENCY"
)

// ObjectStorageAccessor は、オブジェクトストレージへアクセスするためのインタフェースです。
type ObjectStorageAccessor interface {
	// ListObjectesは、フォルダ配下のオブジェクトストレージのオブジェクト一覧を取得します。
	ListObjects(bucketName string, folderPath string) ([]types.Object, error)
	// ExistsObject は、オブジェクトストレージにオブジェクトが存在するか確認します。
	ExistsObject(bucketName string, objectKey string) (bool, error)
	// GetObjectSize は、オブジェクトストレージのオブジェクトのサイズを取得します。
	GetSize(bucketName string, objectKey string) (int64, error)
	// GetObjectMetadata は、オブジェクトストレージのオブジェクトのメタデータを取得します。
	GetObjectMetadata(bucketName string, objectKey string) (*s3.HeadObjectOutput, error)
	// Upload は、オブジェクトストレージへbyteスライスのデータをアップロードします。
	// サイズが5MiBを超える場合は、透過的にマルチパートアップロードを行います。
	Upload(bucketName string, objectKey string, objectBody []byte) error
	// UploadWithOwnerFullControl は、 bucket-owner-full-controlのACLを付与しオブジェクトストレージへbyteスライスのデータをアップロードします。
	// サイズが5MiBを超える場合は、透過的にマルチパートアップロードを行います。
	//（使用しないが参考実装）
	UploadWithOwnerFullControl(bucketName string, objectKey string, objectBody []byte) error
	// UploadFromReader は、オブジェクトストレージへReaderから読み込んだデータをアップロードします。
	// readerは、クローズは、呼び出し元にて行う必要があります。
	// サイズが5MiBを超える場合は、透過的にマルチパートアップロードを行います。
	UploadFromReader(bucketName string, objectKey string, reader io.Reader) error
	// ReadAt は、オブジェクトストレージから指定のオフセットからバイトスライス分読み込みます。
	ReadAt(bucketName string, objectKey string, p []byte, offset int64) (int, error)
	// Download は、オブジェクトストレージからデータをbyteスライスのデータでダウンロードします。
	// マルチパートダウンロードは行いません。
	Download(bucketName string, objectKey string) ([]byte, error)
	// DownloadToReader は、オブジェクトストレージからデータをReaderでダウンロードします。
	// readerは、クローズは、呼び出し元にて行う必要があります。
	// マルチパートダウンロードは行いません。
	DownloadToReader(bucketName string, objectKey string) (io.ReadCloser, error)
	// DownloadToFile は、オブジェクトストレージから大きなデータを指定のローカルファイルに保存します。
	// サイズが5MiBを超える場合は、透過的にマルチパートダウンロードを行います。
	DownloadToFile(bucketName string, objectKey string, filePath string) error
	// Delele は、オブジェクトストレージからデータを削除します。
	Delele(bucketName string, objectKey string) error
	// DeleteFolder は、オブジェクトストレージのフォルダごと削除します。
	// なお、エラーが発生した時点で中断されるため、削除されないファイルが残る可能性があります。
	DeleteFolder(bucketName string, folderPath string) error
	// Copy は、オブジェクトストレージのオブジェクトを指定フォルダにコピーします。
	// 例えば、objectKey = input/xxxx/hoge.txt、output= output とした場合、output/hoge.txtにコピーします。
	Copy(bucketName string, objectKey string, targetFolderPath string) error
	// CopyFolder は、オブジェクトストレージのフォルダごと指定フォルダにコピーします。
	// nestedがtrueの場合、サブフォルダ含めてコピーします。falseの場合、直下のファイルのみコピーします。
	// なお、エラーが発生した時点で中断されるため、途中までコピーされたファイルが残る可能性があります。
	CopyFolder(bucketName string, srcFolderPath string, targetFolderPath string, nested bool) error
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
			key := myCfg.Get(S3_MINIO_ACCESS_KEY_NAME, "minioadmin")
			secret := myCfg.Get(S3_MINIO_SECRET_KEY_NAME, "minioadmin")
			o.Credentials = aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(key, secret, ""))
		}
	})

	// パートサイズのパラメータ取得
	uploadPartMiBs := myCfg.GetInt(S3_UPLOAD_PART_SIZE_MiB_NAME, 5)
	downloadPartMiBs := myCfg.GetInt(S3_DOWNLOAD_PART_SIZE_MiB_NAME, 5)
	// 並列実行数のパラメータ取得
	uploadConcurrency := myCfg.GetInt(S3_UPLOAD_CONCURRENCY, 5)
	downloadConCurrency := myCfg.GetInt(S3_DOWNLOAD_CONCURRENCY, 5)

	// https://aws.github.io/aws-sdk-go-v2/docs/sdk-utilities/s3/#configuration-options
	uploader := manager.NewUploader(client, func(u *manager.Uploader) {
		u.PartSize = int64(uploadPartMiBs) * 1024 * 1024
		u.Concurrency = uploadConcurrency
	})
	// https://aws.github.io/aws-sdk-go-v2/docs/sdk-utilities/s3/#configuration-options-1
	downloader := manager.NewDownloader(client, func(d *manager.Downloader) {
		d.PartSize = int64(downloadPartMiBs) * 1024 * 1024
		d.Concurrency = downloadConCurrency
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
func (a *defaultObjectStorageAccessor) ListObjects(bucketName string, folderPath string) ([]types.Object, error) {
	a.log.Debug("ListObjects bucketName:%s, folderPath:%s", bucketName, folderPath)
	input := &s3.ListObjectsV2Input{
		Bucket:  aws.String(bucketName),
		Prefix:  aws.String(folderPath),
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
	_, err := a.GetObjectMetadata(bucketName, objectKey)
	if err != nil {
		var notFound *types.NotFound
		if errors.As(err, &notFound) {
			a.log.Debug("Object not found. bucketName:%s, objectKey:%s", bucketName, objectKey)
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetSize implements ObjectStorageAccessor.
func (a *defaultObjectStorageAccessor) GetSize(bucketName string, objectKey string) (int64, error) {
	a.log.Debug("GetSize bucketName:%s, objectKey:%s", bucketName, objectKey)
	output, err := a.GetObjectMetadata(bucketName, objectKey)
	if err != nil {
		return 0, err
	}
	return *output.ContentLength, nil
}

// GetObjectMetadata implements ObjectStorageAccessor.
func (a *defaultObjectStorageAccessor) GetObjectMetadata(bucketName string, objectKey string) (*s3.HeadObjectOutput, error) {
	a.log.Debug("GetObjectMetadata bucketName:%s, objectKey:%s", bucketName, objectKey)
	input := &s3.HeadObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	}
	output, err := a.s3Client.HeadObject(apcontext.Context, input)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return output, nil
}

// Upload implements ObjectStorageAccessor.
func (a *defaultObjectStorageAccessor) Upload(bucketName string, objectKey string, objectBody []byte) error {
	a.log.Debug("Upload bucketName:%s, objectKey:%s", bucketName, objectKey)
	reader := bytes.NewReader(objectBody)
	return a.UploadFromReader(bucketName, objectKey, reader)
}

// UploadWithOwnerFullControl implements ObjectStorageAccessor.
func (a *defaultObjectStorageAccessor) UploadWithOwnerFullControl(bucketName string, objectKey string, objectBody []byte) error {
	a.log.Debug("UPloadWithOwnerFullControl bucketName:%s, objectKey:%s", bucketName, objectKey)
	reader := bytes.NewReader(objectBody)
	input := &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
		Body:   reader,
		ACL:    types.ObjectCannedACLBucketOwnerFullControl,
	}
	_, err := a.uploader.Upload(apcontext.Context, input)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// UploadFromReader implements ObjectStorageAccessor.
func (a *defaultObjectStorageAccessor) UploadFromReader(bucketName string, objectKey string, reader io.Reader) error {
	a.log.Debug("UploadFromReader bucketName:%s, objectKey:%s", bucketName, objectKey)
	input := &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
		Body:   reader,
	}
	// https://aws.github.io/aws-sdk-go-v2/docs/sdk-utilities/s3/
	_, err := a.uploader.Upload(apcontext.Context, input)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// ReadAt implements ObjectStorageAccessor.
func (a *defaultObjectStorageAccessor) ReadAt(bucketName string, objectKey string, p []byte, offset int64) (int, error) {
	a.log.Debug("ReadAt bucketName:%s, objectKey:%s, offset:%d", bucketName, objectKey, offset)
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
		Range:  aws.String(fmt.Sprintf("bytes=%d-%d", offset, offset+int64(len(p)))),
	}
	output, err := a.s3Client.GetObject(apcontext.Context, input)
	if err != nil {
		return 0, errors.WithStack(err)
	}
	defer output.Body.Close()
	n, err := io.ReadFull(output.Body, p)
	if err != nil {
		return n, errors.WithStack(err)
	}
	return n, nil
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

// DownloadToFile implements ObjectStorageAccessor.
func (a *defaultObjectStorageAccessor) DownloadToFile(bucketName string, objectKey string, filePath string) error {
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
	// https://aws.github.io/aws-sdk-go-v2/docs/sdk-utilities/s3/
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

// DeleteFolder implements ObjectStorageAccessor.
func (a *defaultObjectStorageAccessor) DeleteFolder(bucketName string, folderPath string) error {
	a.log.Debug("DeleteFolder bucketName:%s, folderPath:%s", bucketName, folderPath)
	// コピー元フォルダに存在するオブジェクトを取得
	objects, err := a.ListObjects(bucketName, folderPath)
	if err != nil {
		return err
	}
	for _, object := range objects {
		err = a.Delele(bucketName, *object.Key)
		if err != nil {
			return err
		}
	}
	return nil
}

// Copy implements ObjectStorageAccessor.
func (a *defaultObjectStorageAccessor) Copy(bucketName string, objectKey string, targetFolderPath string) error {
	a.log.Debug("Copy bucketName:%s, objectKey:%s, targetFolderPath:%s", bucketName, objectKey, targetFolderPath)
	i := strings.LastIndex(objectKey, "/")
	fileName := objectKey[i+1:]
	a.log.Debug("fileName:%s", fileName)
	input := &s3.CopyObjectInput{
		Bucket:     aws.String(bucketName),
		CopySource: aws.String(fmt.Sprintf("%s/%s", bucketName, objectKey)),
		Key:        aws.String(fmt.Sprintf("%s/%s", targetFolderPath, fileName)),
	}
	_, err := a.s3Client.CopyObject(apcontext.Context, input)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// CopyFolder implements ObjectStorageAccessor.
func (a *defaultObjectStorageAccessor) CopyFolder(bucketName string, srcFolderPath string, targetFolderPath string, nested bool) error {
	a.log.Debug("CopyFolder bucketName:%s, srcFolderPath:%s, targetFolderPath:%s, nested:%v", bucketName, srcFolderPath, targetFolderPath, nested)
	srcFolderPath = strings.Trim(srcFolderPath, "/")
	// コピー元フォルダに存在するオブジェクトを取得
	objects, err := a.ListObjects(bucketName, srcFolderPath)
	if err != nil {
		return err
	}
	// 対象のオブジェクトに対して繰り返し処理
	for _, object := range objects {
		a.log.Debug("object.Key:%s", *object.Key)
		// コピー元フォルダ名を除いたパスを取得
		lastPath := strings.TrimPrefix(*object.Key, srcFolderPath)
		// nestedならすべてコピーする
		// nestedでないなら直下のファイルのみ（lastPathに"/"が含まれていない）コピーする
		if nested || !strings.Contains(lastPath, "/") {
			// サブフォルダの場合は、フォルダ名を付与してコピーする
			i := strings.LastIndex(lastPath, "/")
			var actualTargetFolderName string
			if i > 0 {
				actualTargetFolderName = targetFolderPath + lastPath[:i]
			} else {
				actualTargetFolderName = targetFolderPath
			}
			err = a.Copy(bucketName, *object.Key, actualTargetFolderName)
			if err != nil {
				return err
			}

		}
	}
	return nil
}
