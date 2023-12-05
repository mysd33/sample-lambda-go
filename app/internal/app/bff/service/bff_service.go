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
	// FindTodo は、指定したuserIdとtodoIdに合致するユーザ情報とTodoを照会します。
	FindTodo(userId string, todoId string) (*entity.User, *entity.Todo, error)
	// RegisterUser は、リクエストデータで受け取ったユーザ情報を登録します。
	RegisterUser(userName string) (*entity.User, error)
	// RegisterTodo は、タイトルtodoTitleのTodoを登録します。
	RegisterTodo(todoTitle string) (*entity.Todo, error)
}

// New は、BffServiceを作成します。
func New(log logging.Logger,
	config config.Config,
	userRepository repository.UserRepository,
	todoRepository repository.TodoRepository,
) BffService {
	return &bffServiceImpl{log: log, config: config, userRepository: userRepository, todoRepository: todoRepository}
}

// todoServiceImpl BffServiceを実装する構造体です。
type bffServiceImpl struct {
	log            logging.Logger
	config         config.Config
	userRepository repository.UserRepository
	todoRepository repository.TodoRepository
}

// RegisterUser implements BffService.
func (bs *bffServiceImpl) RegisterUser(userName string) (*entity.User, error) {
	user := entity.User{Name: userName}
	return bs.userRepository.CreateOne(&user)
}

// RegisterTodo implements BffService.
func (bs *bffServiceImpl) RegisterTodo(todoTitle string) (*entity.Todo, error) {
	todo := entity.Todo{Title: todoTitle}
	return bs.todoRepository.CreateOne(&todo)
}

// FindTodo implements BffService.
func (bs *bffServiceImpl) FindTodo(userId string, todoId string) (*entity.User, *entity.Todo, error) {
	bs.log.Debug("userId:%s,todoId:%s", userId, todoId)

	user, err := bs.userRepository.FindOne(userId)
	if err != nil {
		return nil, nil, err
	}
	bs.log.Debug("user:%+v", user)
	todo, err := bs.todoRepository.FindOne(todoId)
	if err != nil {
		return nil, nil, err
	}
	bs.log.Debug("todo:%+v", todo)
	return user, todo, nil
}
