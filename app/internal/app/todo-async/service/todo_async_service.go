// serviceのパッケージ
package service

import (
	"app/internal/pkg/entity"
	"app/internal/pkg/message"
	"app/internal/pkg/repository"

	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/logging"
)

// TodoAsyncService は、Todoの非同期処理を管理するServiceインタフェースです。
type TodoAsyncService interface {
	// RegisterTodosAsync は、Todoを非同期で登録します。
	RegisterTodosAsync(asyncMesssage entity.AsyncMessage) error
}

// NewTodoAsyncService は、TodoAsyncServiceを生成します。
func New(log logging.Logger,
	config config.Config,
	tempRepository repository.TempRepository,
	todoRepository repository.TodoRepository) TodoAsyncService {
	return &todoAsyncServiceImpl{log: log,
		config:         config,
		tempRepository: tempRepository,
		todoRepository: todoRepository,
	}
}

type todoAsyncServiceImpl struct {
	log            logging.Logger
	config         config.Config
	tempRepository repository.TempRepository
	todoRepository repository.TodoRepository
}

// RegisterTodosAsync implements TodoAsyncService.
func (ts *todoAsyncServiceImpl) RegisterTodosAsync(asyncMesssage entity.AsyncMessage) error {
	if asyncMesssage.TempId == "" {
		// TempIdが空の場合は、何もしない
		return nil
	}

	// TODO: tempテーブルのIDをもとに、todoTitles情報から取得して登録するように変更
	temp, err := ts.tempRepository.FindOne(asyncMesssage.TempId)
	if err != nil {
		return err
	}
	ts.log.Debug("temp: %+v", temp)
	todoTitles := asyncMesssage.TodoTitles
	for _, v := range todoTitles {
		todo := entity.Todo{Title: v}
		newTodo, err := ts.todoRepository.CreateOneTx(&todo)
		if err != nil {
			return err
		}
		ts.log.Info(message.I_EX_0003, newTodo.ID)
	}
	return nil
}
