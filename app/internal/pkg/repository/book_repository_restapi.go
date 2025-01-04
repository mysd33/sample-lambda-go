// repositoryのパッケージ
package repository

import (
	"app/internal/pkg/message"
	"app/internal/pkg/model"
	"encoding/json"
	"fmt"

	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/errors"
	"example.com/appbase/pkg/httpclient"
	"example.com/appbase/pkg/logging"
)

const (
	BOOKS_API_BASE_URL_NAME = "USERS_API_BASE_URL"
)

func NewBookRepositoryForRestAPI(httpClient httpclient.HTTPClient, logger logging.Logger, config config.Config) BookRepository {
	return &bookRepositoryImplByRestAPI{
		httpClient: httpClient,
		logger:     logger,
		config:     config,
	}
}

type bookRepositoryImplByRestAPI struct {
	httpClient httpclient.HTTPClient
	logger     logging.Logger
	config     config.Config
}

// FindSomeByCriteria implements BookRepository.
func (b *bookRepositoryImplByRestAPI) FindSomeByCriteria(criteria *BookCriteria) ([]model.Book, error) {
	baseUrl, found := b.config.GetWithContains(BOOKS_API_BASE_URL_NAME)
	if !found {
		return nil, errors.NewSystemError(fmt.Errorf("BOOKS_API_BASE_URLがありません"), message.E_EX_9001)
	}
	url := fmt.Sprintf("%s/books-api/v1/books", baseUrl)
	b.logger.Debug("url:%s", url)
	params := map[string]string{}
	if criteria != nil {
		if criteria.Title != "" {
			params["title"] = criteria.Title
		}
		if criteria.Author != "" {
			params["author"] = criteria.Author
		}
		if criteria.Publisher != "" {
			params["publisher"] = criteria.Publisher
		}
	}
	response, err := b.httpClient.Get(url, nil, params)
	if err != nil {
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}
	var books []model.Book
	if err = json.Unmarshal(response.Body, &books); err != nil {
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}
	if response.StatusCode != 200 {
		// TODO: 200以外の処理
		return nil, errors.NewBusinessError(message.W_EX_8001, "xxxx")
	}
	return books, nil
}

// CreateOne implements BookRepository.
func (b *bookRepositoryImplByRestAPI) CreateOne(book *model.Book) (*model.Book, error) {
	baseUrl, found := b.config.GetWithContains(BOOKS_API_BASE_URL_NAME)
	if !found {
		return nil, errors.NewSystemError(fmt.Errorf("BOOKS_API_BASE_URLがありません"), message.E_EX_9001)
	}
	url := fmt.Sprintf("%s/books-api/v1/books", baseUrl)
	b.logger.Debug("url:%s", url)
	// リクエストデータをアンマーシャル
	data, err := json.MarshalIndent(book, "", "    ")
	if err != nil {
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}
	// REST APIの呼び出し
	response, err := b.httpClient.Post(url, nil, data)
	if err != nil {
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}
	if response.StatusCode != 200 {
		// TODO: 200以外の処理
		return nil, errors.NewBusinessError(message.W_EX_8001, "xxxx")
	}
	// レスポンスデータをアンマーシャル
	var newBook model.Book
	if err = json.Unmarshal(response.Body, &newBook); err != nil {
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}
	return &newBook, nil
}
