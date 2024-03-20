/*
transaction パッケージは、トランザクション管理に関する機能を提供するパッケージです。
*/
package transaction

import (
	"context"

	"example.com/appbase/pkg/apcontext"
	"example.com/appbase/pkg/constant"
	"example.com/appbase/pkg/domain"
	"example.com/appbase/pkg/logging"
	"example.com/appbase/pkg/message"
	"example.com/appbase/pkg/transaction/entity"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/cockroachdb/errors"
)

const (
	TRANSACTION_CTX_KEY = "TRANSACTION"
)

// TransactionManager はトランザクションを管理するインタフェースです
type TransactionManager interface {
	// ExecuteTransaction は、Serviceの関数serviceFuncの実行前後でDynamoDBトランザクション実行します。
	ExecuteTransaction(serviceFunc domain.ServiceFunc) (any, error)
	// ExecuteTransactionWithContext は、goroutine向けに、渡されたContextを利用して、
	// Serviceの関数serviceFuncの実行前後でDynamoDBトランザクション実行します。
	// goroutineで実施する場合は、この関数を利用してください。また、ServiceFuncWithContextで渡されるContextを引き継いで
	// TransactionalDynamoDBAccessor.AppendTransactWriteItemWithContext、
	// TransactionalSQSAccessor.AppendTransactMessageWithContextの引数に渡して利用してください。
	// そうしないと、トランザクションデータが正しく伝番されません。
	ExecuteTransactionWithContext(context context.Context, serviceFunc domain.ServiceFuncWithContext) (any, error)
}

// NewTransactionManager は、TransactionManagerを作成します
func NewTransactionManager(log logging.Logger,
	dynamodbAccessor TransactionalDynamoDBAccessor,
	sqsAccessor TransactionalSQSAccessor,
	messageRegsiterer MessageRegisterer) TransactionManager {
	return &defaultTransactionManager{log: log,
		dynamodbAccessor:  dynamodbAccessor,
		sqsAccessor:       sqsAccessor,
		messageRegsiterer: messageRegsiterer,
	}
}

// NewTransactionManagerFoDBOnly は、DynamoDBのみのトランザクションに対応するTransactionManagerを作成します。
// SQSのトランザクションは利用しない場合に使用します。
func NewTransactionManagerForDBOnly(log logging.Logger,
	dynamodbAccessor TransactionalDynamoDBAccessor,
	messageRegsterer MessageRegisterer,
) TransactionManager {
	return &defaultTransactionManager{log: log,
		dynamodbAccessor:  dynamodbAccessor,
		messageRegsiterer: messageRegsterer,
	}
}

// defaultTransactionManager は、TransactionManagerを実装する構造体です。
type defaultTransactionManager struct {
	log               logging.Logger
	dynamodbAccessor  TransactionalDynamoDBAccessor
	sqsAccessor       TransactionalSQSAccessor
	messageRegsiterer MessageRegisterer
}

// ExecuteTransaction implements TransactionManager.
func (tm *defaultTransactionManager) ExecuteTransaction(serviceFunc domain.ServiceFunc) (any, error) {
	return tm.ExecuteTransactionWithContext(apcontext.Context, func(ctx context.Context) (any, error) {
		// トランザクション付きのContextを設定
		apcontext.Context = ctx
		return serviceFunc()
	})
}

// ExecuteTransactionWithContext implements TransactionManager.
func (tm *defaultTransactionManager) ExecuteTransactionWithContext(ctx context.Context,
	serviceFunc domain.ServiceFuncWithContext) (result any, err error) {
	if ctx == nil {
		ctx = apcontext.Context
	}
	// 新しいトランザクションを作成
	transaction := newTrasaction(tm.log, tm.messageRegsiterer)
	// トランザクション付きのContextを作成
	ctxWithTx := context.WithValue(ctx, TRANSACTION_CTX_KEY, transaction)

	// トランザクションを開始
	transaction.Start(tm.dynamodbAccessor, tm.sqsAccessor)

	defer func() {
		if r := recover(); r != nil {
			// panic発生時トランザクションをロールバック
			transaction.Rollback()
			// 上位にpanicをリスロー
			panic(r)
		} else if err != nil {
			// Serviceの実行エラー時トランザクションをロールバック
			transaction.Rollback()
		} else {
			// Serviceの実行成功時トランザクションをコミット
			_, err = transaction.Commit()
			// TODO: TransactWriteItemsOutputの利用（ログ出力等）
		}
	}()

	// サービスの実行
	result, err = serviceFunc(ctxWithTx)

	return
}

// Transactionは トランザクションを表すインタフェースです
type Transaction interface {
	// Start は、トランザクションを開始します。
	Start(dynamodbAccessor TransactionalDynamoDBAccessor, sqsAccessor TransactionalSQSAccessor)
	// AppendTransactWriteItemは、DBへトランザクション書き込みしたい場合に対象のTransactWriteItemを追加します。
	AppendTransactWriteItem(item *types.TransactWriteItem)
	// AppendTransactMessageは、SQSへトランザクション管理してメッセージ送信したい場合に対象のMessageを追加します。
	AppendTransactMessage(message *Message)
	// CheckTransactWriteItems は、TransactWriteItemが存在するかを確認します。
	CheckTransactWriteItems() bool
	// Commit は、トランザクションをコミットします。
	Commit() (*dynamodb.TransactWriteItemsOutput, error)
	// Rollback は、トランザクションをロールバックします。
	Rollback()
}

