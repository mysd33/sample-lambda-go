// repositoryのパッケージ
package repository

import (
	"app/internal/pkg/entity"
	"app/internal/pkg/message"
	mytables "app/internal/pkg/repository/tables"

	"example.com/appbase/pkg/config"
	mydynamodb "example.com/appbase/pkg/dynamodb"
	"example.com/appbase/pkg/dynamodb/input"
	"example.com/appbase/pkg/dynamodb/tables"
	"example.com/appbase/pkg/errors"
	"example.com/appbase/pkg/id"
	"example.com/appbase/pkg/logging"

	"example.com/appbase/pkg/transaction"
)

const (
	TODO_TABLE_NAME = "TODO_TABLE_NAME"
)

// NewTodoRepositoryForDynamoDB は、TodoRepositoryを作成します。
func NewTodoRepositoryForDynamoDB(dynamoDBTempalte transaction.TransactionalDynamoDBTemplate,
	accessor transaction.TransactionalDynamoDBAccessor,
	log logging.Logger, config config.Config,
	id id.IDGenerator) TodoRepository {
	// テーブル名の取得
	tableName := tables.DynamoDBTableName(config.Get(TODO_TABLE_NAME, "todo"))
	// テーブル定義の設定
	mytables.Todo{}.InitPK(tableName)
	// プライマリキーの設定
	primaryKey := tables.GetPrimaryKey(tableName)

	return &todoRepositoryImplByDynamoDB{
		dynamodbTemplate: dynamoDBTempalte,
		accessor:         accessor,
		log:              log,
		config:           config,
		tableName:        tableName,
		primaryKey:       primaryKey,
		id:               id,
	}
}

// todoRepositoryImplByDynamoDB は、TodoRepositoryを実装する構造体です。
type todoRepositoryImplByDynamoDB struct {
	dynamodbTemplate transaction.TransactionalDynamoDBTemplate
	accessor         transaction.TransactionalDynamoDBAccessor
	log              logging.Logger
	config           config.Config
	tableName        tables.DynamoDBTableName
	primaryKey       *tables.PKKeyPair
	id               id.IDGenerator
}

func (tr *todoRepositoryImplByDynamoDB) FindOne(todoId string) (*entity.Todo, error) {
	// DynamoDBTemplateを使ったコード
	input := input.PKOnlyQueryInput{
		PrimaryKey: input.PrimaryKey{
			PartitionKey: input.Attribute{
				Name:  tr.primaryKey.PartitionKey,
				Value: todoId,
			},
		},
	}
	var todo entity.Todo
	// Itemの取得
	err := tr.dynamodbTemplate.FindOneByTableKey(tr.tableName, input, &todo)
	if err != nil {
		if errors.Is(err, mydynamodb.ErrRecordNotFound) {
			// レコード未取得の場合
			return nil, errors.NewBusinessError(message.W_EX_8002, todoId)
		}
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}

	// 従来のDynamoDBAccessorを使ったコード
	// AWS SDK for Go v2 Migration
	// https://docs.aws.amazon.com/ja_jp/code-library/latest/ug/go_2_dynamodb_code_examples.html
	// https://github.com/awsdocs/aws-doc-sdk-examples/tree/main/gov2/dynamodb
	/*
		todo := entity.Todo{ID: todoId}
		key, err := todo.GetKey()
		if err != nil {
			return nil, errors.NewSystemError(err, message.E_EX_9001)
		}
		// Itemの取得
		result, err := tr.accessor.GetItemSdk(&dynamodb.GetItemInput{
			TableName: aws.String(tr.tableName),
			Key:       key,
		})
		if err != nil {
			return nil, errors.NewSystemError(err, message.E_EX_9001)
		}
		// レコード未取得の場合
		if len(result.Item) == 0 {
			return nil, errors.NewBusinessError(message.W_EX_8002, todoId)
		}
		err = attributevalue.UnmarshalMap(result.Item, &todo)
		if err != nil {
			return nil, errors.NewSystemError(err, message.E_EX_9001)
		}
	*/
	return &todo, nil
}

func (tr *todoRepositoryImplByDynamoDB) CreateOne(todo *entity.Todo) (*entity.Todo, error) {
	// ID採番
	todoId, err := tr.id.GenerateUUID()
	if err != nil {
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}
	//todoId := "dummy"
	todo.ID = todoId
	// DynamoDBTemplateを使ったコード
	err = tr.dynamodbTemplate.CreateOne(tr.tableName, todo)
	if err != nil {
		if errors.Is(err, mydynamodb.ErrKeyDuplicaiton) {
			// キーの重複の場合
			return nil, errors.NewBusinessError(message.W_EX_8003, todoId)
		}
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}
	// 従来のDynamoDBAccessorを使ったコード
	// AWS SDK for Go v2 Migration
	// https://docs.aws.amazon.com/ja_jp/code-library/latest/ug/go_2_dynamodb_code_examples.html
	// https://github.com/awsdocs/aws-doc-sdk-examples/tree/main/gov2/dynamodb
	/*
		av, err := attributevalue.MarshalMap(todo)
		if err != nil {
			return nil, errors.NewSystemError(err, message.E_EX_9001)
		}
		input := &dynamodb.PutItemInput{
			Item:      av,
			TableName: aws.String(tr.tableName),
		}
		// Itemの登録
		_, err = tr.accessor.PutItemSdk(input)
		if err != nil {
			return nil, errors.NewSystemError(err, message.E_EX_9001)
		}
	*/
	return todo, nil

}

// CreateOneTx implements TodoRepository.
func (tr *todoRepositoryImplByDynamoDB) CreateOneTx(todo *entity.Todo) (*entity.Todo, error) {
	// ID採番
	todoId, err := tr.id.GenerateUUID()
	if err != nil {
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}
	//todoId := "dummy"
	todo.ID = todoId
	// DynamoDBTemplateを使ったコード
	err = tr.dynamodbTemplate.CreateOneWithTransaction(tr.tableName, todo)
	if err != nil {
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}

	// 従来のDynamoDBAccessorを使ったコード
	/*
		av, err := attributevalue.MarshalMap(todo)
		if err != nil {
			return nil, errors.NewSystemError(err, message.E_EX_9001)
		}
		put := &types.Put{
			Item:      av,
			TableName: aws.String(tr.tableName),
		}
		// TransactWriteItemの追加
		input := &types.TransactWriteItem{Put: put}
		err := tr.accessor.AppendTransactWriteItem(input)
		if err != nil {
			return nil, errors.NewSystemError(err, message.E_EX_9001)
		}
	*/
	return todo, nil
}
