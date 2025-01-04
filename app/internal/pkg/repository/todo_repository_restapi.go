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
	TODO_API_BASE_URL_NAME = "TODO_API_BASE_URL"
)

// NewTodoRepositoryForRestAPI は、REST APIのためのTodoRepository実装を作成します。
func NewTodoRepositoryForRestAPI(httpClient httpclient.HTTPClient, logger logging.Logger, config config.Config) TodoRepository {
	return &todoRepositoryImplByRestAPI{httpClient: httpClient, logger: logger, config: config}
}

type todoRepositoryImplByRestAPI struct {
	httpClient httpclient.HTTPClient
	logger     logging.Logger
	config     config.Config
}

// FindOne implements TodoRepository.
func (tr *todoRepositoryImplByRestAPI) FindOne(todoId string) (*model.Todo, error) {
	baseUrl, found := tr.config.GetWithContains(TODO_API_BASE_URL_NAME)
	if !found {
		return nil, errors.NewSystemError(fmt.Errorf("TODO_API_BASE_URLがありません"), message.E_EX_9001)
	}
	url := fmt.Sprintf("%s/todo-api/v1/todo/%s", baseUrl, todoId)
	tr.logger.Debug("url:%s", url)

	response, err := tr.httpClient.Get(url, nil, nil)
	if err != nil {
		return nil, errors.NewSystemError(err, message.E_EX_9004)
	}

	if response.StatusCode != 200 {
		// 200以外の処理
		var data ErrorResponse
		// エラーレスポンスデータをアンマーシャル
		if err = json.Unmarshal(response.Body, &data); err != nil {
			return nil, errors.NewSystemError(err, message.E_EX_9001)
		}
		// この例では、システムエラーとして扱っているが、実際にはエラーコードに応じて業務エラーとして扱うケースもある
		return nil, errors.NewSystemError(err, message.E_EX_9005, response.StatusCode, data.Code, data.Detail)
	}

	var todo model.Todo
	if err = json.Unmarshal(response.Body, &todo); err != nil {
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}
	return &todo, nil
}

// CreateOne implements TodoRepository.
func (tr *todoRepositoryImplByRestAPI) CreateOne(todo *model.Todo) (*model.Todo, error) {
	baseUrl, found := tr.config.GetWithContains(TODO_API_BASE_URL_NAME)
	if !found {
		return nil, errors.NewSystemError(fmt.Errorf("TODO_API_BASE_URLがありません"), message.E_EX_9001)
	}
	url := fmt.Sprintf("%s/todo-api/v1/todo", baseUrl)
	tr.logger.Debug("url:%s", url)
	// リクエストデータをアンマーシャル
	data, err := json.MarshalIndent(todo, "", "    ")
	if err != nil {
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}
	// REST APIの呼び出し
	response, err := tr.httpClient.Post(url, nil, data)
	if err != nil {
		return nil, errors.NewSystemError(err, message.E_EX_9004)
	}

	if response.StatusCode != 200 {
		// 200以外の処理
		var data ErrorResponse
		// エラーレスポンスデータをアンマーシャル
		if err = json.Unmarshal(response.Body, &data); err != nil {
			return nil, errors.NewSystemError(err, message.E_EX_9001)
		}
		// この例では、システムエラーとして扱っているが、実際にはエラーコードに応じて業務エラーとして扱うケースもある
		return nil, errors.NewSystemError(err, message.E_EX_9005, response.StatusCode, data.Code, data.Detail)
	}

	// レスポンスデータをアンマーシャル
	var newTodo model.Todo
	if err = json.Unmarshal(response.Body, &newTodo); err != nil {
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}
	return &newTodo, nil
}

// CreateOneTx implements TodoRepository.
func (*todoRepositoryImplByRestAPI) CreateOneTx(todo *model.Todo) (*model.Todo, error) {
	panic("unimplemented")
}
