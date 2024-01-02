package tables

import "example.com/appbase/pkg/dynamodb/tables"

type Temp struct {
}

func (Temp) InitPK(tableName tables.DynamoDBTableName) {
	pkKeyPair := &tables.PKKeyPair{
		PartitionKey: "id",
	}
	tables.SetPrimaryKey(tableName, pkKeyPair)
}
