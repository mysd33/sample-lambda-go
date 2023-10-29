// repositoryのパッケージ
package repository

import (
	"app/internal/pkg/code"
	"app/internal/pkg/entity"
	"encoding/json"
	"fmt"
	"os"

	"example.com/appbase/pkg/errors"
	"example.com/appbase/pkg/httpclient"
	"example.com/appbase/pkg/logging"
)

// NewTodoRepositoryForRestAPI は、REST APIのためのTodoRepository実装を作成します。
func NewTodoRepositoryForRestAPI(httpClient httpclient.HttpClient, log logging.Logger) TodoRepository {
	return &todoRepositoryImplByRestAPI{httpClient: httpClient, log: log, baseUrl: os.Getenv("TODO_API_BASE_URL")}
}

type todoRepositoryImplByRestAPI struct {
	httpClient httpclient.HttpClient
	log        logging.Logger
	baseUrl    string
}

// GetTodo implements TodoRepository.
func (tr *todoRepositoryImplByRestAPI) GetTodo(todoId string) (*entity.Todo, error) {
	url := fmt.Sprintf("%s/todo-api/v1/todo/%s", tr.baseUrl, todoId)
	tr.log.Debug("url:%s", url)

	response, err := tr.httpClient.Get(url, nil, nil)
	if err != nil {
		return nil, errors.NewSystemError(err, code.E_EX_9001)
	}

	var todo entity.Todo
	if err = json.Unmarshal(response.Body, &todo); err != nil {
		return nil, errors.NewSystemError(err, code.E_EX_9001)
	}
	return &todo, nil
}

// PutTodo implements TodoRepository.
func (tr *todoRepositoryImplByRestAPI) PutTodo(todo *entity.Todo) (*entity.Todo, error) {
	url := fmt.Sprintf("%s/todo-api/v1/todo", tr.baseUrl)
	tr.log.Debug("url:%s", url)
	// リクエストデータをアンマーシャル
	data, err := json.MarshalIndent(todo, "", "    ")
	if err != nil {
		return nil, errors.NewSystemError(err, code.E_EX_9001)
	}
	// REST APIの呼び出し
	response, err := tr.httpClient.Post(url, nil, data)
	if err != nil {
		return nil, errors.NewSystemError(err, code.E_EX_9001)
	}
	if response.StatusCode != 200 {
		// TODO: 200以外の処理
		return nil, errors.NewBusinessError(code.W_EX_8001, "xxxx")
	}
	// レスポンスデータをアンマーシャル
	var newTodo entity.Todo
	if err = json.Unmarshal(response.Body, &newTodo); err != nil {
		return nil, errors.NewSystemError(err, code.E_EX_9001)
	}
	return &newTodo, nil
}
