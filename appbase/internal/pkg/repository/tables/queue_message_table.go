package tables

import "example.com/appbase/pkg/dynamodb/tables"

type QueueMessageTable struct {
}

func (QueueMessageTable) InitPK(tableName tables.DynamoDBTableName) {
	pkKeyPair := &tables.PKKeyPair{
		PartitionKey: "message_id",
	}
	tables.SetPrimaryKey(tableName, pkKeyPair)
}
