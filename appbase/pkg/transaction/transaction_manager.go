/*
transaction パッケージは、トランザクション管理に関する機能を提供するパッケージです。
*/
package transaction

import (
	"example.com/appbase/pkg/domain"
	"example.com/appbase/pkg/logging"
	"example.com/appbase/pkg/message"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/cockroachdb/errors"
)

// TransactionManager はトランザクションを管理するインタフェースです
type TransactionManager interface {
	// ExecuteTransaction は、Serviceの関数serviceFuncの実行前後でDynamoDBトランザクション実行します。
	ExecuteTransaction(serviceFunc domain.ServiceFunc) (any, error)
}

// NewTransactionManager は、TransactionManagerを作成します
func NewTransactionManager(log logging.Logger,
	dynamodbAccessor TransactionalDynamoDBAccessor,
	sqsAccessor TransactionalSQSAccessor) TransactionManager {
	return &defaultTransactionManager{log: log,
		dynamodbAccessor: dynamodbAccessor,
		sqsAccessor:      sqsAccessor,
	}
}

type defaultTransactionManager struct {
	log              logging.Logger
	dynamodbAccessor TransactionalDynamoDBAccessor
	sqsAccessor      TransactionalSQSAccessor
}

// ExecuteTransaction implements TransactionManager.
func (tm *defaultTransactionManager) ExecuteTransaction(serviceFunc domain.ServiceFunc) (any, error) {
	// 新しいトランザクションを作成
	transction := newTrasaction(tm.log)
	// トランザクションを開始
	transction.Start(tm.dynamodbAccessor, tm.sqsAccessor)

	// TODO: panicを考慮したdefferによるトランザクション実行コードに修正

	// サービスの実行
	result, err := serviceFunc()
	// DynamoDBのトランザクションを終了
	_, err = transction.End(err)
	if err != nil {
		return nil, err
	}
	return result, nil
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
	// End は、エラーがなければ、AWS SDKによるTransactionWriteItemsを実行しトランザクション実行し、エラーがある場合には実行しません。
	End(err error) (*dynamodb.TransactWriteItemsOutput, error)
}

// newTrasactionは 新しいTransactionを作成します。
func newTrasaction(log logging.Logger) Transaction {
	return &defaultTransaction{log: log}
}

// defaultTransactionは、transactionを実装する構造体です。
type defaultTransaction struct {
	log              logging.Logger
	dynamodbAccessor TransactionalDynamoDBAccessor
	sqsAccessor      TransactionalSQSAccessor
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
	dynamodbAccessor.StartTransaction(t)
	sqsAccessor.StartTransaction(t)
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

// endTransaction implements Transaction.
func (t *defaultTransaction) End(err error) (*dynamodb.TransactWriteItemsOutput, error) {
	if t.sqsAccessor != nil {
		// 業務テーブルのDBトランザクションがあるかチェック
		hasDbTrancation := t.CheckTransactWriteItems()
		// SQSのメッセージの送信とトランザクションを管理
		err := t.sqsAccessor.TransactSendMessages(t.messages, hasDbTrancation)
		if err != nil {
			t.log.Debug("SQSのメッセージ送信失敗でロールバック")
			return nil, errors.WithStack(err)
		}
	}
	// DBトランザクションの実行
	if !t.CheckTransactWriteItems() {
		t.log.Debug("トランザクション処理なし")
		return nil, err
	}
	// 処理結果がどんな場合でもDynamoDBAccessorとSQSAccessorのトランザクションを開放
	defer func() {
		t.dynamodbAccessor.EndTransaction()
		if t.sqsAccessor != nil {
			t.sqsAccessor.EndTransaction()
		}
	}()

	if err != nil {
		t.log.Debug("業務処理エラーでトランザクションロールバック")
		// Serviceの処理結果がエラー場合は、トランザクションを実行せず、元のエラーを返却し終了
		return nil, err
	}
	// DynamoDBトランザクション実行
	output, err := t.dynamodbAccessor.TransactWriteItemsSDK(t.transactWriteItems)
	if err != nil {
		t.log.Debug("トランザクション実行失敗")
		// https://docs.aws.amazon.com/ja_jp/amazondynamodb/latest/developerguide/transaction-apis.html
		var txCanceledException *types.TransactionCanceledException
		var txConflictException *types.TransactionConflictException
		// トランザクションロールバックの理由をログ出力
		if errors.As(err, &txCanceledException) {
			for _, v := range txCanceledException.CancellationReasons {
				t.log.Info(message.I_FW_0003, *v.Code, *v.Message, v.Item)
			}
		} else if errors.As(err, &txConflictException) {
			t.log.Info(message.I_FW_0004, *txConflictException.ErrorCodeOverride, *txConflictException.Message)
		}
		return nil, errors.WithStack(err)
	}
	t.log.Debug("トランザクション終了")
	return output, nil
}
