/*
transaction パッケージは、トランザクション管理に関する機能を提供するパッケージです。
*/
package transaction

import (
	"example.com/appbase/pkg/dynamodb"
	"example.com/appbase/pkg/dynamodb/criteria"
	"example.com/appbase/pkg/dynamodb/tables"
)

type TransactinalDynamoDBTemplate interface {
	dynamodb.DynamoDBTemplate
	CreateOneWithTransaction(tableName tables.DynamoDBTableName, inputEntity any) error
	UpdateOneWithTransaction(tableName tables.DynamoDBTableName, input criteria.UpdateInput) error
	DeleteOneWithTransaction(tableName tables.DynamoDBTableName, input criteria.DeleteInput) error
}

//TODO:　TransactinalDynamoDBTemplateインタフェースの実装
