/*
objectstorage パッケージは、オブジェクトストレージを扱うためのパッケージです。
*/
package objectstorage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"example.com/appbase/pkg/apcontext"
	"example.com/appbase/pkg/awssdk"
	myconfig "example.com/appbase/pkg/config"
	"example.com/appbase/pkg/logging"
	"example.com/appbase/pkg/message"
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
	S3_UPLOAD_CONCURRENCY_NAME     = "S3_UPLOAD_CONCURRENCY"
	S3_DOWNLOAD_CONCURRENCY_NAME   = "S3_DOWNLOAD_CONCURRENCY"
	S3_MAX_KEY_NUM_NAME            = "S3_MAX_KEY_NUM"
)

// デフォルトのurl.QueryEscape関数の挙動を変えるためのReplacer
// 「+」を「%20」に変換し、「%2」Fを「/」に戻す
var r = strings.NewReplacer("+", "%20", "%2F", "/")

// ObjectStorageAccessor は、オブジェクトストレージへアクセスするためのインタフェースです。
type ObjectStorageAccessor interface {
	// Listは、フォルダ配下のオブジェクトストレージのオブジェクト一覧を取得します。
	List(bucketName string, folderPath string, optFns ...func(*s3.Options)) ([]types.Object, error)
	// Exists は、オブジェクトストレージにオブジェクトが存在するか確認します。
	Exists(bucketName string, objectKey string, optFns ...func(*s3.Options)) (bool, error)
	// GetObjectSize は、オブジェクトストレージのオブジェクトのサイズを取得します。
	GetSize(bucketName string, objectKey string, optFns ...func(*s3.Options)) (int64, error)
	// GetMetadata は、オブジェクトストレージのオブジェクトのメタデータを取得します。
	GetMetadata(bucketName string, objectKey string, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error)
	// Upload は、オブジェクトストレージへbyteスライスのデータをアップロードします。
	// サイズが5MiBを超える場合は、透過的にマルチパートアップロードを行いますが、オンメモリのためあまり大きなサイズは推奨されないメソッドです。
	Upload(bucketName string, objectKey string, objectBody []byte, optFns ...func(*s3.Options)) (*manager.UploadOutput, error)
	// UploadWithOwnerFullControl は、 bucket-owner-full-controlのACLを付与しオブジェクトストレージへbyteスライスのデータをアップロードします。
	// サイズが5MiBを超える場合は、透過的にマルチパートアップロードを行いますが、オンメモリのためあまり大きなサイズは推奨されないメソッドです。
	//（使用しないが参考実装）
	UploadWithOwnerFullControl(bucketName string, objectKey string, objectBody []byte, optFns ...func(*s3.Options)) (*manager.UploadOutput, error)
	// UploadString は、オブジェクトストレージへ文字列のデータをアップロードします。
	UploadString(bucketName string, objectKey string, objectBody string, optFns ...func(*s3.Options)) (*manager.UploadOutput, error)
	// UploadFromReader は、オブジェクトストレージへReaderから読み込んだデータをアップロードします。
	// readerは、クローズは、呼び出し元にて行う必要があります。
	// サイズが5MiBを超える場合は、透過的にマルチパートアップロードを行います。
	UploadFromReader(bucketName string, objectKey string, reader io.Reader, optFns ...func(*s3.Options)) (*manager.UploadOutput, error)
	// UploadFile は、オブジェクトストレージへローカルファイルをアップロードします。
	// サイズが5MiBを超える場合は、透過的にマルチパートアップロードを行います。
	UploadFile(bucketName string, objectKey string, filePath string, optFns ...func(*s3.Options)) (*manager.UploadOutput, error)
	// ReadAt は、オブジェクトストレージから指定のオフセットからバイトスライス分読み込みます。
	// io.ReaderAtと似たインタフェースを提供しています。
	ReadAt(bucketName string, objectKey string, p []byte, offset int64, optFns ...func(*s3.Options)) (int, error)
	// Download は、オブジェクトストレージからデータをbyteスライスのデータでダウンロードします。
	// サイズが5MiBを超える場合は、透過的にマルチパートアップロードを行いますが、オンメモリのためあまり大きなサイズは推奨されないメソッドです。
	Download(bucketName string, objectKey string, optFns ...func(*s3.Options)) ([]byte, error)
	// DownloadAsString は、オブジェクトストレージからデータをダウンロードし、文字列として返却します。
	// サイズが5MiBを超える場合は、透過的にマルチパートアップロードを行いますが、オンメモリのためあまり大きなサイズは推奨されないメソッドです。
	DownloadAsString(bucketName string, objectKey string, optFns ...func(*s3.Options)) (string, error)
	// DownloadAsReader は、オブジェクトストレージからデータをダウンロードし、Readerとして返却します。
	// readerは、クローズは、呼び出し元にて行う必要があります。Readerで返却可能ですが、マルチパートダウンロードは行いません。
	DownloadAsReader(bucketName string, objectKey string, optFns ...func(*s3.Options)) (io.ReadCloser, error)
	// DownloadToWriter は、オブジェクトストレージからデータをダウンロードし、指定のWriterに保存します。
	// サイズが5MiBを超える場合は、透過的にマルチパートダウンロードを行います。
	DownloadToWriter(bucketName string, objectKey string, writer io.WriterAt, optFns ...func(*s3.Options)) error
	// DownloadToFile は、オブジェクトストレージから大きなデータをダウンロードし、指定のローカルファイルに保存します。
	// サイズが5MiBを超える場合は、透過的にマルチパートダウンロードを行います。
	DownloadToFile(bucketName string, objectKey string, filePath string, optFns ...func(*s3.Options)) error
	// Delele は、オブジェクトストレージからデータを削除します。
	Delele(bucketName string, objectKey string, optFns ...func(*s3.Options)) error
	// DeleteByVersionId は、オブジェクトストレージから特定のバージョンのデータを削除します。
	DeleteByVersionId(bucketName string, objectKey string, versionId string, optFns ...func(*s3.Options)) error
	// DeleteAllVersions は、オブジェクトストレージから全てのバージョンのデータを削除します。
	DeleteAllVersions(bucketName string, objectKey string, optFns ...func(*s3.Options)) error
	// DeleteFolder は、オブジェクトストレージのフォルダごと削除します。
	// なお、エラーが発生した時点で中断されるため、削除されないファイルが残る可能性があります。
	DeleteFolder(bucketName string, folderPath string, optFns ...func(*s3.Options)) error
	// Copy は、オブジェクトストレージのオブジェクトを指定フォルダにコピーします。
	// 例えば、objectKey = input/xxxx/hoge.txt、targetFolderPath= output とした場合、output/hoge.txtにコピーします。
	Copy(bucketName string, objectKey string, targetFolderPath string, optFns ...func(*s3.Options)) error
	// CopyAcrossBuckets は、オブジェクトストレージのオブジェクトを指定の別のバケットのフォルダにコピーします。
	// 例えば、bucketName = inputBucket、objectKey = input/xxxx/hoge.txt、
	// targetBucketName = outputBucket、targetFolderPath= output とした場合、outputBucketのoutput/hoge.txtにコピーします。
	CopyAcrossBuckets(bucketName string, objectKey string, targetBucketName string, targetFolderPath string, optFns ...func(*s3.Options)) error
	// CopyFolder は、オブジェクトストレージのフォルダごと指定フォルダにコピーします。
	// nestedがtrueの場合、サブフォルダ含めてコピーします。falseの場合、直下のファイルのみコピーします。
	// なお、エラーが発生した時点で中断されるため、途中までコピーされたファイルが残る可能性があります。
	CopyFolder(bucketName string, srcFolderPath string, targetFolderPath string, nested bool, optFns ...func(*s3.Options)) error
	// CopyFolderAcrossBuckets は、オブジェクトストレージのフォルダごと指定の別のバケットのフォルダにコピーします。
	// nestedがtrueの場合、サブフォルダ含めてコピーします。falseの場合、直下のファイルのみコピーします。
	// なお、エラーが発生した時点で中断されるため、途中までコピーされたファイルが残る可能性があります。
	CopyFolderAcrossBuckets(bucketName string, srcFolderPath string, targetBucketName string, targetFolderPath string, nested bool, optFns ...func(*s3.Options)) error
}

