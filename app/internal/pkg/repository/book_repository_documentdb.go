package repository

import (
	"app/internal/pkg/message"
	"app/internal/pkg/model"
	"context"
	"fmt"

	"example.com/appbase/pkg/apcontext"
	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/errors"
	"example.com/appbase/pkg/logging"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TODO: DocumentDB(MongoDB)アクセス機能のソフトウェアフレームワーク化

const (
	DOCUMENTDB_LOCAL_ENDPOINT_NAME = "DOCUMENTDB_LOCAL_ENDPOINT"
	DOCUMENTDB_DBNAME_NAME         = "DOCUMENTDB_DB_NAME"
	DOCUMENTDB_USERNAME_NAME       = "DOCUMENTDB_USERNAME_NAME"
	DOCUMENTDB_PASSWORD_NAME       = "DOCUMENTDB_PASSWORD"
	BOOK_COLLECTION_NAME           = "books"
)

// NewBookRepositoryForDocumentDB は、BookRepositoryを作成します。
func NewBookRepositoryForDocumentDB(logger logging.Logger, config config.Config) BookRepository {
	// TODO: 接続情報のプロパティのデフォルト値取得見直し
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

	// TODO: DocumentDBの場合でのTLS接続との切替
	//client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(connectionString).SetTLSConfig(tlsConfig))

	// ローカルのMongoDBに接続時の処理
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(connectionString))
	if err != nil {
		//TODO: フレームワーク機能化した後、エラーハンドリング
		panic(err)
	}

	// Collectionの取得
	collection := client.Database(dbName).Collection(BOOK_COLLECTION_NAME)
	return &bookRepositoryImplByDocumentDB{
		collection: collection,
		logger:     logger,
		config:     config,
	}
}

type bookRepositoryImplByDocumentDB struct {
	collection *mongo.Collection
	logger     logging.Logger
	config     config.Config
}

// CreateOne implements BookRepository.
func (b *bookRepositoryImplByDocumentDB) CreateOne(book *model.Book) (*model.Book, error) {
	result, err := b.collection.InsertOne(apcontext.Context, &book)
	if err != nil {
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}
	resultID := result.InsertedID
	book.ObjectID = resultID.(primitive.ObjectID)
	return book, nil
}

// FindSomeByCriteria implements BookRepository.
func (b *bookRepositoryImplByDocumentDB) FindSomeByCriteria(criteria *BookCriteria) ([]model.Book, error) {
	ctx := apcontext.Context
	// TODO: 部分一致の対応
	cur, err := b.collection.Find(ctx, &criteria)

	if err != nil {
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}
	var books []model.Book
	for cur.Next(ctx) {
		var book model.Book
		err := cur.Decode(&book)
		if err != nil {
			return nil, errors.NewSystemError(err, message.E_EX_9001)
		}
		books = append(books, book)
	}
	if err := cur.Err(); err != nil {
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}
	return books, nil
}
