package tables

import "example.com/appbase/pkg/dynamodb/tables"

type Todo struct {
}

func (Todo) InitPk(tableName string) {
	pkKeyPair := &tables.PKKeyPair{
		PartitionKey: "todo_id",
	}
	tables.SetPrimaryKey(tables.DynamoDBTableName(tableName), pkKeyPair)
}
