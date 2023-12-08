// serviceのパッケージ
package service

import (
	"app/internal/pkg/entity"
	"app/internal/pkg/message"
	"app/internal/pkg/repository"

	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/logging"
)

type TodoAsyncService interface {
	RegisterTodosAsync(todoTitles []string) error
}

func New(log logging.Logger,
	config config.Config,
	repository repository.TodoRepository) TodoAsyncService {
	return &todoAsyncServiceImpl{log: log,
		config:     config,
		repository: repository,
	}
}

type todoAsyncServiceImpl struct {
	log        logging.Logger
	config     config.Config
	repository repository.TodoRepository
}

// RegisterTodosAsync implements TodoAsyncService.
func (ts *todoAsyncServiceImpl) RegisterTodosAsync(todoTitles []string) error {
	for _, v := range todoTitles {
		todo := entity.Todo{Title: v}
		newTodo, err := ts.repository.CreateOneTx(&todo)
		if err != nil {
			return err
		}
		ts.log.Info(message.I_EX_0003, newTodo.ID)
	}
	return nil
}
