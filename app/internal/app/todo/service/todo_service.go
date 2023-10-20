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
	// Regist は、タイトルtodoTitleのTodoを登録します。
	Regist(todoTitle string) (*entity.Todo, error)
}

// New は、TodoServiceを作成します。
func New(log logging.Logger,
	config *config.Config,
	repository repository.TodoRepository,
) TodoService {
	return &TodoServiceImpl{log: log, config: config, repository: repository}
}

// TodoServiceImpl TodoServiceを実装する構造体です。
type TodoServiceImpl struct {
	log        logging.Logger
	config     *config.Config
	repository repository.TodoRepository
}

func (ts *TodoServiceImpl) Find(todoId string) (*entity.Todo, error) {
	return ts.repository.GetTodo(todoId)
}

func (ts *TodoServiceImpl) Regist(todoTitle string) (*entity.Todo, error) {
	ts.log.Info("TodoTitle=%s", todoTitle)
	// TODO: メッセージIDを使ったログのメソッドの例に修正
	// ts.log.Info(code.I_EX_0001, todoTitle)

	// 業務エラーの例
	// if (...) {
	//   return nil, errors.NewBusinessError(nil, code.W_EX_8001, "xxxx")
	// }

	todo := entity.Todo{}
	todo.Title = todoTitle
	return ts.repository.PutTodo(&todo)
}
