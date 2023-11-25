// serviceのパッケージ
package service

import (
	"app/internal/pkg/entity"
	"app/internal/pkg/message"
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
	// RegisterTx は、タイトルtodoTitleのTodoをトランザクションを使って登録します。
	// DynamoDBトランザクションを使った動作確認用に定義したもの
	RegisterTx(todoTitle string) (*entity.Todo, error)
}

// New は、TodoServiceを作成します。
func New(log logging.Logger,
	config config.Config,
	repository repository.TodoRepository,
) TodoService {
	return &todoServiceImpl{log: log, config: config, repository: repository}
}

// todoServiceImpl TodoServiceを実装する構造体です。
type todoServiceImpl struct {
	log        logging.Logger
	config     config.Config
	repository repository.TodoRepository
}

// Find implements TodoService.
func (ts *todoServiceImpl) Find(todoId string) (*entity.Todo, error) {
	return ts.repository.GetTodo(todoId)
}

// Register implements TodoService.
func (ts *todoServiceImpl) Register(todoTitle string) (*entity.Todo, error) {
	// デバッグログの例
	ts.log.Debug("TodoTitle=%s", todoTitle)
	// メッセージIDを使った情報ログの例
	ts.log.Info(message.I_EX_0001, todoTitle)

	// 業務エラーの例
	// if (...) {
	//   return nil, errors.NewBusinessError(nil, message.W_EX_8001, "xxxx")
	// }

	todo := entity.Todo{Title: todoTitle}

	return ts.repository.PutTodo(&todo)
}

// RegisterTx implements TodoService.
func (ts *todoServiceImpl) RegisterTx(todoTitle string) (*entity.Todo, error) {
	todo := entity.Todo{Title: todoTitle}
	return ts.repository.PutTodoTx(&todo)
}
