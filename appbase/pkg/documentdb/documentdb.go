/*
documentdb パッケージは、DocumentDB（MongoDB）アクセス機能を提供するパッケージです。
*/
package documentdb

import (
	"context"
	"fmt"

	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/logging"
	"github.com/cockroachdb/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	DOCUMENTDB_LOCAL_ENDPOINT_NAME = "DOCUMENTDB_LOCAL_ENDPOINT"
	DOCUMENTDB_DBNAME_NAME         = "DOCUMENTDB_DB_NAME"
	DOCUMENTDB_USERNAME_NAME       = "DOCUMENTDB_USERNAME_NAME"
	DOCUMENTDB_PASSWORD_NAME       = "DOCUMENTDB_PASSWORD"
)

// DocumentDBAccessor は、DocumentDB（MongoDB）アクセス機能を提供するインターフェースです。
type DocumentDBAccessor interface {
	// GetMongoClient は、MongoDBのクライアントを取得します。
	GetMongoClient() *mongo.Client
	// GetMongoDatabase は、MongoDBのデータベースを取得します。
	GetMongoDatabase() *mongo.Database
}

// defaultDocumentDBAccessor は、DocumentDBAccessorインタフェースを実装する構造体です。
type defaultDocumentDBAccessor struct {
	client   *mongo.Client
	database *mongo.Database
}

func NewDocumentDBAccessor(config config.Config, logger logging.Logger) (DocumentDBAccessor, error) {
	// TODO: 接続情報のプロパティのデフォルト値取得処理の見直し
	documentdbEndpoint := config.Get(DOCUMENTDB_LOCAL_ENDPOINT_NAME, "host.docker.internal:27017")
	dbName := config.Get(DOCUMENTDB_DBNAME_NAME, "sampledb")
	userName := config.Get(DOCUMENTDB_USERNAME_NAME, "root")
	password := config.Get(DOCUMENTDB_PASSWORD_NAME, "password")

	// ローカルのMongoDBの接続文字列作成
	// （参考）https://qiita.com/chenglin/items/ecf6f67e8f80c4750204
	connectionStringTemplate := "mongodb://%s:%s@%s/%s?authSource=admin"
	connectionString := fmt.Sprintf(connectionStringTemplate, userName, password, documentdbEndpoint, dbName)

	// TODO: DocumentDBの場合の接続文字列作成との切替
	//connectionStringTemplate := "mongodb://%s:%s@%s/%s?tls=true&replicaSet=rs0&readpreference=%s"

	logger.Debug("接続文字列：%s", connectionString)

	// ローカルのMongoDBに接続時の処理
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(connectionString))
	// TODO: DocumentDBの場合でのTLS接続との切替
	//client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(connectionString).SetTLSConfig(tlsConfig))

	if err != nil {
		return nil, errors.Wrap(err, "DocumentDB(MongoDB)への接続に失敗しました")
	}

	return &defaultDocumentDBAccessor{
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
