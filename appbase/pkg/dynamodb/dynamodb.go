/*
dynamodb パッケージは、DynamoDBアクセスに関する機能を提供するパッケージです。
*/
package dynamodb

import "example.com/appbase/pkg/domain"

// ExecuteTransaction は、Serviceの関数serviceFuncの実行前後でDynamoDBトランザクション実行します。
func ExecuteTransaction(serviceFunc domain.ServiceFunc) (interface{}, error) {
	// サービスの実行
	result, err := serviceFunc()
	// TODO: DynamoDBトランザクションオブジェクトが存在する場合はトランザクション実行の実装
	// コンテキスト領域にTransactWriteItemsInputや
	// TransactGetItemsInputがある場合にまとめてトランザクション実行

	if err != nil {
		return nil, err
	}
	return result, nil
}
