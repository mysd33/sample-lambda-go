// controllerのパッケージ
package controller

import (
	"app/internal/app/todo-async/service"
	"app/internal/pkg/entity"
	"app/internal/pkg/message"
	"encoding/json"

	"example.com/appbase/pkg/errors"
	myerrors "example.com/appbase/pkg/errors"
	"example.com/appbase/pkg/logging"
	"example.com/appbase/pkg/transaction"
	"github.com/aws/aws-lambda-go/events"
)

// TodoAsyncController は、Todo業務のControllerインタフェースです。
type TodoAsyncController interface {
	// RegisterAllAsync は、SQSメッセージとして受け取ったTodoのリストを全て登録します。
	RegisterAllAsync(sqsMessage events.SQSMessage) error
}

func New(log logging.Logger,
	transactionManager transaction.TransactionManager,
	service service.TodoAsyncService) TodoAsyncController {
	return &todoAsyncControllerImpl{
		log:                log,
		transactionManager: transactionManager,
		service:            service,
	}
}

type todoAsyncControllerImpl struct {
	log                logging.Logger
	transactionManager transaction.TransactionManager
	service            service.TodoAsyncService
}

// RegisterAllAsync implements TodoAsyncController.
func (c *todoAsyncControllerImpl) RegisterAllAsync(sqsMessage events.SQSMessage) error {
	body := sqsMessage.Body
	c.log.Debug("Message: %s", body)

	//メッセージをjsonデコードして、AsyncMessageを取得する処理
	var asyncMessage entity.AsyncMessage
	err := json.Unmarshal([]byte(body), &asyncMessage)
	if err != nil {
		return errors.NewSystemError(err, message.E_EX_9003)
	}

	// DynamoDBトランザクション管理してサービスの実行
	_, err = c.transactionManager.ExecuteTransaction(func() (any, error) {
		err := c.service.RegisterTodosAsync(asyncMessage)
		return nil, err
	})
	if transaction.IsTransactionConditionalCheckFailed(err) {
		// 登録失敗の業務エラーにするか、スキップするかはケースバイケース
		return myerrors.NewBusinessErrorWithCause(err, message.W_EX_8008)
	} else if transaction.IsTransactionConflict(err) {
		// 登録失敗の業務エラーにするか、スキップするかはケースバイケース
		return myerrors.NewBusinessErrorWithCause(err, message.W_EX_8008)
	}
	return err
}
