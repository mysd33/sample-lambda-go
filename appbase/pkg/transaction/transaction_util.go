/*
transaction パッケージは、トランザクション管理に関する機能を提供するパッケージです。
*/
package transaction

import (
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const (
	reasonCodeNone                   = "None"
	reasonCodeConditionalCheckFailed = "ConditionalCheckFailed"
	reasonCodeTransactionConflict    = "TransactionConflict"
)

// IsTransactionConditionalCheckFailed は、エラーの原因がトランザクション実行中にConditionCheckに失敗
// （TransactionCanceledExceptionが発生しConditionalCheckFailedのみが含まれている）かどうかを判定します。
// トランザクションのキャンセルの原因については以下のドキュメントを参照してください。
// https://docs.aws.amazon.com/ja_jp/amazondynamodb/latest/APIReference/API_TransactWriteItems.html
func IsTransactionConditionalCheckFailed(err error) bool {
	var txCanceledException *types.TransactionCanceledException
	return errors.As(err, &txCanceledException) && containsOnlyTargetCancellationReason(txCanceledException, reasonCodeConditionalCheckFailed)
}

// IsTransactionConflict は、エラーの原因がトランザクション実行中にトランザクションの競合が発生
// （TransactionCanceledExceptionが発生しTransactionConflictのみが含まれている）
// または、通常のDB操作中に、他のトランザクションが実行されてトランザクションの競合が発生した
// （TransactionConflictExceptionが発生） かどうかを判定します。
// トランザクションのキャンセルの原因については以下のドキュメントを参照してください。
// https://docs.aws.amazon.com/ja_jp/amazondynamodb/latest/APIReference/API_TransactWriteItems.html
func IsTransactionConflict(err error) bool {
	var txCanceledException *types.TransactionCanceledException
	var txConflictException *types.TransactionConflictException
	if errors.As(err, &txCanceledException) {
		return containsOnlyTargetCancellationReason(txCanceledException, reasonCodeTransactionConflict)
	}
	return errors.As(err, &txConflictException)
}

// IsTransactionConditionalCheckFailedOrTransactionConflict は、エラーの原因がトランザクション実行中にConditionCheckに失敗
// またはトランザクション実行中にトランザクションの競合が発生したかどうかを判定します。
// （TransactionCanceledExceptionが発生し、いずれか、両方のエラーが混在して含まれている）
// または、通常のDB操作中に、他のトランザクションが実行されてトランザクションの競合が発生した
// （TransactionConflictExceptionが発生） かどうかを判定します。
// トランザクションのキャンセルの原因については以下のドキュメントを参照してください。
// https://docs.aws.amazon.com/ja_jp/amazondynamodb/latest/APIReference/API_TransactWriteItems.html
func IsTransactionConditionalCheckFailedOrTransactionConflict(err error) bool {
	var txCanceledException *types.TransactionCanceledException
	var txConflictException *types.TransactionConflictException
	if errors.As(err, &txCanceledException) {
		contains := false
		for _, reason := range txCanceledException.CancellationReasons {
			if *reason.Code == reasonCodeNone {
				// Noneの場合は、スキップ
				continue
				// 対象の原因以外のエラーが含まれている場合は、falseを返す
			} else if *reason.Code != reasonCodeConditionalCheckFailed && *reason.Code != reasonCodeTransactionConflict {
				return false
			}
			// 対象の原因が含まれている場合は、trueにする
			contains = true
		}
		return contains
	}
	return errors.As(err, &txConflictException)
}

// containsOnlyTargetCancellationReason は、TransactionCanceledExceptionの原因に指定された原因が含まれているかを判定します。
// 指定された原因以外のエラーが含まれている場合は、falseを返します。
func containsOnlyTargetCancellationReason(txCanceledException *types.TransactionCanceledException, targetReason string) bool {
	contains := false
	for _, reason := range txCanceledException.CancellationReasons {
		if *reason.Code == reasonCodeNone {
			// Noneの場合は、スキップ
			continue
		} else if *reason.Code != targetReason {
			//	対象の原因以外のエラーが含まれている場合は、falseを返す
			return false
		}
		// 対象の原因が含まれている場合は、trueにする
		contains = true
	}
	return contains
}
