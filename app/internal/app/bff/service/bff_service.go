// serviceのパッケージ
package service

import (
	"app/internal/pkg/entity"
	"app/internal/pkg/repository"

	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/logging"
)

// TodoService は、Bff業務のServiceインタフェースです。
type BffService interface {
	// Find は、指定したuserIdとtodoIdに合致するユーザ情報とTodoを照会します。
	Find(userId string, todoId string) (*entity.User, *entity.Todo, error)
}

// New は、BffServiceを作成します。
func New(log logging.Logger,
	config *config.Config,
	userRepository repository.UserRepository,
	todoRepository repository.TodoRepository,
) BffService {
	return &bffServiceImpl{log: log, config: config, userRepository: userRepository, todoRepository: todoRepository}
}

// todoServiceImpl BffServiceを実装する構造体です。
type bffServiceImpl struct {
	log            logging.Logger
	config         *config.Config
	userRepository repository.UserRepository
	todoRepository repository.TodoRepository
}

// Find implements BffService.
func (bs *bffServiceImpl) Find(userId string, todoId string) (*entity.User, *entity.Todo, error) {
	bs.log.Debug("userId:%s,todoId:%s", userId, todoId)

	user, err := bs.userRepository.GetUser(userId)
	if err != nil {
		return nil, nil, err
	}
	todo, err := bs.todoRepository.GetTodo(todoId)
	if err != nil {
		return nil, nil, err
	}
	return user, todo, nil
}
