package dynamodb

import "example.com/appbase/pkg/domain"

func HandleTransaction(serviceFunc domain.ServiceFunc) (interface{}, error) {
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
