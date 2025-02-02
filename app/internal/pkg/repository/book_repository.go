// repositoryのパッケージ
package repository

import (
	"app/internal/pkg/model"
	"strings"
)

// BookCriteria は、書籍の検索条件を表す構造体です。
type BookCriteria struct {
	// （参考） bsonの構造タグを利用した、MongoDBのフィールド名の指定
	// https://www.mongodb.com/ja-jp/docs/drivers/go/current/usage-examples/findOne/

	// https://www.mongodb.com/ja-jp/docs/drivers/go/current/fundamentals/bson/#struct-tags
	Title string `bson:"title,omitempty"`
	// Author は、書籍の著者です。
	Author string `bson:"author,omitempty"`
	// Publisher は、書籍の出版社です。
	Publisher string `bson:"publisher,omitempty"`
}

// BookRepository は、書籍を管理するRepositoryインタフェースです。
type BookRepository interface {
	// FindSomeByCriteria は、条件に合致する書籍を取得します。
	FindSomeByCriteria(criteria *BookCriteria) ([]model.Book, error)
	// CreateOne は、指定されたユーザを登録します。
	CreateOne(book *model.Book) (*model.Book, error)
}

// NewBookRepositoryStub は、BookRepositoryを作成します。
func NewBookRepositoryStub() BookRepository {
	return &bookRepositoryStub{}
}

// bookRepositoryStub は、BookRepositoryのスタブです。
type bookRepositoryStub struct {
	books []model.Book
}

// FindSomeByCriteria implements BookRepository.
func (b *bookRepositoryStub) FindSomeByCriteria(criteria *BookCriteria) ([]model.Book, error) {
	var results []model.Book
	if criteria == nil {
		return b.books, nil
	}
	for _, book := range b.books {
		if criteria.Title != "" && !strings.Contains(book.Title, criteria.Title) {
			continue
		}
		if criteria.Author != "" && !strings.Contains(book.Author, criteria.Author) {
			continue
		}
		if criteria.Publisher != "" && !strings.Contains(book.Publisher, criteria.Publisher) {
			continue
		}
		results = append(results, book)
	}
	return results, nil
}

// CreateOne implements BookRepository.
func (b *bookRepositoryStub) CreateOne(book *model.Book) (*model.Book, error) {
	b.books = append(b.books, *book)
	return book, nil
}
