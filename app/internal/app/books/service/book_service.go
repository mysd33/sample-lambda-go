// serviceのパッケージ
package service

import (
	"app/internal/pkg/model"
	"app/internal/pkg/repository"

	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/logging"
)

// BookService は、書籍業務のServiceインタフェースです。
type BookService interface {
	// FindByCriteria は、条件に合致する書籍を取得します。
	FindByCriteria(criteria *repository.BookCriteria) ([]model.Book, error)
	// Register は、書籍を登録します。
	Register(book *model.Book) (*model.Book, error)
}

// New は、BookServiceを作成します。
func New(logger logging.Logger,
	config config.Config,
	repository repository.BookRepository,
) BookService {
	return &bookServiceImpl{
		logger:     logger,
		config:     config,
		repository: repository,
	}
}

// bookServiceImpl BookServiceを実装する構造体です。
type bookServiceImpl struct {
	logger     logging.Logger
	config     config.Config
	repository repository.BookRepository
}

// FindByCriteria implements BookService.
func (b *bookServiceImpl) FindByCriteria(criteria *repository.BookCriteria) ([]model.Book, error) {
	return b.repository.FindSomeByCriteria(criteria)
}

// Register implements BookService.
func (b *bookServiceImpl) Register(book *model.Book) (*model.Book, error) {
	return b.repository.CreateOne(book)
}
