/*
transaction パッケージは、トランザクション管理に関する機能を提供するパッケージです。
*/
package transaction

import (
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// IsTransactionConditionalCheckFailed は、エラーの原因がトランザクション実行中にConditionCheckに失敗
// （TransactionCanceledExceptionが発生しConditionalCheckFailedが含まれている）かどうかを判定します。
func IsTransactionConditionalCheckFailed(err error) bool {
	var txCanceledException *types.TransactionCanceledException
	return errors.As(err, &txCanceledException) && containsConditionalCheckFailed(txCanceledException)
}

// IsTransactionConflict は、エラーの原因がトランザクション実行中にトランザクションの競合が発生
// （TransactionCanceledExceptionが発生しTransactionConflictが含まれている）
// または、通常のDB操作中に、他のトランザクションが実行されてトランザクションの競合が発生した
// （TransactionConflictExceptionが発生） かどうかを判定します。
func IsTransactionConflict(err error) bool {
	var txCanceledException *types.TransactionCanceledException
	var txConflictException *types.TransactionConflictException
	if errors.As(err, &txCanceledException) {
		return containsTransactionConflict(txCanceledException)
	} else if errors.As(err, &txConflictException) {
		return true
	}
	return false
}

// containsConditionalCheckFailed は、TransactionCanceledExceptionの原因に
// ConditionalCheckFailedが含まれている場合はtrueを返します。
func containsConditionalCheckFailed(txCanceledException *types.TransactionCanceledException) bool {
	conditionalCheckFailed := false
	for _, reason := range txCanceledException.CancellationReasons {
		// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/dynamodb/types#TransactionCanceledException
		if *reason.Code == "None" {
			continue
		} else if *reason.Code != "ConditionalCheckFailed" {
			// ConditionalCheckFailed以外のエラーが含まれている場合は、falseを返す
			return false
		}
		// ConditionalCheckFailedが含まれている場合は、trueにする
		conditionalCheckFailed = true
	}
	return conditionalCheckFailed
}

// containsTransactionConflict は、TransactionCanceledExceptionの原因に
// TransactionConflictが含まれているかを判定します。
func containsTransactionConflict(txCanceledException *types.TransactionCanceledException) bool {
	transactionConflict := false
	for _, reason := range txCanceledException.CancellationReasons {
		if *reason.Code == "None" {
			continue
		} else if *reason.Code != "TransactionConflict" {
			//	TransactionConflict以外のエラーが含まれている場合は、trueを返す
			return true
		}
		// TransactionConflictが含まれている場合は、trueにする
		transactionConflict = true
	}
	return transactionConflict
}
