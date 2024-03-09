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
	return errors.As(err, &txCanceledException) && ContainsConditionalCheckFailed(txCanceledException)
}

// IsTransactionConflict は、エラーの原因がトランザクション実行中にトランザクションの競合が発生
// （TransactionCanceledExceptionが発生しTransactionConflictが含まれている）かどうかを判定します。
func IsTransactionConflict(err error) bool {
	var txCanceledException *types.TransactionCanceledException
	return errors.As(err, &txCanceledException) && ContainsTransactionConflict(txCanceledException)
}

// ContainsConditionalCheckFailed は、TransactionCanceledExceptionの原因に
// ConditionalCheckFailedが含まれているかを判定します。
func ContainsConditionalCheckFailed(txCanceledException *types.TransactionCanceledException) bool {
	for _, reason := range txCanceledException.CancellationReasons {
		// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/dynamodb/types#TransactionCanceledException
		if *reason.Code == "ConditionalCheckFailed" {
			return true
		}
	}
	return false
}

// ContainsTransactionConflict は、TransactionCanceledExceptionの原因に
// TransactionConflictが含まれているかを判定します。
func ContainsTransactionConflict(txCanceledException *types.TransactionCanceledException) bool {
	for _, reason := range txCanceledException.CancellationReasons {
		if *reason.Code == "TransactionConflict" {
			return true
		}
	}
	return false
}
