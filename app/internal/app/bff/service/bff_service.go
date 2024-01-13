// serviceのパッケージ
package service

import (
	"app/internal/pkg/entity"
	"app/internal/pkg/message"
	"app/internal/pkg/repository"
	"encoding/json"

	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/errors"
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
	id id.IDGenerator,
	userRepository repository.UserRepository,
	todoRepository repository.TodoRepository,
	tempRepository repository.TempRepository,
	asyncMessageRepository repository.AsyncMessageRepository,
) BffService {
	return &bffServiceImpl{
		log:                    log,
		config:                 config,
		id:                     id,
		userRepository:         userRepository,
		todoRepository:         todoRepository,
		tempRepository:         tempRepository,
		asyncMessageRepository: asyncMessageRepository,
	}
}

// todoServiceImpl BffServiceを実装する構造体です。
type bffServiceImpl struct {
	log                    logging.Logger
	config                 config.Config
	id                     id.IDGenerator
	userRepository         repository.UserRepository
	todoRepository         repository.TodoRepository
	tempRepository         repository.TempRepository
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
	var tempId string
	// TODO: todoTitlesをS3のファイルに入れて登録するように変更
	if dbtx != "no" {
		bs.log.Debug("業務のDB登録処理あり")
		// TODO: Valueに、S3のパスを入れて登録するように変更
		byteMessage, err := json.Marshal(todoTitles)
		if err != nil {
			return errors.NewSystemError(err, message.E_EX_9001)
		}
		value := string(byteMessage)
		temp := &entity.Temp{Value: value}
		// Tempテーブルの登録
		bs.tempRepository.CreateOneTx(temp)
		tempId = temp.ID
	}
	// TODOリストの登録を非同期処理実行依頼
	asyncMessage := &entity.AsyncMessage{TempId: tempId}
	bs.asyncMessageRepository.Send(asyncMessage)
	return nil
}

// RegisterTodosAsyncByFIFO implements BffService.
func (bs *bffServiceImpl) RegisterTodosAsyncByFIFO(todoTitles []string, dbtx string) error {
	bs.log.Debug("RegisterTodosAsyncByFIFO")
	var tempId string
	// TODO: todoTitlesをS3上のファイルに入れて登録するように変更
	if dbtx != "no" {
		bs.log.Debug("業務のDB登録処理あり")
		// TODO: Valueに、S3のパスを入れて登録するように変更
		byteMessage, err := json.Marshal(todoTitles)
		if err != nil {
			return errors.NewSystemError(err, message.E_EX_9001)
		}
		value := string(byteMessage)
		temp := &entity.Temp{Value: value}
		// Tempテーブルの登録
		bs.tempRepository.CreateOneTx(temp)
		tempId = temp.ID
	}

	// TODOリストの登録を非同期処理実行依頼
	asyncMessage := &entity.AsyncMessage{TempId: tempId}
	// メッセージグループIDの生成
	msgGroupId := bs.id.GenerateUUID()
	bs.asyncMessageRepository.SendToFIFOQueue(asyncMessage, msgGroupId)
	return nil

}
