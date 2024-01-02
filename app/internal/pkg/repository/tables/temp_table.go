// tables パッケージは、DynamoDB のテーブルの定義を行うパッケージです。
package tables

import "example.com/appbase/pkg/dynamodb/tables"

// Temp は、temp テーブルの定義を行う構造体です。
type Temp struct {
}

// InitPK は、テーブルの主キーを設定します。
func (Temp) InitPK(tableName tables.DynamoDBTableName) {
	pkKeyPair := &tables.PKKeyPair{
		PartitionKey: "id",
	}
	tables.SetPrimaryKey(tableName, pkKeyPair)
}
