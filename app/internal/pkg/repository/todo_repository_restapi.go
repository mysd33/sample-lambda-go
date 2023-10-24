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
func (*todoRepositoryImplByRestAPI) PutTodo(todo *entity.Todo) (*entity.Todo, error) {
	// TODO:実装
	panic("unimplemented")
}
