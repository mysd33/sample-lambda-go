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
	// Listは、フォルダ配下のオブジェクトストレージのオブジェクト一覧を取得します。
	List(bucketName string, folderPath string) ([]types.Object, error)
	// Exists は、オブジェクトストレージにオブジェクトが存在するか確認します。
	Exists(bucketName string, objectKey string) (bool, error)
	// GetObjectSize は、オブジェクトストレージのオブジェクトのサイズを取得します。
	GetSize(bucketName string, objectKey string) (int64, error)
	// GetMetadata は、オブジェクトストレージのオブジェクトのメタデータを取得します。
	GetMetadata(bucketName string, objectKey string) (*s3.HeadObjectOutput, error)
	// Upload は、オブジェクトストレージへbyteスライスのデータをアップロードします。
	// サイズが5MiBを超える場合は、透過的にマルチパートアップロードを行いますが、オンメモリのためあまり大きなサイズは推奨されないメソッドです。
	Upload(bucketName string, objectKey string, objectBody []byte) error
	// UploadWithOwnerFullControl は、 bucket-owner-full-controlのACLを付与しオブジェクトストレージへbyteスライスのデータをアップロードします。
	// サイズが5MiBを超える場合は、透過的にマルチパートアップロードを行いますが、オンメモリのためあまり大きなサイズは推奨されないメソッドです。
	//（使用しないが参考実装）
	UploadWithOwnerFullControl(bucketName string, objectKey string, objectBody []byte) error
	// UploadString は、オブジェクトストレージへ文字列のデータをアップロードします。
	UploadString(bucketName string, objectKey string, objectBody string) error
	// UploadFromReader は、オブジェクトストレージへReaderから読み込んだデータをアップロードします。
	// readerは、クローズは、呼び出し元にて行う必要があります。
	// サイズが5MiBを超える場合は、透過的にマルチパートアップロードを行います。
	UploadFromReader(bucketName string, objectKey string, reader io.Reader) error
	// UploadFile は、オブジェクトストレージへローカルファイルをアップロードします。
	// サイズが5MiBを超える場合は、透過的にマルチパートアップロードを行います。
	UploadFile(bucketName string, objectKey string, filePath string) error
	// ReadAt は、オブジェクトストレージから指定のオフセットからバイトスライス分読み込みます。
	// io.ReaderAtと似たインタフェースを提供しています。
	ReadAt(bucketName string, objectKey string, p []byte, offset int64) (int, error)
	// Download は、オブジェクトストレージからデータをbyteスライスのデータでダウンロードします。
	// サイズが5MiBを超える場合は、透過的にマルチパートアップロードを行いますが、オンメモリのためあまり大きなサイズは推奨されないメソッドです。
	Download(bucketName string, objectKey string) ([]byte, error)
	// DownloadAsString は、オブジェクトストレージからデータをダウンロードし、文字列として返却します。
	// サイズが5MiBを超える場合は、透過的にマルチパートアップロードを行いますが、オンメモリのためあまり大きなサイズは推奨されないメソッドです。
	DownloadAsString(bucketName string, objectKey string) (string, error)
	// DownloadAsReader は、オブジェクトストレージからデータをダウンロードし、Readerとして返却します。
	// readerは、クローズは、呼び出し元にて行う必要があります。Readerで返却可能ですが、マルチパートダウンロードは行いません。
	DownloadAsReader(bucketName string, objectKey string) (io.ReadCloser, error)
	// DownloadToWriter は、オブジェクトストレージからデータをダウンロードし、指定のWriterに保存します。
	// サイズが5MiBを超える場合は、透過的にマルチパートダウンロードを行います。
	DownloadToWriter(bucketName string, objectKey string, writer io.WriterAt) error
	// DownloadToFile は、オブジェクトストレージから大きなデータをダウンロードし、指定のローカルファイルに保存します。
	// サイズが5MiBを超える場合は、透過的にマルチパートダウンロードを行います。
	DownloadToFile(bucketName string, objectKey string, filePath string) error
	// Delele は、オブジェクトストレージからデータを削除します。
	Delele(bucketName string, objectKey string) error
	// DeleteFolder は、オブジェクトストレージのフォルダごと削除します。
	// なお、エラーが発生した時点で中断されるため、削除されないファイルが残る可能性があります。
	DeleteFolder(bucketName string, folderPath string) error
	// Copy は、オブジェクトストレージのオブジェクトを指定フォルダにコピーします。
	// 例えば、objectKey = input/xxxx/hoge.txt、targetFolderPath= output とした場合、output/hoge.txtにコピーします。
	Copy(bucketName string, objectKey string, targetFolderPath string) error
	// CopyAcrossBuckets は、オブジェクトストレージのオブジェクトを指定の別のバケットのフォルダにコピーします。
	// 例えば、bucketName = inputBucket、objectKey = input/xxxx/hoge.txt、
	// targetBucketName = outputBucket、targetFolderPath= output とした場合、outputBucketのoutput/hoge.txtにコピーします。
	CopyAcrossBuckets(bucketName string, objectKey string, targetBucketName string, targetFolderPath string) error
	// CopyFolder は、オブジェクトストレージのフォルダごと指定フォルダにコピーします。
	// nestedがtrueの場合、サブフォルダ含めてコピーします。falseの場合、直下のファイルのみコピーします。
	// なお、エラーが発生した時点で中断されるため、途中までコピーされたファイルが残る可能性があります。
	CopyFolder(bucketName string, srcFolderPath string, targetFolderPath string, nested bool) error
	// CopyFolderAcrossBuckets は、オブジェクトストレージのフォルダごと指定の別のバケットのフォルダにコピーします。
	// nestedがtrueの場合、サブフォルダ含めてコピーします。falseの場合、直下のファイルのみコピーします。
	// なお、エラーが発生した時点で中断されるため、途中までコピーされたファイルが残る可能性があります。
	CopyFolderAcrossBuckets(bucketName string, srcFolderPath string, targetBucketName string, targetFolderPath string, nested bool) error
}

