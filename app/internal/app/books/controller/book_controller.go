// controllerのパッケージ
package controller

import (
	"app/internal/app/books/service"
	"app/internal/pkg/message"
	"app/internal/pkg/model"
	"app/internal/pkg/repository"

	"example.com/appbase/pkg/errors"
	"example.com/appbase/pkg/logging"
	"github.com/gin-gonic/gin"
)

// RequestFind は、書籍検索のREST APIで受け取るリクエストデータの構造体です。
type RequestFind struct {
	// Title は、書籍のタイトルです。
	Title string `label:"タイトル" form:"title"`
	// Author は、書籍の著者です。
	Author string `label:"著者" form:"author"`
	// Publisher は、書籍の出版社です。
	Publisher string `label:"出版社" form:"publisher"`
}

// RequestRegister は、書籍登録のREST APIで受け取るリクエストデータの構造体です。
type RequestRegister struct {
	// Title は、書籍のタイトルです。
	Title string `label:"タイトル" json:"title" binding:"required"`
	// Author は、書籍の著者です。
	Author string `label:"著者" json:"author" binding:"required"`
	// Publisher は、書籍の出版社です。
	Publisher string `label:"出版社" json:"publisher"`
	// PublishedDate は、書籍の発売日です。
	PublishedDate string `label:"発売日" json:"published_date" binding:"datetime=2006-01-02"`
	// ISBNは、書籍のISBNです。
	ISBN string `label:"ISBN" json:"isbn"`
}

// BookController は、書籍管理業務のControllerインタフェースです。
type BookController interface {
	// FindByCriteria は、クエリパラメータで指定された検索条件に合致する書籍を検索します。
	FindByCriteria(ctx *gin.Context) (any, error)
	// Register は、リクエストデータで受け取った書籍を登録します。
	Register(ctx *gin.Context) (any, error)
}

// New は、TodoControllerを作成します。
func New(logger logging.Logger,
	service service.BookService,
) BookController {
	return &bookControllerImpl{
		logger:  logger,
		service: service,
	}
}

// bookControllerImpl は、TodoControllerを実装する構造体です。
type bookControllerImpl struct {
	logger  logging.Logger
	service service.BookService
}

// FindByCriteria implements BookController.
func (b *bookControllerImpl) FindByCriteria(ctx *gin.Context) (any, error) {
	var request RequestFind
	if err := ctx.ShouldBindQuery(&request); err != nil {
		return nil, errors.NewValidationErrorWithCause(err, message.W_EX_5001)
	}
	bookCriteria := &repository.BookCriteria{
		Title:     request.Title,
		Author:    request.Author,
		Publisher: request.Publisher,
	}
	return b.service.FindByCriteria(bookCriteria)
}

// Register implements BookController.
func (b *bookControllerImpl) Register(ctx *gin.Context) (any, error) {
	var request RequestRegister
	if err := ctx.ShouldBindJSON(&request); err != nil {
		return nil, errors.NewValidationErrorWithCause(err, message.W_EX_5001)
	}
	book := &model.Book{
		Title:         request.Title,
		Author:        request.Author,
		Publisher:     request.Publisher,
		PublishedDate: request.PublishedDate,
		ISBN:          request.ISBN,
	}
	return b.service.Register(book)
}
