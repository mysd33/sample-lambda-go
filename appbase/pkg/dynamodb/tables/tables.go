/*
tables パッケージは、DynamoDBのテーブル定義のためのパッケージです。
*/
package tables

type DynamoDBTableName string

var pkMap map[DynamoDBTableName]*PKKeyPair

type PK struct {
	PartitionKey string
	SortKey      *string
}

//TODO: 不要なインタフェースかも
type Tables interface {
	initPk(tableName string)
}

type PKKeyPair struct {
	PartitionKey string
	SortKey      *string
}

func init() {
	if pkMap == nil {
		pkMap = make(map[DynamoDBTableName]*PKKeyPair)
	}
}

func GetPrimaryKey(tableName DynamoDBTableName) *PKKeyPair {
	return pkMap[tableName]
}

func SetPrimaryKey(tableName DynamoDBTableName, primaryKey *PKKeyPair) {
	if pkMap == nil {
		pkMap = make(map[DynamoDBTableName]*PKKeyPair)
	}
	pkMap[tableName] = primaryKey
}