// NewObjectStorageAccessor は、ObjectStorageAccessorを作成します。
func NewObjectStorageAccessor(myCfg myconfig.Config, logger logging.Logger) (ObjectStorageAccessor, error) {
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
	uploadConcurrency := myCfg.GetInt(S3_UPLOAD_CONCURRENCY_NAME, 5)
	downloadConCurrency := myCfg.GetInt(S3_DOWNLOAD_CONCURRENCY_NAME, 5)

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
		logger:     logger,
		config:     myCfg,
		s3Client:   client,
		uploader:   uploader,
		downloader: downloader,
	}, nil
}

// defaultObjectStorageAccessor は、ObjectStorageAccessorのデフォルト実装です。
type defaultObjectStorageAccessor struct {
	logger     logging.Logger
	config     myconfig.Config
	s3Client   *s3.Client
	uploader   *manager.Uploader
	downloader *manager.Downloader
}

// List implements ObjectStorageAccessor.
func (a *defaultObjectStorageAccessor) List(bucketName string, folderPath string, optFns ...func(*s3.Options)) ([]types.Object, error) {
	a.logger.Debug("ListObjects bucketName:%s, folderPath:%s", bucketName, folderPath)
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
		Prefix: aws.String(folderPath),
	}
	paginator := s3.NewListObjectsV2Paginator(a.s3Client, input)

	var objects []types.Object
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(apcontext.Context, optFns...)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		objects = append(objects, page.Contents...)
	}
	return objects, nil

}

