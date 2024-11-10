/*
documentdb パッケージは、DocumentDB（MongoDB）アクセス機能を提供するパッケージです。
*/
package documentdb

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"example.com/appbase/pkg/apcontext"
	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/env"
	"example.com/appbase/pkg/logging"
	"github.com/cockroachdb/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	DOCUMENTDB_ENDPOINT_NAME           = "DOCUMENTDB_ENDPOINT"
	DOCUMENTDB_PORT_NAME               = "DOCUMENTDB_PORT"
	DOCUMENTDB_DBNAME_NAME             = "DOCUMENTDB_DB_NAME"
	DOCUMENTDB_USERNAME_NAME           = "DOCUMENTDB_USER_NAME"
	DOCUMENTDB_PASSWORD_NAME           = "DOCUMENTDB_PASSWORD"
	DOCUMENTDB_READ_PREFERENCE_NAME    = "DOCUMENTDB_READ_PREFERENCE"
	DOCUMENTDB_CA_FILEPATH_NAME        = "DOCUMENTDB_CA_FILEPATH"
	DOCUMENTDB_CONNECTION_TIMEOUT_NAME = "DOCUMENTDB_CONNECTION_TIMEOUT"
)

// DocumentDBAccessor は、DocumentDB（MongoDB）アクセス機能を提供するインターフェースです。
type DocumentDBAccessor interface {
	// GetMongoClient は、MongoDBのクライアントを取得します。
	GetMongoClient() *mongo.Client
	// GetMongoDatabase は、MongoDBのデータベースを取得します。
	GetMongoDatabase() *mongo.Database
	// GetCurrentContextWithTimout は、MongoDBへアクセスする際のタイムアウト設定付きのコンテキストを取得します。
	GetDefaultContextWithTimeout() (context.Context, context.CancelFunc)
	// GetContextWithTimeout は、指定したコンテキストに対してMongoDBへアクセスする際のタイムアウト設定付きのコンテキストを取得します。
	GetContextWithTimeout(ctx context.Context) (context.Context, context.CancelFunc)
}

// defaultDocumentDBAccessor は、DocumentDBAccessorインタフェースを実装する構造体です。
type defaultDocumentDBAccessor struct {
	config   config.Config
	client   *mongo.Client
	database *mongo.Database
}

