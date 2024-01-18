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
		return nil, myerrors.NewValidationError(message.W_EX_5002, "todo_id")
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
		return nil, myerrors.NewValidationErrorWithCause(err, message.W_EX_5001)
	}
	// クエリパラメータの取得
	tx := ctx.Query("tx")
	var serviceFunc domain.ServiceFunc

	if tx != "" {
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
		var bizErrs *myerrors.BusinessErrors
		// トランザクションキャンセル時に、各Controllerでハンドリング
		var txCanceledException *types.TransactionCanceledException
		var txConflictException *types.TransactionConflictException
		// 業務エラーの場合にハンドリングしたい場合は、BusinessErrorsのみAsで判定すればよい
		// BusinessError(単一の業務エラー)の場合もBusinessErrorsとして判定できるようになっている
		if errors.As(err, &bizErrs) {
			// 付加情報が付与できる
			bizErrs.WithInfo("label1")
		} else if errors.As(err, &txCanceledException) &&
			transaction.ContainsConditionalCheckFailed(txCanceledException) {
			// 登録失敗の業務エラーにするか、スキップするかはケースバイケース
			return nil, myerrors.NewBusinessError(message.W_EX_8005)
		} else if errors.As(err, &txConflictException) {
			return nil, myerrors.NewBusinessError(message.W_EX_8005)
		}
		return nil, err
	}
	return result, nil
}