// Exists implements ObjectStorageAccessor.
func (a *defaultObjectStorageAccessor) Exists(bucketName string, objectKey string, optFns ...func(*s3.Options)) (bool, error) {
	a.logger.Debug("ExistsObject bucketName:%s, objectKey:%s", bucketName, objectKey)
	_, err := a.GetMetadata(bucketName, objectKey, optFns...)
	if err != nil {
		var notFound *types.NotFound
		if errors.As(err, &notFound) {
			a.logger.Debug("Object not found. bucketName:%s, objectKey:%s", bucketName, objectKey)
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetSize implements ObjectStorageAccessor.
func (a *defaultObjectStorageAccessor) GetSize(bucketName string, objectKey string, optFns ...func(*s3.Options)) (int64, error) {
	a.logger.Debug("GetSize bucketName:%s, objectKey:%s", bucketName, objectKey)
	output, err := a.GetMetadata(bucketName, objectKey, optFns...)
	if err != nil {
		return 0, err
	}
	return *output.ContentLength, nil
}

// GetMetadata implements ObjectStorageAccessor.
func (a *defaultObjectStorageAccessor) GetMetadata(bucketName string, objectKey string, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
	a.logger.Debug("GetObjectMetadata bucketName:%s, objectKey:%s", bucketName, objectKey)
	input := &s3.HeadObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	}
	output, err := a.s3Client.HeadObject(apcontext.Context, input, optFns...)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return output, nil
}

// Upload implements ObjectStorageAccessor.
func (a *defaultObjectStorageAccessor) Upload(bucketName string, objectKey string, objectBody []byte, optFns ...func(*s3.Options)) (*manager.UploadOutput, error) {
	a.logger.Debug("Upload bucketName:%s, objectKey:%s", bucketName, objectKey)
	reader := bytes.NewReader(objectBody)
	return a.UploadFromReader(bucketName, objectKey, reader, optFns...)
}

// UploadWithOwnerFullControl implements ObjectStorageAccessor.
func (a *defaultObjectStorageAccessor) UploadWithOwnerFullControl(bucketName string, objectKey string, objectBody []byte, optFns ...func(*s3.Options)) (*manager.UploadOutput, error) {
	a.logger.Debug("UPloadWithOwnerFullControl bucketName:%s, objectKey:%s", bucketName, objectKey)
	reader := bytes.NewReader(objectBody)
	input := &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
		Body:   reader,
		ACL:    types.ObjectCannedACLBucketOwnerFullControl,
	}
	option := func(o *manager.Uploader) {
		o.ClientOptions = append(o.ClientOptions, optFns...)
	}

	uploadOutput, err := a.uploader.Upload(apcontext.Context, input, option)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return uploadOutput, nil
}

// UploadString implements ObjectStorageAccessor.
func (a *defaultObjectStorageAccessor) UploadString(bucketName string, objectKey string, objectBody string, optFns ...func(*s3.Options)) (*manager.UploadOutput, error) {
	a.logger.Debug("UploadFromString bucketName:%s, objectKey:%s", bucketName, objectKey)
	return a.Upload(bucketName, objectKey, []byte(objectBody), optFns...)
}

// UploadFromReader implements ObjectStorageAccessor.
func (a *defaultObjectStorageAccessor) UploadFromReader(bucketName string, objectKey string, reader io.Reader, optFns ...func(*s3.Options)) (*manager.UploadOutput, error) {
	a.logger.Debug("UploadFromReader bucketName:%s, objectKey:%s", bucketName, objectKey)
	input := &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
		Body:   reader,
	}

	option := func(o *manager.Uploader) {
		o.ClientOptions = append(o.ClientOptions, optFns...)
	}

	// https://aws.github.io/aws-sdk-go-v2/docs/sdk-utilities/s3/
	uploadOutput, err := a.uploader.Upload(apcontext.Context, input, option)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return uploadOutput, nil
}

