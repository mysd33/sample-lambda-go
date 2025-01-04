// serviceのパッケージ
package service

import (
	"app/internal/pkg/message"
	"app/internal/pkg/model"
	"app/internal/pkg/repository"
	"encoding/json"
	"fmt"

	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/errors"
	"example.com/appbase/pkg/id"
	"example.com/appbase/pkg/logging"
	"example.com/appbase/pkg/objectstorage"
)

const (
	S3_BUCKET_NAME = "S3_BUCKET_NAME"
	tempFilePath   = "todoFiles/%s.json"
)

// TodoService は、Bff業務のServiceインタフェースです。
type BffService interface {
	// FindTodo は、指定したuserIdとtodoIdに合致するユーザ情報とTodoを照会します。
	FindTodo(userId string, todoId string) (*model.User, *model.Todo, error)
	// RegisterUser は、リクエストデータで受け取ったユーザ情報を登録します。
	RegisterUser(userName string) (*model.User, error)
	// RegisterTodo は、タイトルtodoTitleのTodoを登録します。
	RegisterTodo(todoTitle string) (*model.Todo, error)
	// RegisterTodosAsync は、（標準キューで）タイトルのリストtodoTitlesのTodoを非同期で登録します。
	RegisterTodosAsync(todoTitles []string, dbtx string) error
	// RegisterTodosAsyncByFIFO は、FIFOキューでタイトルのリストtodoTitlesのTodoを非同期で登録します。
	RegisterTodosAsyncByFIFO(todoTitles []string, dbtx string) error
	// FindBooksByCriteria は、条件に合致する書籍を取得します。
	FindBooksByCriteria(criteria *repository.BookCriteria) ([]model.Book, error)
	// RegisterBook は、書籍を登録します。
	RegisterBook(book *model.Book) (*model.Book, error)
}

// New は、BffServiceを作成します。
func New(logger logging.Logger,
	config config.Config,
	id id.IDGenerator,
	obectStorageAccessor objectstorage.ObjectStorageAccessor,
	userRepository repository.UserRepository,
	todoRepository repository.TodoRepository,
	tempRepository repository.TempRepository,
	asyncMessageRepository repository.AsyncMessageRepository,
	bookRepository repository.BookRepository,
) BffService {
	return &bffServiceImpl{
		logger:                 logger,
		config:                 config,
		id:                     id,
		obectStorageAccessor:   obectStorageAccessor,
		userRepository:         userRepository,
		todoRepository:         todoRepository,
		tempRepository:         tempRepository,
		asyncMessageRepository: asyncMessageRepository,
		bookRepository:         bookRepository,
	}
}

// todoServiceImpl BffServiceを実装する構造体です。
type bffServiceImpl struct {
	logger                 logging.Logger
	config                 config.Config
	id                     id.IDGenerator
	obectStorageAccessor   objectstorage.ObjectStorageAccessor
	userRepository         repository.UserRepository
	todoRepository         repository.TodoRepository
	tempRepository         repository.TempRepository
	asyncMessageRepository repository.AsyncMessageRepository
	bookRepository         repository.BookRepository
}

// RegisterUser implements BffService.
func (bs *bffServiceImpl) RegisterUser(userName string) (*model.User, error) {
	user := model.User{Name: userName}
	return bs.userRepository.CreateOne(&user)
}

// RegisterTodo implements BffService.
func (bs *bffServiceImpl) RegisterTodo(todoTitle string) (*model.Todo, error) {
	todo := model.Todo{Title: todoTitle}
	return bs.todoRepository.CreateOne(&todo)
}

// FindTodo implements BffService.
func (bs *bffServiceImpl) FindTodo(userId string, todoId string) (*model.User, *model.Todo, error) {
	bs.logger.Debug("userId:%s,todoId:%s", userId, todoId)

	user, err := bs.userRepository.FindOne(userId)
	if err != nil {
		return nil, nil, err
	}
	bs.logger.Debug("user:%+v", user)
	todo, err := bs.todoRepository.FindOne(todoId)
	if err != nil {
		return nil, nil, err
	}
	bs.logger.Debug("todo:%+v", todo)
	return user, todo, nil
}

// RegisterTodosAsync implements TodoService.
func (bs *bffServiceImpl) RegisterTodosAsync(todoTitles []string, dbtx string) error {
	bs.logger.Debug("RegisterTodosAsync")
	var tempId string
	if dbtx != "no" {
		bs.logger.Debug("業務のDB登録処理あり")
		temp, err := bs.registerTemp(todoTitles)
		if err != nil {
			return err
		}
		tempId = temp.ID
	}
	// TODOリストの登録を非同期処理実行依頼
	asyncMessage := &model.AsyncMessage{TempId: tempId}
	bs.asyncMessageRepository.Send(asyncMessage)
	return nil
}

// RegisterTodosAsyncByFIFO implements BffService.
func (bs *bffServiceImpl) RegisterTodosAsyncByFIFO(todoTitles []string, dbtx string) error {
	bs.logger.Debug("RegisterTodosAsyncByFIFO")
	var tempId string
	if dbtx != "no" {
		bs.logger.Debug("業務のDB登録処理あり")
		temp, err := bs.registerTemp(todoTitles)
		if err != nil {
			return err
		}
		tempId = temp.ID
	}

	// TODOリストの登録を非同期処理実行依頼
	asyncMessage := &model.AsyncMessage{TempId: tempId}
	// メッセージグループIDの生成
	msgGroupId, err := bs.id.GenerateUUID()
	if err != nil {
		return errors.NewSystemError(err, message.E_EX_9001)
	}
	bs.asyncMessageRepository.SendToFIFOQueue(asyncMessage, msgGroupId)
	return nil

}

func (bs *bffServiceImpl) registerTemp(todoTitles []string) (*model.Temp, error) {
	// todoTitlesの内容をS3にファイルとして格納する
	byteMessage, err := json.Marshal(todoTitles)
	if err != nil {
		// TODO: エラー処理
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}
	bucketName, found := bs.config.GetWithContains(S3_BUCKET_NAME)
	if !found {
		// TODO: エラー処理
		return nil, errors.NewSystemError(fmt.Errorf("バケットの設定[%s]が見つかりません", S3_BUCKET_NAME), message.E_EX_9001)
	}
	// ファイル名の生成
	fileName, err := bs.id.GenerateUUID()
	if err != nil {
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}
	objectKey := fmt.Sprintf(tempFilePath, fileName)
	_, err = bs.obectStorageAccessor.Upload(bucketName, objectKey, byteMessage)
	if err != nil {
		// TODO: エラー処理
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}
	// Valueに、S3のパスを入れて登録するように変更
	temp := &model.Temp{Value: objectKey}
	// Tempテーブルの登録
	bs.tempRepository.CreateOneTx(temp)
	return temp, nil
}

// FindBooksByCriteria implements BffService.
func (bs *bffServiceImpl) FindBooksByCriteria(criteria *repository.BookCriteria) ([]model.Book, error) {
	return bs.bookRepository.FindSomeByCriteria(criteria)
}

// RegisterBook implements BffService.
func (bs *bffServiceImpl) RegisterBook(book *model.Book) (*model.Book, error) {
	return bs.bookRepository.CreateOne(book)
}
