/*
dynamodb パッケージは、DynamoDBアクセスに関する機能を提供するパッケージです。
*/
package dynamodb

import (
	"example.com/appbase/pkg/domain"
	"example.com/appbase/pkg/logging"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/cockroachdb/errors"
)

// TransactionManager はトランザクションを管理するインタフェースです
type TransactionManager interface {
	// AppendTransactWriteItemは、トランザクション書き込みしたい場合に対象のTransactWriteItemを追加します。
	AppendTransactWriteItem(item *types.TransactWriteItem)
	// ExecuteTransaction は、Serviceの関数serviceFuncの実行前後でDynamoDBトランザクション実行します。
	ExecuteTransaction(serviceFunc domain.ServiceFunc) (interface{}, error)
}

// NewTransactionManager は、TransactionManagerを作成します
func NewTransactionManager(log logging.Logger, dynamodbAccessor DynamoDBAccessor) TransactionManager {
	return &defaultTransactionManager{log: log, dynamodbAccessor: dynamodbAccessor}
}

type defaultTransactionManager struct {
	log              logging.Logger
	dynamodbAccessor DynamoDBAccessor
	// 書き込みトランザクション
	transactWriteItems []types.TransactWriteItem
	// TODO: 読み込みトランザクションTransactGetItems
	// transactGetItems []types.TransactGetItem
}

// AppendTransactWriteItem implements TransactionManager.
func (tm *defaultTransactionManager) AppendTransactWriteItem(item *types.TransactWriteItem) {
	tm.transactWriteItems = append(tm.transactWriteItems, *item)
}

// ExecuteTransaction implements TransactionManager.
func (tm *defaultTransactionManager) ExecuteTransaction(serviceFunc domain.ServiceFunc) (interface{}, error) {
	// トランザクションの開始
	tm.startTransaction()
	// サービスの実行
	result, err := serviceFunc()
	// DynamoDBのトランザクションを終了
	_, err = tm.endTransaction(err)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// checkTransactWriteItems は、TransactWriteItemが存在するかを確認します。
func (tm *defaultTransactionManager) checkTransactWriteItems() bool {
	return len(tm.transactWriteItems) > 0
}

// clearTransactWriteItems() は、TransactWriteItemをクリアします。
func (tm *defaultTransactionManager) clearTransactWriteItems() {
	tm.transactWriteItems = nil
}

func (tm *defaultTransactionManager) startTransaction() {
	tm.log.Debug("トランザクション開始")
	tm.dynamodbAccessor.startTransaction(tm)
}

// endTransaction は、エラーがなければ、AWS SDKによるTransactionWriteItemsを実行しトランザクション実行し、エラーがある場合には実行しません。
// TODO: TransactGetItemsの考慮
func (tm *defaultTransactionManager) endTransaction(err error) (*dynamodb.TransactWriteItemsOutput, error) {
	if !tm.checkTransactWriteItems() {
		tm.log.Debug("トランザクション処理なし")
		return nil, nil
	}
	// 処理結果がどんな場合でもTransactWriteItemをクリア
	defer tm.clearTransactWriteItems()
	if err != nil {
		tm.log.Debug("業務処理エラーでトランザクションロールバック")
		// Serviceの処理結果がエラー場合は、トランザクションを実行せず、元のエラーを返却し終了
		return nil, err
	}
	// トランザクション実行
	output, err := tm.dynamodbAccessor.transactWriteItemsSDK(tm.transactWriteItems)
	if err != nil {
		tm.log.Debug("トランザクション実行失敗でロールバック")
		return nil, errors.WithStack(err)
	}
	tm.log.Debug("トランザクション終了")
	return output, nil
}
