/*
transaction パッケージは、トランザクション管理に関する機能を提供するパッケージです。
*/
package transaction

import "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

// TODO: ユーティリティ化を検討中
// ContainsConditionalCheckFailed は、TransactionCanceledExceptionの原因に
// ConditionalCheckFailedが含まれているかを判定します。
func ContainsConditionalCheckFailed(txCanceledException *types.TransactionCanceledException) bool {
	for _, reason := range txCanceledException.CancellationReasons {
		if *reason.Code == "ConditionalCheckFailed" {
			return true
		}
	}
	return false
}