// NewObjectStorageAccessor は、ObjectStorageAccessorを作成します。
func NewObjectStorageAccessor(myCfg myconfig.Config, log logging.Logger) (ObjectStorageAccessor, error) {
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

// List implements ObjectStorageAccessor.
func (a *defaultObjectStorageAccessor) List(bucketName string, folderPath string) ([]types.Object, error) {
	a.log.Debug("ListObjects bucketName:%s, folderPath:%s", bucketName, folderPath)
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
		Prefix: aws.String(folderPath),
	}
	paginator := s3.NewListObjectsV2Paginator(a.s3Client, input)

	var objects []types.Object
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(apcontext.Context)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		objects = append(objects, page.Contents...)
	}

	/*
		output, err := a.s3Client.ListObjectsV2(apcontext.Context, input)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		token := output.NextContinuationToken
		if token == nil {
			return output.Contents, nil
		}
		// ページネーション処理
		objects := output.Contents
		for {
			input := &s3.ListObjectsV2Input{
				Bucket:            aws.String(bucketName),
				Prefix:            aws.String(folderPath),
				ContinuationToken: token,
			}
			output, err := a.s3Client.ListObjectsV2(apcontext.Context, input)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			objects = append(objects, output.Contents...)
			token = output.NextContinuationToken
			if token == nil {
				break
			}
		}
	*/
	return objects, nil

}

// Exists implements ObjectStorageAccessor.
func (a *defaultObjectStorageAccessor) Exists(bucketName string, objectKey string) (bool, error) {
	a.log.Debug("ExistsObject bucketName:%s, objectKey:%s", bucketName, objectKey)
	_, err := a.GetMetadata(bucketName, objectKey)
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
	output, err := a.GetMetadata(bucketName, objectKey)
	if err != nil {
		return 0, err
	}
	return *output.ContentLength, nil
}

