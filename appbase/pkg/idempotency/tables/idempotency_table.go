/*
tables パッケージは、冪等性管理テーブルに関連するテーブル情報を提供します。
*/
package tables

import "example.com/appbase/pkg/dynamodb/tables"

// 冪等性管理テーブルの属性名
const (
	IDEMPOTENCY_KEY   = "idempotency_key"
	EXPIRY            = "expiry"
	INPROGRESS_EXPIRY = "inprogress_expiry"
	STATUS            = "status"
)

// 冪等性管理テーブルのステータス
const (
	STATUS_INPROGRESS = "INPROGRESS"
	STATUS_COMPLETE   = "COMPLETE"
)

// IdempotencyTable は、冪等性管理テーブルのテーブル情報を提供します。
type IdempotencyTable struct {
}

// InitPK は、冪等性管理テーブルのプライマリキーを初期化します。
func (IdempotencyTable) InitPK(tableName tables.DynamoDBTableName) {

	pkKeyPair := &tables.PKKeyPair{
		PartitionKey: IDEMPOTENCY_KEY,
	}
	tables.SetPrimaryKey(tableName, pkKeyPair)
}
