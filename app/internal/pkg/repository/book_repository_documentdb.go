package repository

import (
	"app/internal/pkg/message"
	"app/internal/pkg/model"

	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/documentdb"
	"example.com/appbase/pkg/errors"
	"example.com/appbase/pkg/logging"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	// コレクション名（テーブル名相当）
	BOOK_COLLECTION_NAME = "books"
)

// NewBookRepositoryForDocumentDB は、BookRepositoryを作成します。
func NewBookRepositoryForDocumentDB(logger logging.Logger, config config.Config,
	documentDBAccessor documentdb.DocumentDBAccessor) BookRepository {
	// Collectionの取得
	collection := documentDBAccessor.GetMongoDatabase().Collection(BOOK_COLLECTION_NAME)
	return &bookRepositoryImplByDocumentDB{
		documentDBAccessor: documentDBAccessor,
		collection:         collection,
		logger:             logger,
		config:             config,
	}
}

// bookRepositoryImplByDocumentDB は、BookRepositoryのDocumentDB（MongoDB）実装構造体です。
type bookRepositoryImplByDocumentDB struct {
	documentDBAccessor documentdb.DocumentDBAccessor
	collection         *mongo.Collection
	logger             logging.Logger
	config             config.Config
}

// （参考）DocumentDBおよびMongoDBを使ったプログラミングの知識は、以下のURLを参照
// https://docs.aws.amazon.com/ja_jp/documentdb/latest/developerguide/get-started-guide.html
// https://docs.aws.amazon.com/ja_jp/documentdb/latest/developerguide/connect_programmatically.html
// https://www.mongodb.com/ja-jp/docs/drivers/go/current/

// CreateOne implements BookRepository.
func (b *bookRepositoryImplByDocumentDB) CreateOne(book *model.Book) (*model.Book, error) {
	// ドキュメントの登録
	//　(参考)https://www.mongodb.com/ja-jp/docs/drivers/go/current/usage-examples/insertOne/

	// タイムアウトの設定
	ctx, cancel := b.documentDBAccessor.GetDefaultContextWithTimeout()
	defer cancel()
	result, err := b.collection.InsertOne(ctx, &book)
	if err != nil {
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}
	resultID := result.InsertedID
	book.ObjectID = resultID.(primitive.ObjectID)
	return book, nil
}

// FindSomeByCriteria implements BookRepository.
func (b *bookRepositoryImplByDocumentDB) FindSomeByCriteria(criteria *BookCriteria) ([]model.Book, error) {
	// タイムアウトの設定
	ctx, cancel := b.documentDBAccessor.GetDefaultContextWithTimeout()
	defer cancel()

	// 複数ドキュメントの検索
	// （参考）https://www.mongodb.com/ja-jp/docs/drivers/go/current/usage-examples/find/
	cur, err := b.collection.Find(ctx, &criteria)

	// TODO: 正規表現による部分一致での検索条件構築
	// （参考）https://www.mongodb.com/ja-jp/docs/drivers/go/master/fundamentals/crud/read-operations/query-document/#evaluation

	if err != nil {
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}
	// 処理結果のドキュメントのデコード
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