// NewDocumentDBAccessor は、DocumentDBAccessorを作成します。
func NewDocumentDBAccessor(config config.Config, logger logging.Logger) (DocumentDBAccessor, error) {
	// DocumentDBのエンドポイント
	documentdbEndpoint, found := config.GetWithContains(DOCUMENTDB_ENDPOINT_NAME)
	if !found {
		return nil, errors.Newf("%s が設定されていません", DOCUMENTDB_ENDPOINT_NAME)
	}
	// DocumentDBのポート（デフォルト27017番）
	documentdbPort := config.Get(DOCUMENTDB_PORT_NAME, "27017")
	// DocumentDBのデータベース名
	dbName, found := config.GetWithContains(DOCUMENTDB_DBNAME_NAME)
	if !found {
		return nil, errors.Newf("%s が設定されていません", DOCUMENTDB_DBNAME_NAME)
	}
	// DocumentDBのユーザ名
	userName, found := config.GetWithContains(DOCUMENTDB_USERNAME_NAME)
	if !found {
		return nil, errors.Newf("%s が設定されていません", DOCUMENTDB_USERNAME_NAME)
	}
	// DocumentDBのパスワード
	password, found := config.GetWithContains(DOCUMENTDB_PASSWORD_NAME)
	if !found {
		return nil, errors.Newf("%s が設定されていません", DOCUMENTDB_PASSWORD_NAME)
	}

	var client *mongo.Client
	var err error

	if env.IsLocalOrLocalTest() {
		// ローカル実行時のMongoDB用の接続文字列作成
		// （参考）https://qiita.com/chenglin/items/ecf6f67e8f80c4750204
		connectionStringTemplate := "mongodb://%s:%s@%s:%s/%s?authSource=admin"
		connectionString := fmt.Sprintf(connectionStringTemplate, userName, password, documentdbEndpoint, documentdbPort, dbName)
		logger.Debug("接続文字列: %s", connectionString)
		// ローカルのMongoDBに接続しmongo.Clientを取得

		client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(connectionString))
	} else {
		// クラウド動作環境の場合、DocumentDB用の接続文字列作成
		//（参考） https://docs.aws.amazon.com/ja_jp/documentdb/latest/developerguide/connect_programmatically.html#connect_programmatically-tls_enabled

		// DocumentDBのリードプリファレンス（デフォルトはsecondaryPreferred）
		readPreference := config.Get(DOCUMENTDB_READ_PREFERENCE_NAME, "secondaryPreferred")
		connectionStringTemplate := "mongodb://%s:%s@%s:%s/%s?tls=true&replicaSet=rs0&readpreference=%s"
		connectionString := fmt.Sprintf(connectionStringTemplate, userName, password, documentdbEndpoint, documentdbPort, dbName, readPreference)
		logger.Debug("接続文字列: %s", connectionString)
		// DocumentDBの場合、global-bundle.pemという DocumentDBの公開鍵を使用してTLS接続を行う
		caFilePath := config.Get(DOCUMENTDB_CA_FILEPATH_NAME, "configs/global-bundle.pem")
		var tlsConfig *tls.Config
		tlsConfig, err = getCustomTLSConfig(caFilePath)
		if err != nil {
			return nil, errors.Wrap(err, "TLS設定の取得に失敗しました")
		}
		// DocumentDBに接続しmongo.Clientを取得
		client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(connectionString).SetTLSConfig(tlsConfig))
	}

	if err != nil {
		return nil, errors.Wrap(err, "DocumentDB(MongoDB)への接続に失敗しました")
	}

	return &defaultDocumentDBAccessor{
		config:   config,
		client:   client,
		database: client.Database(dbName),
	}, nil
}

// GetMongoClient implements DocumentDBAccessor.
func (d *defaultDocumentDBAccessor) GetMongoClient() *mongo.Client {
	return d.client
}

// GetMongoDatabase implements DocumentDBAccessor.
func (d *defaultDocumentDBAccessor) GetMongoDatabase() *mongo.Database {
	return d.database
}

// GetCurrentContext implements DocumentDBAccessor.
func (d *defaultDocumentDBAccessor) GetDefaultContextWithTimeout() (context.Context, context.CancelFunc) {
	return d.GetContextWithTimeout(apcontext.Context)
}

// GetContextWithTimeout implements DocumentDBAccessor.
func (d *defaultDocumentDBAccessor) GetContextWithTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	// MongoDB Goドライバーは contextパッケージを使用してタイムアウトを設定する
	// https://www.mongodb.com/ja-jp/docs/drivers/go/current/fundamentals/context/
	timeout := d.config.GetInt(DOCUMENTDB_CONNECTION_TIMEOUT_NAME, 3)
	return apcontext.GetContextWithTimeout(ctx, timeout)
}

// getCustomTLSConfig は、指定された公開鍵証明書をもとに、カスタムのTLS設定を取得します。
func getCustomTLSConfig(caFile string) (*tls.Config, error) {
	tlsConfig := new(tls.Config)
	certs, err := os.ReadFile(caFile)
	if err != nil {
		return tlsConfig, errors.Wrap(err, "pemファイルの読み込みに失敗しました")
	}
	// ルート証明書として指定の証明書を追加
	tlsConfig.RootCAs = x509.NewCertPool()
	ok := tlsConfig.RootCAs.AppendCertsFromPEM(certs)

	if !ok {
		return tlsConfig, errors.Newf("pemファイル:%sの解析に失敗しました", caFile)
	}

	return tlsConfig, nil
}