// UploadFile implements ObjectStorageAccessor.
func (a *defaultObjectStorageAccessor) UploadFile(bucketName string, objectKey string, filePath string, optFns ...func(*s3.Options)) (*manager.UploadOutput, error) {
	a.logger.Debug("UploadFromFile bucketName:%s, objectKey:%s, filePath:%s", bucketName, objectKey, filePath)
	f, err := os.Open(filePath)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer f.Close()
	return a.UploadFromReader(bucketName, objectKey, f, optFns...)
}

// ReadAt implements ObjectStorageAccessor.
func (a *defaultObjectStorageAccessor) ReadAt(bucketName string, objectKey string, p []byte, offset int64, optFns ...func(*s3.Options)) (int, error) {
	a.logger.Debug("ReadAt bucketName:%s, objectKey:%s, offset:%d", bucketName, objectKey, offset)
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
		Range:  aws.String(fmt.Sprintf("bytes=%d-%d", offset, offset+int64(len(p)))),
	}
	output, err := a.s3Client.GetObject(apcontext.Context, input, optFns...)
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
func (a *defaultObjectStorageAccessor) Download(bucketName string, objectKey string, optFns ...func(*s3.Options)) ([]byte, error) {
	a.logger.Debug("Download bucketName:%s, objectKey:%s", bucketName, objectKey)
	// GetObjectを使用する方法だと、コメントアウト部分のコードの通り、1回のリクエスト呼び出しでデータ取得できるが、
	// 本サンプルでは、Downloaderを使用してマルチパート対応した方法を使用している。
	// その代わり、HeadObjectを呼び出し、サイズ情報を取得しバッファを確保してからダウンロードする必要があるので、
	// APIの呼び出しが2回になる。
	/*
		body, err := a.DownloadAsReader(bucketName, objectKey, optFns...)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		defer body.Close()
		data, err := io.ReadAll(body)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return data, nil
	*/
	size, err := a.GetSize(bucketName, objectKey)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, size)
	w := manager.NewWriteAtBuffer(buf)
	err = a.DownloadToWriter(bucketName, objectKey, w, optFns...)
	if err != nil {
		return nil, err

	}
	return buf, nil
}

