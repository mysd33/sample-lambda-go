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
	"example.com/appbase/pkg/logging"
	"example.com/appbase/pkg/objectstorage"
)

const (
	S3_BUCKET_NAME = "S3_BUCKET_NAME"
)

// TodoAsyncService は、Todoの非同期処理を管理するServiceインタフェースです。
type TodoAsyncService interface {
	// RegisterTodosAsync は、Todoを非同期で登録します。
	RegisterTodosAsync(asyncMesssage model.AsyncMessage) error
}

// NewTodoAsyncService は、TodoAsyncServiceを生成します。
func New(logger logging.Logger,
	config config.Config,
	obectStorageAccessor objectstorage.ObjectStorageAccessor,
	tempRepository repository.TempRepository,
	todoRepository repository.TodoRepository) TodoAsyncService {
	return &todoAsyncServiceImpl{logger: logger,
		config:                config,
		objectstorageAccessor: obectStorageAccessor,
		tempRepository:        tempRepository,
		todoRepository:        todoRepository,
	}
}

type todoAsyncServiceImpl struct {
	logger                logging.Logger
	config                config.Config
	objectstorageAccessor objectstorage.ObjectStorageAccessor
	tempRepository        repository.TempRepository
	todoRepository        repository.TodoRepository
}

// RegisterTodosAsync implements TodoAsyncService.
func (ts *todoAsyncServiceImpl) RegisterTodosAsync(asyncMesssage model.AsyncMessage) error {
	if asyncMesssage.TempId == "" {
		// TempIdが空の場合は、何もしない
		return nil
	}
	// tempテーブルのIDをもとに、TodoListを取得
	temp, err := ts.tempRepository.FindOne(asyncMesssage.TempId)
	if err != nil {
		return err
	}
	ts.logger.Debug("temp: %+v", temp)
	bucketName, found := ts.config.GetWithContains(S3_BUCKET_NAME)
	if !found {
		// TODO: エラー処理
		return errors.NewSystemError(fmt.Errorf("バケットの設定[%s]が見つかりません", S3_BUCKET_NAME), message.E_EX_9001)
	}
	// S3からファイルを取得
	filePath := temp.Value
	data, err := ts.objectstorageAccessor.Download(bucketName, filePath)
	if err != nil {
		// TODO: エラー処理
		return errors.NewSystemError(err, message.E_EX_9001)
	}
	// jsonファイルをアンマーシャルして、Todoのリストを取得
	var todoTitles []string
	err = json.Unmarshal(data, &todoTitles)
	if err != nil {
		// TODO: エラー処理
		return errors.NewSystemError(err, message.E_EX_9001)
	}
	// TODO: todoTitlesをS3上のファイルから取得して登録するように変更
	for _, v := range todoTitles {
		ts.logger.Debug("todoTitle: %s", v)
		todo := model.Todo{Title: v}
		newTodo, err := ts.todoRepository.CreateOneTx(&todo)
		if err != nil {
			return err
		}
		ts.logger.Info(message.I_EX_0003, newTodo.ID)
	}
	// TODO: tempテーブルのアイテムの削除
	return nil
}