// newTrasactionは 新しいTransactionを作成します。
func newTrasaction(log logging.Logger, messageRegsiterer MessageRegisterer) Transaction {
	return &defaultTransaction{log: log, messageRegsiterer: messageRegsiterer}
}

// defaultTransactionは、transactionを実装する構造体です。
type defaultTransaction struct {
	log               logging.Logger
	messageRegsiterer MessageRegisterer
	dynamodbAccessor  TransactionalDynamoDBAccessor
	sqsAccessor       TransactionalSQSAccessor
	// DynamoDBの書き込みトランザクション
	transactWriteItems []types.TransactWriteItem
	// SQSのメッセージ
	messages []*Message

	// TODO: 読み込みトランザクションTransactGetItems
	// transactGetItems []types.TransactGetItem
}

// Start implements Transaction.
func (t *defaultTransaction) Start(dynamodbAccessor TransactionalDynamoDBAccessor, sqsAccessor TransactionalSQSAccessor) {
	t.log.Debug("トランザクション開始")
	t.dynamodbAccessor = dynamodbAccessor
	t.sqsAccessor = sqsAccessor
}

// AppendTransactWriteItem implements Transaction.
func (t *defaultTransaction) AppendTransactWriteItem(item *types.TransactWriteItem) {
	t.transactWriteItems = append(t.transactWriteItems, *item)
}

// AppendTransactMessage implements transaction.
func (t *defaultTransaction) AppendTransactMessage(message *Message) {
	t.messages = append(t.messages, message)
}

// CheckTransactWriteItems implements Transaction.
func (t *defaultTransaction) CheckTransactWriteItems() bool {
	return len(t.transactWriteItems) > 0
}

// Commit implements Transaction.
func (t *defaultTransaction) Commit() (*dynamodb.TransactWriteItemsOutput, error) {
	var err error
	if t.sqsAccessor != nil {
		// SQSのメッセージの送信とメッセージのDBトランザクション管理
		err = t.sqsAccessor.TransactSendMessages(t.messages)
		if err != nil {
			t.log.Debug("SQSのメッセージ送信失敗でロールバック")
			return nil, errors.WithStack(err)
		}
	}
	// ディレード処理の場合は、メッセージ管理テーブルのアイテムの重複メッセージIDを登録する更新トランザクションを追加
	err = t.transactUpdateQueueMessageItem()
	if err != nil {
		return nil, err
	}

	// DBトランザクションの実行
	if !t.CheckTransactWriteItems() {
		t.log.Debug("トランザクション処理なし")
		return nil, err
	}

	// DynamoDBトランザクション実行
	output, err := t.dynamodbAccessor.TransactWriteItemsSDK(t.transactWriteItems)
	if err != nil {
		t.log.Debug("トランザクションコミットエラー")
		// https://docs.aws.amazon.com/ja_jp/amazondynamodb/latest/developerguide/transaction-apis.html
		var txCanceledException *types.TransactionCanceledException
		var txConflictException *types.TransactionConflictException
		// トランザクションコミット失敗の理由をログ出力
		if errors.As(err, &txCanceledException) {
			for _, v := range txCanceledException.CancellationReasons {
				codePtr := v.Code
				messagePtr := v.Message
				var code string
				if codePtr == nil {
					code = ""
				} else {
					code = *codePtr
				}
				var msg string
				if messagePtr == nil {
					msg = ""
				} else {
					msg = *messagePtr
				}
				t.log.Info(message.I_FW_0003, code, msg, v.Item)
			}
		} else if errors.As(err, &txConflictException) {
			t.log.Info(message.I_FW_0004, *txConflictException.ErrorCodeOverride, *txConflictException.Message)
		}
		return nil, errors.WithStack(err)
	}
	t.log.Debug("トランザクションコミット")
	return output, nil
}

// Rollback implements Transaction.
func (t *defaultTransaction) Rollback() {
	if t.CheckTransactWriteItems() {
		t.log.Debug("業務処理エラーでトランザクションロールバック")
	} else {
		t.log.Debug("業務処理エラーだがトランザクション処理なし")
	}
}

// transactUpdateQueueMessageItem は、メッセージ管理テーブルのアイテムの重複メッセージIDを登録する更新トランザクションを追加します。
func (t *defaultTransaction) transactUpdateQueueMessageItem() error {
	// Contextから非同期処理情報を取得
	asyncHandlerInfo := apcontext.Context.Value(constant.ASYNC_HANDLER_INFO_CTX_KEY)
	if asyncHandlerInfo == nil {
		t.log.Debug("非同期処理情報なし")
		return nil
	}
	queueMessageItem, ok := asyncHandlerInfo.(*entity.QueueMessageItem)
	if ok {
		// メッセージ管理テーブルのアイテムのステータスを完了に更新するトランザクションを追加
		t.log.Debug("メッセージ管理テーブルに重複メッセージIDを登録する更新トランザクションを追加")
		queueMessageItem.Status = constant.QUEUE_MESSAGE_STATUS_COMPLETE
		return t.messageRegsiterer.UpdateMessage(queueMessageItem)
	}
	//TODO: エラー定義
	return errors.Errorf("非同期処理情報の型が誤りのため、処理できません。")
}
