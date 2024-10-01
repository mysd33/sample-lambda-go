/*
gsi パッケージは、DynamoDBのGSI定義のためのパッケージです。
*/
package gsi

import "example.com/appbase/pkg/dynamodb/tables"

type DynamoDBGSIName string

var gsiMap map[tables.DynamoDBTableName]*GSIDefinition

type GSIDefinition struct {
	IndexMap map[DynamoDBGSIName]*GSIKeyPair
}

type GSIKeyPair struct {
	PartitionKey string
	SortKey      string
}

func init() {
	if gsiMap == nil {
		gsiMap = make(map[tables.DynamoDBTableName]*GSIDefinition)
	}
}

func GetGSIKeyPair(tableName tables.DynamoDBTableName, indexName DynamoDBGSIName) *GSIKeyPair {
	gsis := gsiMap[tableName]
	return gsis.IndexMap[indexName]
}

func AddGSIKeyPair(tableName tables.DynamoDBTableName, indexName DynamoDBGSIName, gsiKeyPair *GSIKeyPair) {
	if gsiMap == nil {
		gsiMap = make(map[tables.DynamoDBTableName]*GSIDefinition)
	}
	gsis, ok := gsiMap[tableName]
	if !ok {
		gsis = &GSIDefinition{IndexMap: make(map[DynamoDBGSIName]*GSIKeyPair)}
		gsiMap[tableName] = gsis
	}
	gsis.IndexMap[indexName] = gsiKeyPair

}
