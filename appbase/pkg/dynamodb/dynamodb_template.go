/*
dynamodb パッケージは、DynamoDBアクセスに関する機能を提供するパッケージです。
*/
package dynamodb

import (
	"example.com/appbase/pkg/dynamodb/criteria"
	"example.com/appbase/pkg/dynamodb/tables"
)

type DynamoDBTemplate interface {
	CreateOne(tableName tables.DynamoDBTableName, inputEntity any) error
	FindOneByPrimaryKey(tableName tables.DynamoDBTableName, input criteria.PkOnlyQueryInput, outEntity any) error
	FindSomeByPrimaryKey(tableName tables.DynamoDBTableName, input criteria.PkOnlyQueryInput, outEntities any) error
	FindSomeByGSI(tableName tables.DynamoDBTableName, input criteria.GsiQueryInput, outEntities any) error
	UpdateOne(tableName tables.DynamoDBTableName, input criteria.UpdateInput) error
	DeleteOne(tableName tables.DynamoDBTableName, input criteria.DeleteInput) error
}

//TODO:　DynamoDBTemplateインタフェースの実装
