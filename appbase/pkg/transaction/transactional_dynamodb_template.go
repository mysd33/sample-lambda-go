/*
transaction パッケージは、トランザクション管理に関する機能を提供するパッケージです。
*/
package transaction

import (
	"example.com/appbase/pkg/dynamodb"
	"example.com/appbase/pkg/dynamodb/criteria"
	"example.com/appbase/pkg/dynamodb/tables"
	"example.com/appbase/pkg/logging"
)

// TransactionalDynamoDBTemplate は、トランザクション管理対応のDynamoDBアクセスを定型化した高次のインタフェースです。
type TransactinalDynamoDBTemplate interface {
	dynamodb.DynamoDBTemplate
	CreateOneWithTransaction(tableName tables.DynamoDBTableName, inputEntity any) error
	UpdateOneWithTransaction(tableName tables.DynamoDBTableName, input criteria.UpdateInput) error
	DeleteOneWithTransaction(tableName tables.DynamoDBTableName, input criteria.DeleteInput) error
}

func NewTransactionalDynamoDBTemplate(log logging.Logger,
	transactionalDynamoDBAccessor TransactionalDynamoDBAccessor) TransactinalDynamoDBTemplate {
	dynamodbTemplate := dynamodb.NewDynamoDBTemplate(log, transactionalDynamoDBAccessor)
	return &defaultTransactionalDynamoDBTemplate{
		log:                           log,
		dynamodbTemplate:              dynamodbTemplate,
		transactionalDynamoDBAccessor: transactionalDynamoDBAccessor,
	}
}

//TODO:　TransactinalDynamoDBTemplateインタフェースの実装

type defaultTransactionalDynamoDBTemplate struct {
	log                           logging.Logger
	dynamodbTemplate              dynamodb.DynamoDBTemplate
	transactionalDynamoDBAccessor TransactionalDynamoDBAccessor
}

// CreateOne implements TransactinalDynamoDBTemplate.
func (t *defaultTransactionalDynamoDBTemplate) CreateOne(tableName tables.DynamoDBTableName, inputEntity any) error {
	return t.dynamodbTemplate.CreateOne(tableName, inputEntity)
}

// FindOneByPrimaryKey implements TransactinalDynamoDBTemplate.
func (t *defaultTransactionalDynamoDBTemplate) FindOneByPrimaryKey(tableName tables.DynamoDBTableName, input criteria.PkOnlyQueryInput, outEntity any) error {
	return t.dynamodbTemplate.FindOneByPrimaryKey(tableName, input, outEntity)
}

// FindSomeByGSI implements TransactinalDynamoDBTemplate.
func (t *defaultTransactionalDynamoDBTemplate) FindSomeByGSI(tableName tables.DynamoDBTableName, input criteria.GsiQueryInput, outEntities any) error {
	return t.dynamodbTemplate.FindSomeByGSI(tableName, input, outEntities)
}

// FindSomeByPrimaryKey implements TransactinalDynamoDBTemplate.
func (t *defaultTransactionalDynamoDBTemplate) FindSomeByPrimaryKey(tableName tables.DynamoDBTableName, input criteria.PkOnlyQueryInput, outEntities any) error {
	return t.dynamodbTemplate.FindSomeByPrimaryKey(tableName, input, outEntities)
}

// UpdateOne implements TransactinalDynamoDBTemplate.
func (t *defaultTransactionalDynamoDBTemplate) UpdateOne(tableName tables.DynamoDBTableName, input criteria.UpdateInput) error {
	return t.dynamodbTemplate.UpdateOne(tableName, input)
}

// DeleteOne implements TransactinalDynamoDBTemplate.
func (t *defaultTransactionalDynamoDBTemplate) DeleteOne(tableName tables.DynamoDBTableName, input criteria.DeleteInput) error {
	return t.dynamodbTemplate.DeleteOne(tableName, input)
}

// CreateOneWithTransaction implements TransactinalDynamoDBTemplate.
func (t *defaultTransactionalDynamoDBTemplate) CreateOneWithTransaction(tableName tables.DynamoDBTableName, inputEntity any) error {
	panic("unimplemented")
}

// UpdateOneWithTransaction implements TransactinalDynamoDBTemplate.
func (*defaultTransactionalDynamoDBTemplate) UpdateOneWithTransaction(tableName tables.DynamoDBTableName, input criteria.UpdateInput) error {
	panic("unimplemented")
}

// DeleteOneWithTransaction implements TransactinalDynamoDBTemplate.
func (*defaultTransactionalDynamoDBTemplate) DeleteOneWithTransaction(tableName tables.DynamoDBTableName, input criteria.DeleteInput) error {
	panic("unimplemented")
}
