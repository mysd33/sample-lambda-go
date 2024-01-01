// serviceのパッケージ
package service

import (
	"app/internal/pkg/entity"
	"app/internal/pkg/repository"

	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/id"
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
	// RegisterTodosAsync は、（標準キューで）タイトルのリストtodoTitlesのTodoを非同期で登録します。
	RegisterTodosAsync(todoTitles []string, dbtx string) error
	// RegisterTodosAsyncByFIFO は、FIFOキューでタイトルのリストtodoTitlesのTodoを非同期で登録します。
	RegisterTodosAsyncByFIFO(todoTitles []string, dbtx string) error
}

// New は、BffServiceを作成します。
func New(log logging.Logger,
	config config.Config,
	userRepository repository.UserRepository,
	todoRepository repository.TodoRepository,
	dummyRepository repository.DummyRepository,
	asyncMessageRepository repository.AsyncMessageRepository,
) BffService {
	return &bffServiceImpl{
		log:                    log,
		config:                 config,
		userRepository:         userRepository,
		todoRepository:         todoRepository,
		dummyRepository:        dummyRepository,
		asyncMessageRepository: asyncMessageRepository,
	}
}

// todoServiceImpl BffServiceを実装する構造体です。
type bffServiceImpl struct {
	log                    logging.Logger
	config                 config.Config
	userRepository         repository.UserRepository
	todoRepository         repository.TodoRepository
	dummyRepository        repository.DummyRepository
	asyncMessageRepository repository.AsyncMessageRepository
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

// RegisterTodosAsync implements TodoService.
func (bs *bffServiceImpl) RegisterTodosAsync(todoTitles []string, dbtx string) error {
	bs.log.Debug("RegisterTodosAsync")
	// DBトランザクションを試すためのダミーのDB登録処理
	if dbtx != "no" {
		bs.log.Debug("業務のDB登録処理あり")
		bs.dummyRepository.CreateOneTx(&entity.Dummy{Value: "dummy"})
	}
	// TODOタイトルのリストの登録を非同期処理実行依頼
	asyncMessage := &entity.AsyncMessage{TodoTitles: todoTitles}
	bs.asyncMessageRepository.Send(asyncMessage)
	return nil
}

// RegisterTodosAsyncByFIFO implements BffService.
func (bs *bffServiceImpl) RegisterTodosAsyncByFIFO(todoTitles []string, dbtx string) error {
	bs.log.Debug("RegisterTodosAsyncByFIFO")
	if dbtx != "no" {
		bs.log.Debug("業務のDB登録処理あり")
		// DBトランザクションを試すためのダミーのDB登録処理
		bs.dummyRepository.CreateOneTx(&entity.Dummy{Value: "dummy2"})
	}

	// TODOタイトルのリストの登録を非同期処理実行依頼
	asyncMessage := &entity.AsyncMessage{TodoTitles: todoTitles}
	// メッセージグループID
	msgGroupId := id.GenerateId()
	bs.asyncMessageRepository.SendToFIFOQueue(asyncMessage, msgGroupId)
	return nil

}
