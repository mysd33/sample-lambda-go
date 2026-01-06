// controllerのパッケージ
package controller

import (
	"app/internal/app/todo/service"
	"app/internal/pkg/message"

	"example.com/appbase/pkg/domain"
	"example.com/appbase/pkg/errors"
	"example.com/appbase/pkg/logging"
	"example.com/appbase/pkg/transaction"
	"github.com/gin-gonic/gin"
)

// RequestFind は、Todo照会のREST APIで受け取るリクエストデータの構造体です。
type RequestFind struct {
	// TodoId は、TodoのIDです。
	TodoId string `label:"Todo ID(todo_id)" uri:"todo_id" binding:"required,uuid"`
}

// RequestRegister は、Todo登録REST APIで受け取るリクエストデータの構造体です。
type RequestRegister struct {
	// TodoTitle は、Todoのタイトルです。
	TodoTitle string `label:"Todoタイトル(todo_title)" json:"todo_title" binding:"required"`
}

// TodoController は、Todo業務のControllerインタフェースです。
type TodoController interface {
	// Find は、パスパラメータで指定されたtodo_idのTodoを照会します。
	Find(ctx *gin.Context) (any, error)
	// Register は、リクエストデータで受け取ったTodoを登録します。
	Register(ctx *gin.Context) (any, error)
}

// New は、TodoControllerを作成します。
func New(logger logging.Logger,
	transactionManager transaction.TransactionManager,
	service service.TodoService,
) TodoController {
	return &todoControllerImpl{logger: logger,
		transactionManager: transactionManager,
		service:            service,
	}
}

// todoControllerImpl は、TodoControllerを実装する構造体です。
type todoControllerImpl struct {
	logger             logging.Logger
	transactionManager transaction.TransactionManager
	service            service.TodoService
}

func (c *todoControllerImpl) Find(ctx *gin.Context) (any, error) {
	// パスパラメータの取得
	var request RequestFind
	if err := ctx.ShouldBindUri(&request); err != nil {
		// 入力チェックエラーのハンドリング
		return nil, errors.NewValidationErrorWithCause(err, message.W_EX_5001)
	}

	// DynamoDBトランザクション管理してサービスの実行
	return c.transactionManager.ExecuteTransaction(func() (any, error) {
		return c.service.Find(request.TodoId)
	})
}

func (c *todoControllerImpl) Register(ctx *gin.Context) (any, error) {
	// POSTデータをバインド
	var request RequestRegister
	if err := ctx.ShouldBindJSON(&request); err != nil {
		// 入力チェックエラーのハンドリング
		return nil, errors.NewValidationErrorWithCause(err, message.W_EX_5001)
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
		var bizErrs *errors.BusinessErrors
		// 業務エラーの場合にハンドリングしたい場合は、BusinessErrorsのみAsで判定すればよい
		// BusinessError(単一の業務エラー)の場合もBusinessErrorsとして判定できるようになっている
		if errors.As(err, &bizErrs) {
			// 付加情報が付与できる
			bizErrs.WithInfo("label1")
		} else if transaction.IsTransactionConditionalCheckFailed(err) {
			// 登録失敗の業務エラーにするか、スキップするかはケースバイケース
			return nil, errors.NewBusinessErrorWithCause(err, message.W_EX_8004, request.TodoTitle)
		} else if transaction.IsTransactionConflict(err) {
			// 登録失敗の業務エラーにするか、システムエラーにするかはケースバイケース
			return nil, errors.NewBusinessErrorWithCause(err, message.W_EX_8004, request.TodoTitle)
		}
		/* 混在するケースでも業務エラーにする配慮する場合はこちらを使用
		} else if transaction.IsTransactionConditionalCheckFailedOrTransactionConflict(err) {
			// 登録失敗の業務エラーにするか、システムエラーにするかはケースバイケース
			return nil, errors.NewBusinessErrorWithCause(err, message.W_EX_8004, request.TodoTitle)
		}*/
		return nil, err
	}
	return result, nil
}
