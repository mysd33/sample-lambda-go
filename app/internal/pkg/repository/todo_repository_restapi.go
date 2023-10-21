// repositoryのパッケージ
package repository

import (
	"app/internal/pkg/code"
	"app/internal/pkg/entity"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"example.com/appbase/pkg/errors"
	"example.com/appbase/pkg/logging"
)

// NewTodoRepositoryForRestAPI は、REST APIのためのTodoRepository実装を作成します。
func NewTodoRepositoryForRestAPI(log logging.Logger) TodoRepository {
	return &todoRepositoryImplByRestAPI{log: log}
}

type todoRepositoryImplByRestAPI struct {
	log logging.Logger
}

// GetTodo implements TodoRepository.
func (tr *todoRepositoryImplByRestAPI) GetTodo(todoId string) (*entity.Todo, error) {
	baseUrl := os.Getenv("TODO_API_BASE_URL")
	url := fmt.Sprintf("%s/todo-api/v1/todo/%s", baseUrl, todoId)
	tr.log.Debug("url:%s", url)

	// TODO: AP基盤機能化
	response, err := http.Get(url)
	if err != nil {
		return nil, errors.NewSystemError(err, code.E_EX_9001)
	}
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, errors.NewSystemError(err, code.E_EX_9001)
	}

	var todo entity.Todo
	if err = json.Unmarshal(data, &todo); err != nil {
		return nil, errors.NewSystemError(err, code.E_EX_9001)
	}
	return &todo, nil
}

// PutTodo implements TodoRepository.
func (*todoRepositoryImplByRestAPI) PutTodo(todo *entity.Todo) (*entity.Todo, error) {
	// TODO:実装
	panic("unimplemented")
}