// GetMetadata implements ObjectStorageAccessor.
func (a *defaultObjectStorageAccessor) GetMetadata(bucketName string, objectKey string) (*s3.HeadObjectOutput, error) {
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

// UploadString implements ObjectStorageAccessor.
func (a *defaultObjectStorageAccessor) UploadString(bucketName string, objectKey string, objectBody string) error {
	a.log.Debug("UploadFromString bucketName:%s, objectKey:%s", bucketName, objectKey)
	return a.Upload(bucketName, objectKey, []byte(objectBody))
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

// UploadFile implements ObjectStorageAccessor.
func (a *defaultObjectStorageAccessor) UploadFile(bucketName string, objectKey string, filePath string) error {
	a.log.Debug("UploadFromFile bucketName:%s, objectKey:%s, filePath:%s", bucketName, objectKey, filePath)
	f, err := os.Open(filePath)
	if err != nil {
		return errors.WithStack(err)
	}
	defer f.Close()
	return a.UploadFromReader(bucketName, objectKey, f)
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
	// GetObjectを使用する方法だと、コメントアウト部分のコードの通り、1回のリクエスト呼び出しでデータ取得できるが、
	// 本サンプルでは、Downloaderを使用してマルチパート対応した方法を使用している。
	// その代わり、HeadObjectを呼び出し、サイズ情報を取得しバッファを確保してからダウンロードする必要があるので、
	// APIの呼び出しが2回になる。
	/*
		body, err := a.DownloadAsReader(bucketName, objectKey)
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
	err = a.DownloadToWriter(bucketName, objectKey, w)
	if err != nil {
		return nil, err

	}
	return buf, nil
}

// DownloadAsString implements ObjectStorageAccessor.
func (a *defaultObjectStorageAccessor) DownloadAsString(bucketName string, objectKey string) (string, error) {
	a.log.Debug("DownloadToString bucketName:%s, objectKey:%s", bucketName, objectKey)
	data, err := a.Download(bucketName, objectKey)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// DownloadAsReader implements ObjectStorageAccessor.
func (a *defaultObjectStorageAccessor) DownloadAsReader(bucketName string, objectKey string) (io.ReadCloser, error) {
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

// DownloadToWriter implements ObjectStorageAccessor.
func (a *defaultObjectStorageAccessor) DownloadToWriter(bucketName string, objectKey string, writer io.WriterAt) error {
	a.log.Debug("DownloadToWriter bucketName:%s, objectKey:%s", bucketName, objectKey)
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	}
	// https://aws.github.io/aws-sdk-go-v2/docs/sdk-utilities/s3/
	_, err := a.downloader.Download(apcontext.Context, writer, input)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// DownloadToFile implements ObjectStorageAccessor.
func (a *defaultObjectStorageAccessor) DownloadToFile(bucketName string, objectKey string, filePath string) error {
	a.log.Debug("DownloadLargeObject bucketName:%s, objectKey:%s, filePath", bucketName, objectKey, filePath)
	f, err := os.Create(filePath)
	if err != nil {
		return errors.WithStack(err)
	}
	defer f.Close()
	return a.DownloadToWriter(bucketName, objectKey, f)
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
	objects, err := a.List(bucketName, folderPath)
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
	return a.CopyAcrossBuckets(bucketName, objectKey, bucketName, targetFolderPath)
}

// CopyAcrossBuckets implements ObjectStorageAccessor.
func (a *defaultObjectStorageAccessor) CopyAcrossBuckets(bucketName string, objectKey string, targetBucketName string, targetFolderPath string) error {
	a.log.Debug("CopyAcrossBuckets bucketName:%s, objectKey:%s, targetBucketName:%s, targetFolderPath:%s", bucketName, objectKey, targetBucketName, targetFolderPath)
	i := strings.LastIndex(objectKey, "/")
	fileName := objectKey[i+1:]
	a.log.Debug("fileName:%s", fileName)
	input := &s3.CopyObjectInput{
		Bucket:     aws.String(targetBucketName),
		CopySource: aws.String(encodeURL(fmt.Sprintf("%s/%s", bucketName, objectKey))),
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
	return a.CopyFolderAcrossBuckets(bucketName, srcFolderPath, bucketName, targetFolderPath, nested)
}

// CopyFolderAcrossBuckets implements ObjectStorageAccessor.
func (a *defaultObjectStorageAccessor) CopyFolderAcrossBuckets(bucketName string, srcFolderPath string, targetBucketName string, targetFolderPath string, nested bool) error {
	a.log.Debug("CopyFolderAcrossBuckets bucketName:%s, srcFolderPath:%s, targetBucketName:%s, targetFolderPath:%s, nested:%v", bucketName, srcFolderPath, targetBucketName, targetFolderPath, nested)
	srcFolderPath = strings.Trim(srcFolderPath, "/")
	targetFolderPath = strings.Trim(targetFolderPath, "/")
	// コピー元とコピー先が同じ場合は何もしない
	if bucketName == targetBucketName && srcFolderPath == targetFolderPath {
		return nil
	}

	// コピー元フォルダに存在するオブジェクトを取得
	objects, err := a.List(bucketName, srcFolderPath)
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
			if *object.Size == 0 && strings.HasSuffix(*object.Key, "/") {
				// サイズが0でキーがスラッシュで終わる場合はフォルダなので、サイズ0のファイルを作成し空フォルダのコピーも行う
				err = a.Upload(targetBucketName, targetFolderPath+lastPath, []byte{})
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
				err = a.CopyAcrossBuckets(bucketName, *object.Key, targetBucketName, actualTargetFolderName)
			}
			if err != nil {
				return err
			}

		}
	}
	return nil
}

// encodeURL は、パスをURLエンコードします。
func encodeURL(uri string) string {
	return url.PathEscape(uri)
}
