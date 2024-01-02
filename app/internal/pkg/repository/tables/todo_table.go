// tables パッケージは、DynamoDB のテーブルの定義を行うパッケージです。
package tables

import "example.com/appbase/pkg/dynamodb/tables"

// Todo は、todo テーブルの定義を行う構造体です。
type Todo struct {
}

// InitPK は、テーブルの主キーを設定します。
func (Todo) InitPK(tableName tables.DynamoDBTableName) {
	pkKeyPair := &tables.PKKeyPair{
		PartitionKey: "todo_id",
	}
	tables.SetPrimaryKey(tableName, pkKeyPair)
}
