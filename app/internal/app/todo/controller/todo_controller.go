// controllerのパッケージ
package controller

import (
	"app/internal/app/todo/service"
	"app/internal/pkg/message"
	"errors"

	"example.com/appbase/pkg/domain"
	myerrors "example.com/appbase/pkg/errors"
	"example.com/appbase/pkg/logging"
	"example.com/appbase/pkg/transaction"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gin-gonic/gin"
)

// Request は、REST APIで受け取るリクエストデータの構造体です。
type Request struct {
	// TodoTitle は、Todoのタイトルです。
	TodoTitle string `json:"todo_title" binding:"required"`
}

// TodoController は、Todo業務のControllerインタフェースです。
type TodoController interface {
	// Find は、パスパラメータで指定されたtodo_idのTodoを照会します。
	Find(ctx *gin.Context) (any, error)
	// Register は、リクエストデータで受け取ったTodoを登録します。
	Register(ctx *gin.Context) (any, error)
}

// New は、TodoControllerを作成します。
func New(log logging.Logger,
	transactionManager transaction.TransactionManager,
	service service.TodoService,
) TodoController {
	return &todoControllerImpl{log: log,
		transactionManager: transactionManager,
		service:            service,
	}
}

// todoControllerImpl は、TodoControllerを実装する構造体です。
type todoControllerImpl struct {
	log                logging.Logger
	transactionManager transaction.TransactionManager
	service            service.TodoService
}

func (c *todoControllerImpl) Find(ctx *gin.Context) (any, error) {
	// パスパラメータの取得
	todoId := ctx.Param("todo_id")
	// 入力チェック
	if todoId == "" {
		// 入力チェックエラーのハンドリング
		return nil, myerrors.NewValidationErrorWithMessage("クエリパラメータtodoIdが未指定です")
	}
	// DynamoDBトランザクション管理してサービスの実行
	return c.transactionManager.ExecuteTransaction(func() (any, error) {
		return c.service.Find(todoId)
	})
}

func (c *todoControllerImpl) Register(ctx *gin.Context) (any, error) {
	// POSTデータをバインド
	var request Request
	if err := ctx.ShouldBindJSON(&request); err != nil {
		// 入力チェックエラーのハンドリング
		return nil, myerrors.NewValidationError(err)
	}
	// クエリパラメータの取得
	transaction := ctx.Query("tx")
	var serviceFunc domain.ServiceFunc

	if transaction != "" {
		serviceFunc = func() (any, error) {
			// トランザクション指定あり
			return c.service.RegisterTx(request.TodoTitle)
		}
	} else {
		// トランザクション指定なし
		serviceFunc = func() (any, error) {
			return c.service.Register(request.TodoTitle)
		}
	}

	// DynamoDBトランザクション管理してサービスの実行
	result, err := c.transactionManager.ExecuteTransaction(serviceFunc)
	if err != nil {
		// TODO: ロールバックの場合に、予期せぬエラーとならないよう各Controllerでハンドリングするか？
		// 集約的にinterceptorで実施するか？
		var txCanceledException *types.TransactionCanceledException
		var txConflictException *types.TransactionConflictException
		// 登録失敗の業務エラー
		if errors.As(err, &txCanceledException) {
			return nil, myerrors.NewBusinessError(message.W_EX_8003, request.TodoTitle)
		} else if errors.As(err, &txConflictException) {
			return nil, myerrors.NewBusinessError(message.W_EX_8004, request.TodoTitle)
		}
		return nil, err
	}
	return result, nil
}
