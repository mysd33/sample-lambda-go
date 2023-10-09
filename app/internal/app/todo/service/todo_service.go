package service

import (
	"app/internal/pkg/entity"
	"app/internal/pkg/repository"

	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/logging"
)

type TodoService interface {
	Regist(todoTitle string) (*entity.Todo, error)
	Find(todoId string) (*entity.Todo, error)
}

func New(log logging.Logger,
	config *config.Config,
	repository *repository.TodoRepository,
) TodoService {
	return &TodoServiceImpl{log: log, config: config, repository: repository}
}

type TodoServiceImpl struct {
	log        logging.Logger
	config     *config.Config
	repository *repository.TodoRepository
}

func (ts *TodoServiceImpl) Regist(todoTitle string) (*entity.Todo, error) {
	//Zapによるログ出力の例
	ts.log.Info("TodoTitle=%s", todoTitle)

	todo := entity.Todo{}
	todo.Title = todoTitle
	return (*ts.repository).PutTodo(&todo)
}

func (ts *TodoServiceImpl) Find(todoId string) (*entity.Todo, error) {
	return (*ts.repository).GetTodo(todoId)
}
