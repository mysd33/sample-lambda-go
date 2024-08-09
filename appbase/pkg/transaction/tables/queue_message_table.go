/*
tables パッケージは、キューメッセージ管理テーブルに関連するテーブル情報を提供します。
*/
package tables

import "example.com/appbase/pkg/dynamodb/tables"

// QueueMessageTable は、キューメッセージ管理テーブルのテーブル情報を提供します。
type QueueMessageTable struct {
}

// InitPK は、冪等性テーブルのプライマリキーを初期化します。
func (QueueMessageTable) InitPK(tableName tables.DynamoDBTableName) {
	pkKeyPair := &tables.PKKeyPair{
		PartitionKey: "message_id",
	}
	tables.SetPrimaryKey(tableName, pkKeyPair)
}
