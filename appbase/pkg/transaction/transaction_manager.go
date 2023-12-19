/*
transaction パッケージは、トランザクション管理に関する機能を提供するパッケージです。
*/
package transaction

import (
	"example.com/appbase/pkg/domain"
	"example.com/appbase/pkg/logging"
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
	transction.start(tm.dynamodbAccessor, tm.sqsAccessor)
	// サービスの実行
	result, err := serviceFunc()
	// DynamoDBのトランザクションを終了
	_, err = transction.end(err)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// TODO: Mock化しやすいよう全て公開インタフェース化する

// transactionは トランザクションを表すインタフェースです
type transaction interface {
	// start は、トランザクションを開始します。
	start(dynamodbAccessor TransactionalDynamoDBAccessor, sqsAccessor TransactionalSQSAccessor)
	// appendTransactWriteItemは、DBへトランザクション書き込みしたい場合に対象のTransactWriteItemを追加します。
	appendTransactWriteItem(item *types.TransactWriteItem)
	// appendTransactMessageは、SQSへトランザクション管理してメッセージ送信したい場合に対象のMessageを追加します。
	appendTransactMessage(message *Message)
	// checkTransactWriteItems は、TransactWriteItemが存在するかを確認します。
	checkTransactWriteItems() bool
	// end は、エラーがなければ、AWS SDKによるTransactionWriteItemsを実行しトランザクション実行し、エラーがある場合には実行しません。
	end(err error) (*dynamodb.TransactWriteItemsOutput, error)
}

// newTrasactionは 新しいTransactionを作成します。
func newTrasaction(log logging.Logger) transaction {
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

// start implements Transaction.
func (t *defaultTransaction) start(dynamodbAccessor TransactionalDynamoDBAccessor, sqsAccessor TransactionalSQSAccessor) {
	t.log.Debug("トランザクション開始")
	t.dynamodbAccessor = dynamodbAccessor
	t.sqsAccessor = sqsAccessor
	dynamodbAccessor.startTransaction(t)
	sqsAccessor.startTransaction(t)
}

// appendTransactWriteItem implements Transaction.
func (t *defaultTransaction) appendTransactWriteItem(item *types.TransactWriteItem) {
	t.transactWriteItems = append(t.transactWriteItems, *item)
}

// appendTransactMessage implements transaction.
func (t *defaultTransaction) appendTransactMessage(message *Message) {
	t.messages = append(t.messages, message)
}

// checkTransactWriteItems implements Transaction.
func (t *defaultTransaction) checkTransactWriteItems() bool {
	return len(t.transactWriteItems) > 0
}

// endTransaction implements Transaction.
func (t *defaultTransaction) end(err error) (*dynamodb.TransactWriteItemsOutput, error) {
	if t.sqsAccessor != nil {
		err := t.sqsAccessor.transactSendMessages(t.messages)
		if err != nil {
			t.log.Debug("SQSのメッセージ送信失敗でロールバック")
			return nil, errors.WithStack(err)
		}
	}

	if !t.checkTransactWriteItems() {
		t.log.Debug("トランザクション処理なし")
		return nil, err
	}
	// 処理結果がどんな場合でもDynamoDBAccessorとSQSAccessorのトランザクションを開放
	defer func() {
		t.dynamodbAccessor.endTransaction()
		if t.sqsAccessor != nil {
			t.sqsAccessor.endTransaction()
		}
	}()

	if err != nil {
		t.log.Debug("業務処理エラーでトランザクションロールバック")
		// Serviceの処理結果がエラー場合は、トランザクションを実行せず、元のエラーを返却し終了
		return nil, err
	}
	// DynamoDBトランザクション実行
	output, err := t.dynamodbAccessor.transactWriteItemsSDK(t.transactWriteItems)
	if err != nil {
		t.log.Debug("トランザクション実行失敗でロールバック")
		return nil, errors.WithStack(err)
	}
	t.log.Debug("トランザクション終了")
	return output, nil
}