// DownloadAsString implements ObjectStorageAccessor.
func (a *defaultObjectStorageAccessor) DownloadAsString(bucketName string, objectKey string, optFns ...func(*s3.Options)) (string, error) {
	a.logger.Debug("DownloadToString bucketName:%s, objectKey:%s", bucketName, objectKey)
	data, err := a.Download(bucketName, objectKey, optFns...)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// DownloadAsReader implements ObjectStorageAccessor.
func (a *defaultObjectStorageAccessor) DownloadAsReader(bucketName string, objectKey string, optFns ...func(*s3.Options)) (io.ReadCloser, error) {
	a.logger.Debug("DownloadToReader bucketName:%s, objectKey:%s", bucketName, objectKey)
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	}
	output, err := a.s3Client.GetObject(apcontext.Context, input, optFns...)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return output.Body, nil
}

// DownloadToWriter implements ObjectStorageAccessor.
func (a *defaultObjectStorageAccessor) DownloadToWriter(bucketName string, objectKey string, writer io.WriterAt, optFns ...func(*s3.Options)) error {
	a.logger.Debug("DownloadToWriter bucketName:%s, objectKey:%s", bucketName, objectKey)
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	}

	option := func(o *manager.Downloader) {
		o.ClientOptions = append(o.ClientOptions, optFns...)
	}

	// https://aws.github.io/aws-sdk-go-v2/docs/sdk-utilities/s3/
	_, err := a.downloader.Download(apcontext.Context, writer, input, option)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// DownloadToFile implements ObjectStorageAccessor.
func (a *defaultObjectStorageAccessor) DownloadToFile(bucketName string, objectKey string, filePath string, optFns ...func(*s3.Options)) error {
	a.logger.Debug("DownloadLargeObject bucketName:%s, objectKey:%s, filePath", bucketName, objectKey, filePath)
	f, err := os.Create(filePath)
	if err != nil {
		return errors.WithStack(err)
	}
	defer f.Close()
	return a.DownloadToWriter(bucketName, objectKey, f, optFns...)
}

// Delele implements ObjectStorageAccessor.
func (a *defaultObjectStorageAccessor) Delele(bucketName string, objectKey string, optFns ...func(*s3.Options)) error {
	a.logger.Debug("Delete bucketName:%s, objectKey:%s", bucketName, objectKey)
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	}
	_, err := a.s3Client.DeleteObject(apcontext.Context, input, optFns...)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// DeleteByVersionId implements ObjectStorageAccessor.
