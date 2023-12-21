package tables

import "example.com/appbase/pkg/dynamodb/tables"

type Todo struct {
}

func (Todo) InitPk(tableName tables.DynamoDBTableName) {
	pkKeyPair := &tables.PKKeyPair{
		PartitionKey: "todo_id",
	}
	tables.SetPrimaryKey(tableName, pkKeyPair)
}
