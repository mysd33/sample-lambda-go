// serviceのパッケージ
package service

import (
	"app/internal/pkg/entity"
	"app/internal/pkg/repository"

	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/logging"
)

// TodoService は、Todo業務のServiceインタフェースです。
type TodoService interface {
	// Find は、todoIdのTodoを照会します。
	Find(todoId string) (*entity.Todo, error)
	// Register は、タイトルtodoTitleのTodoを登録します。
	Register(todoTitle string) (*entity.Todo, error)
}

// New は、TodoServiceを作成します。
func New(log logging.Logger,
	config *config.Config,
	repository repository.TodoRepository,
) TodoService {
	return &todoServiceImpl{log: log, config: config, repository: repository}
}

// todoServiceImpl TodoServiceを実装する構造体です。
type todoServiceImpl struct {
	log        logging.Logger
	config     *config.Config
	repository repository.TodoRepository
}

func (ts *todoServiceImpl) Find(todoId string) (*entity.Todo, error) {
	return ts.repository.GetTodo(todoId)
}

func (ts *todoServiceImpl) Register(todoTitle string) (*entity.Todo, error) {
	ts.log.Debug("TodoTitle=%s", todoTitle)
	// TODO: メッセージIDを使ったログのメソッドの例に修正
	// ts.log.Info(code.I_EX_0001, todoTitle)

	// 業務エラーの例
	// if (...) {
	//   return nil, errors.NewBusinessError(nil, code.W_EX_8001, "xxxx")
	// }

	todo := entity.Todo{Title: todoTitle}

	return ts.repository.PutTodo(&todo)
}