func (a *defaultObjectStorageAccessor) DeleteByVersionId(bucketName string, objectKey string, versionId string, optFns ...func(*s3.Options)) error {
	a.logger.Debug("DeleteByVersionId bucketName:%s, objectKey:%s, versionId:%s", bucketName, objectKey, versionId)
	input := &s3.DeleteObjectInput{
		Bucket:    aws.String(bucketName),
		Key:       aws.String(objectKey),
		VersionId: aws.String(versionId),
	}
	_, err := a.s3Client.DeleteObject(apcontext.Context, input, optFns...)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// DeleteAllVersions implements ObjectStorageAccessor.
func (a *defaultObjectStorageAccessor) DeleteAllVersions(bucketName string, objectKey string, optFns ...func(*s3.Options)) error {
	a.logger.Debug("DeleteAllVersions bucketName:%s, objectKey:%s", bucketName, objectKey)

	// オブジェクトの全てのバージョンIDを取得
	versions, err := a.listObjectVersions(bucketName, objectKey, optFns...)
	if err != nil {
		return err
	}
	if len(versions) == 0 {
		return errors.New(fmt.Sprintf("削除対象のオブジェクトが存在しません。 bucketName:%s, objectKey:%s", bucketName, objectKey))
	}
	// DeleteObjectesは最大1000件までの削除が可能
	// 設定のChunkSizeごとに分割して、削除処理を行う
	chunkSize := a.config.GetInt(S3_MAX_KEY_NUM_NAME, 1000)
	chunkedVersion := chunkBy(versions, chunkSize)

	for i, chunk := range chunkedVersion {
		a.logger.Debug("chunk: %2d", i+1)
		objects := make([]types.ObjectIdentifier, 0, len(chunk))
		for j, v := range chunk {
			a.logger.Debug("%3d: key:%s versionId:%s", j+1, *v.Key, *v.VersionId)
			objects = append(objects, types.ObjectIdentifier{
				Key: v.Key,
			})
		}
		input := &s3.DeleteObjectsInput{
			Bucket: aws.String(bucketName),
			Delete: &types.Delete{
				Objects: objects,
			},
		}
		output, err := a.s3Client.DeleteObjects(apcontext.Context, input, optFns...)
		if err != nil {
			return errors.WithStack(err)
		}
		if output.Errors != nil {
			for _, e := range output.Errors {
				a.logger.Warn(message.W_FW_8011, bucketName, *e.Key, *e.VersionId, *e.Code, *e.Message)
			}
			return errors.New("DeleteAllVersionsでオブジェクトの削除操作に失敗")
		}
	}
	return nil
}

// listObjectVersions は、指定のバケットとキーに対する全てのバージョンを取得します。
func (a *defaultObjectStorageAccessor) listObjectVersions(bucketName string, objectKey string, optFns ...func(*s3.Options)) ([]types.ObjectVersion, error) {
	input := &s3.ListObjectVersionsInput{
		Bucket:  aws.String(bucketName),
		MaxKeys: aws.Int32(int32(a.config.GetInt(S3_MAX_KEY_NUM_NAME, 1000))),
		Prefix:  aws.String(objectKey),
	}
	var err error
	var output *s3.ListObjectVersionsOutput
	var versions []types.ObjectVersion
	paginator := s3.NewListObjectVersionsPaginator(a.s3Client, input)
	for paginator.HasMorePages() {
		output, err = paginator.NextPage(apcontext.Context, optFns...)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		for _, version := range output.Versions {
			// ListObjectVersionsは、Prefix指定しかできないため
			// キーが完全一致するバージョンのみを取得
			if *version.Key == objectKey {
				versions = append(versions, version)
			}
		}
	}
	return versions, nil
}

// chunkBy は、指定のサイズでスライスを分割します。
func chunkBy[T any](items []T, chunkSize int) [][]T {
	if len(items) == 0 || chunkSize <= 0 {
		return [][]T{}
	}

	// https://go.dev/wiki/SliceTricks#batching-with-minimal-allocation
	chunks := make([][]T, 0, (len(items)+chunkSize-1)/chunkSize)
	for chunkSize < len(items) {
		items, chunks = items[chunkSize:], append(chunks, items[0:chunkSize:chunkSize])
	}
	return append(chunks, items)
}

// DeleteFolder implements ObjectStorageAccessor.
func (a *defaultObjectStorageAccessor) DeleteFolder(bucketName string, folderPath string, optFns ...func(*s3.Options)) error {
	a.logger.Debug("DeleteFolder bucketName:%s, folderPath:%s", bucketName, folderPath)
	// コピー元フォルダに存在するオブジェクトを取得
	objects, err := a.List(bucketName, folderPath)
	if err != nil {
		return err
	}
	for _, object := range objects {
		err = a.Delele(bucketName, *object.Key, optFns...)
		if err != nil {
			return err
		}
	}
	return nil
}

// Copy implements ObjectStorageAccessor.
func (a *defaultObjectStorageAccessor) Copy(bucketName string, objectKey string, targetFolderPath string, optFns ...func(*s3.Options)) error {
	a.logger.Debug("Copy bucketName:%s, objectKey:%s, targetFolderPath:%s", bucketName, objectKey, targetFolderPath)
	return a.CopyAcrossBuckets(bucketName, objectKey, bucketName, targetFolderPath, optFns...)
}

// CopyAcrossBuckets implements ObjectStorageAccessor.
func (a *defaultObjectStorageAccessor) CopyAcrossBuckets(bucketName string, objectKey string, targetBucketName string, targetFolderPath string,
	optFns ...func(*s3.Options)) error {
	a.logger.Debug("CopyAcrossBuckets bucketName:%s, objectKey:%s, targetBucketName:%s, targetFolderPath:%s", bucketName, objectKey, targetBucketName, targetFolderPath)
	i := strings.LastIndex(objectKey, "/")
	fileName := objectKey[i+1:]
	a.logger.Debug("fileName:%s", fileName)
	input := &s3.CopyObjectInput{
		Bucket:     aws.String(targetBucketName),
		CopySource: aws.String(encodeURL(fmt.Sprintf("%s/%s", bucketName, objectKey))),
		Key:        aws.String(fmt.Sprintf("%s/%s", targetFolderPath, fileName)),
	}
	_, err := a.s3Client.CopyObject(apcontext.Context, input, optFns...)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// CopyFolder implements ObjectStorageAccessor.
func (a *defaultObjectStorageAccessor) CopyFolder(bucketName string, srcFolderPath string, targetFolderPath string, nested bool, optFns ...func(*s3.Options)) error {
	a.logger.Debug("CopyFolder bucketName:%s, srcFolderPath:%s, targetFolderPath:%s, nested:%v", bucketName, srcFolderPath, targetFolderPath, nested)
	return a.CopyFolderAcrossBuckets(bucketName, srcFolderPath, bucketName, targetFolderPath, nested, optFns...)
}

// CopyFolderAcrossBuckets implements ObjectStorageAccessor.
func (a *defaultObjectStorageAccessor) CopyFolderAcrossBuckets(bucketName string, srcFolderPath string, targetBucketName string, targetFolderPath string, nested bool, optFns ...func(*s3.Options)) error {
	a.logger.Debug("CopyFolderAcrossBuckets bucketName:%s, srcFolderPath:%s, targetBucketName:%s, targetFolderPath:%s, nested:%v", bucketName, srcFolderPath, targetBucketName, targetFolderPath, nested)
	srcFolderPath = strings.Trim(srcFolderPath, "/")
	targetFolderPath = strings.Trim(targetFolderPath, "/")
	// コピー元とコピー先が同じ場合は何もしない
	if bucketName == targetBucketName && srcFolderPath == targetFolderPath {
		return nil
	}

	// コピー元フォルダに存在するオブジェクトを取得
	objects, err := a.List(bucketName, srcFolderPath, optFns...)
	if err != nil {
		return err
	}
	// 対象のオブジェクトに対して繰り返し処理
	for _, object := range objects {
		a.logger.Debug("object.Key:%s", *object.Key)
		// コピー元フォルダ名を除いたパスを取得
		lastPath := strings.TrimPrefix(*object.Key, srcFolderPath)
		// nestedならすべてコピーする
		// nestedでないなら直下のファイルのみ（lastPathに"/"が含まれていない）コピーする
		if nested || !strings.Contains(lastPath, "/") {
			if *object.Size == 0 && strings.HasSuffix(*object.Key, "/") {
				// サイズが0でキーがスラッシュで終わる場合はフォルダなので、サイズ0のファイルを作成し空フォルダのコピーも行う
				a.logger.Debug("Create empty folder. targetBucketName:%s, targetFolderPath:%s", targetBucketName, targetFolderPath+lastPath)
				_, err = a.Upload(targetBucketName, targetFolderPath+lastPath, []byte{}, optFns...)
			} else {
				i := strings.LastIndex(lastPath, "/")
				var actualTargetFolderName string
				if i > 0 {
					// （スラッシュを含むので）サブフォルダの場合は、サブフォルダ名を付与
					actualTargetFolderName = targetFolderPath + lastPath[:i]
				} else {
					// （スラッシュを含まないので）サブフォルダがない場合は、引数のフォルダ名をそのまま使用
					actualTargetFolderName = targetFolderPath
				}
				err = a.CopyAcrossBuckets(bucketName, *object.Key, targetBucketName, actualTargetFolderName, optFns...)
			}
			if err != nil {
				return err
			}

		}
	}
	return nil
}

// encodeURL は、「/」以外をRFC3986に基づくURLエンコードします。
func encodeURL(uri string) string {
	return r.Replace(url.QueryEscape(uri))
}
