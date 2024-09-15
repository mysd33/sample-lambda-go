// serviceのパッケージ
package service

import (
	"app/internal/pkg/message"
	"app/internal/pkg/model"
	"app/internal/pkg/repository"

	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/logging"
)

// TodoService は、Todo業務のServiceインタフェースです。
type TodoService interface {
	// Find は、todoIdのTodoを照会します。
	Find(todoId string) (*model.Todo, error)
	// Register は、タイトルtodoTitleのTodoを登録します。
	Register(todoTitle string) (*model.Todo, error)
	// RegisterTx は、タイトルtodoTitleのTodoをトランザクションを使って登録します。
	// DynamoDBトランザクションを使った動作確認用に定義したもの
	RegisterTx(todoTitle string) (*model.Todo, error)
}

// New は、TodoServiceを作成します。
func New(logger logging.Logger,
	config config.Config,
	repository repository.TodoRepository,
) TodoService {
	return &todoServiceImpl{logger: logger, config: config, repository: repository}
}

// todoServiceImpl TodoServiceを実装する構造体です。
type todoServiceImpl struct {
	logger     logging.Logger
	config     config.Config
	repository repository.TodoRepository
}

// Find implements TodoService.
func (ts *todoServiceImpl) Find(todoId string) (*model.Todo, error) {
	return ts.repository.FindOne(todoId)
}

// Register implements TodoService.
func (ts *todoServiceImpl) Register(todoTitle string) (*model.Todo, error) {
	// デバッグログの例
	ts.logger.Debug("TodoTitle=%s", todoTitle)
	// メッセージIDを使った情報ログの例
	ts.logger.Info(message.I_EX_0001, todoTitle)

	// 業務エラーの例
	// if (...) {
	//   return nil, errors.NewBusinessError(nil, message.W_EX_8001, "xxxx")
	// }

	todo := model.Todo{Title: todoTitle}

	return ts.repository.CreateOne(&todo)
}

// RegisterTx implements TodoService.
func (ts *todoServiceImpl) RegisterTx(todoTitle string) (*model.Todo, error) {
	todo := model.Todo{Title: todoTitle}
	return ts.repository.CreateOneTx(&todo)
}
