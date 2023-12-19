// controllerのパッケージ
package controller

import (
	"app/internal/app/todo-async/service"

	"example.com/appbase/pkg/logging"
	"example.com/appbase/pkg/transaction"
	"github.com/aws/aws-lambda-go/events"
)

// TodoAsyncController は、Todo業務のControllerインタフェースです。
type TodoAsyncController interface {
	// RegisterAll は、SQSメッセージとして受け取ったTodoのリストを全て登録します。
	RegisterAll(sqsMessage events.SQSMessage) error
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

// RegisterAll implements TodoAsyncController.
func (c *todoAsyncControllerImpl) RegisterAll(sqsMessage events.SQSMessage) error {
	body := sqsMessage.Body
	c.log.Debug("Message: %s", body)

	//TODO: メッセージをjsonデコードして、todoTitileを取得する処理
	//現状、ダミーの処理
	todoTitles := []string{body}

	// DynamoDBトランザクション管理してサービスの実行
	_, err := c.transactionManager.ExecuteTransaction(func() (any, error) {
		err := c.service.RegisterTodosAsync(todoTitles)
		return nil, err
	})
	return err
}
