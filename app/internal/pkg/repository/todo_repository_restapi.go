// repositoryのパッケージ
package repository

import (
	"app/internal/pkg/entity"
	"app/internal/pkg/message"
	"encoding/json"
	"fmt"

	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/errors"
	"example.com/appbase/pkg/httpclient"
	"example.com/appbase/pkg/logging"
)

const (
	TODO_API_BASE_URL = "TODO_API_BASE_URL"
)

// NewTodoRepositoryForRestAPI は、REST APIのためのTodoRepository実装を作成します。
func NewTodoRepositoryForRestAPI(httpClient httpclient.HttpClient, log logging.Logger, config config.Config) TodoRepository {
	return &todoRepositoryImplByRestAPI{httpClient: httpClient, log: log, config: config}
}

type todoRepositoryImplByRestAPI struct {
	httpClient httpclient.HttpClient
	log        logging.Logger
	config     config.Config
}

// FindOne implements TodoRepository.
func (tr *todoRepositoryImplByRestAPI) FindOne(todoId string) (*entity.Todo, error) {
	baseUrl, found := tr.config.GetWithContains(TODO_API_BASE_URL)
	if !found {
		return nil, errors.NewSystemError(fmt.Errorf("TODO_API_BASE_URLがありません"), message.E_EX_9001)
	}
	url := fmt.Sprintf("%s/todo-api/v1/todo/%s", baseUrl, todoId)
	tr.log.Debug("url:%s", url)

	response, err := tr.httpClient.Get(url, nil, nil)
	if err != nil {
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}

	var todo entity.Todo
	if err = json.Unmarshal(response.Body, &todo); err != nil {
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}
	return &todo, nil
}

// CreateOne implements TodoRepository.
func (tr *todoRepositoryImplByRestAPI) CreateOne(todo *entity.Todo) (*entity.Todo, error) {
	baseUrl, found := tr.config.GetWithContains(TODO_API_BASE_URL)
	if !found {
		return nil, errors.NewSystemError(fmt.Errorf("TODO_API_BASE_URLがありません"), message.E_EX_9001)
	}
	url := fmt.Sprintf("%s/todo-api/v1/todo", baseUrl)
	tr.log.Debug("url:%s", url)
	// リクエストデータをアンマーシャル
	data, err := json.MarshalIndent(todo, "", "    ")
	if err != nil {
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}
	// REST APIの呼び出し
	response, err := tr.httpClient.Post(url, nil, data)
	if err != nil {
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}
	if response.StatusCode != 200 {
		// TODO: 200以外の処理
		return nil, errors.NewBusinessError(message.W_EX_8001, "xxxx")
	}
	// レスポンスデータをアンマーシャル
	var newTodo entity.Todo
	if err = json.Unmarshal(response.Body, &newTodo); err != nil {
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}
	return &newTodo, nil
}

// CreateOneTx implements TodoRepository.
func (*todoRepositoryImplByRestAPI) CreateOneTx(todo *entity.Todo) (*entity.Todo, error) {
	panic("unimplemented")
}
