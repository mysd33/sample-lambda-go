// repositoryのパッケージ
package repository

import (
	"app/internal/pkg/entity"
	"app/internal/pkg/message"
	mytables "app/internal/pkg/repository/tables"
	"errors"

	"example.com/appbase/pkg/config"
	mydynamodb "example.com/appbase/pkg/dynamodb"
	"example.com/appbase/pkg/dynamodb/tables"
	myerrors "example.com/appbase/pkg/errors"
	"example.com/appbase/pkg/id"
	"example.com/appbase/pkg/logging"
	"example.com/appbase/pkg/transaction"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

const (
	TODO_TABLE_NAME = "TODO_TABLE_NAME"
)

// NewTodoRepositoryForDynamoDB は、TodoRepositoryを作成します。
func NewTodoRepositoryForDynamoDB(dynamoDBTempalte transaction.TransactinalDynamoDBTemplate,
	accessor transaction.TransactionalDynamoDBAccessor,
	log logging.Logger, config config.Config) TodoRepository {
	// テーブル名の取得
	tableName := tables.DynamoDBTableName(config.Get(TODO_TABLE_NAME))
	// テーブル定義の設定
	mytables.Todo{}.InitPk(tableName)

	return &todoRepositoryImplByDynamoDB{
		dynamodbTemplate: dynamoDBTempalte,
		accessor:         accessor,
		log:              log,
		config:           config,
		tableName:        tableName,
	}
}

// todoRepositoryImplByDynamoDB は、TodoRepositoryを実装する構造体です。
type todoRepositoryImplByDynamoDB struct {
	dynamodbTemplate transaction.TransactinalDynamoDBTemplate
	accessor         transaction.TransactionalDynamoDBAccessor
	log              logging.Logger
	config           config.Config
	tableName        tables.DynamoDBTableName
}

func (tr *todoRepositoryImplByDynamoDB) FindOne(todoId string) (*entity.Todo, error) {
	// AWS SDK for Go v2 Migration
	// https://docs.aws.amazon.com/ja_jp/code-library/latest/ug/go_2_dynamodb_code_examples.html
	// https://github.com/awsdocs/aws-doc-sdk-examples/tree/main/gov2/dynamodb

	// 従来のDynamoDBAccessorを使ったコード
	todo := entity.Todo{ID: todoId}
	key, err := todo.GetKey()
	if err != nil {
		return nil, myerrors.NewSystemError(err, message.E_EX_9001)
	}
	// Itemの取得
	result, err := tr.accessor.GetItemSdk(&dynamodb.GetItemInput{
		TableName: aws.String(tr.config.Get(TODO_TABLE_NAME)),
		Key:       key,
	})
	if err != nil {
		return nil, myerrors.NewSystemError(err, message.E_EX_9001)
	}
	// レコード未取得の場合
	if len(result.Item) == 0 {
		return nil, myerrors.NewBusinessError(message.W_EX_8002, todoId)
	}
	err = attributevalue.UnmarshalMap(result.Item, &todo)
	if err != nil {
		return nil, myerrors.NewSystemError(err, message.E_EX_9001)
	}
	return &todo, nil

	//TODO:	DynamoDBTemplateを使ったコード
	/*
		var todo *entity.Todo
		err := tr.dynamodbTemplate.FindOneByPrimaryKey(tr.tableName, input, outEntity)
		if err != nil {
			if errors.Is(err, mydynamodb.ErrRecordNotFound) {
				// レコード未取得の場合
				return nil, myerrors.NewBusinessError(message.W_EX_8002, todoId)
			}
			return nil, myerrors.NewSystemError(err, message.E_EX_9001)
		}
		return todo, nil
	*/
}

func (tr *todoRepositoryImplByDynamoDB) CreateOne(todo *entity.Todo) (*entity.Todo, error) {
	// AWS SDK for Go v2 Migration
	// https://docs.aws.amazon.com/ja_jp/code-library/latest/ug/go_2_dynamodb_code_examples.html
	// https://github.com/awsdocs/aws-doc-sdk-examples/tree/main/gov2/dynamodb

	// ID採番
	todoId := id.GenerateId()
	todo.ID = todoId
	// 従来のDynamoDBAccessorを使ったコード
	/*
		av, err := attributevalue.MarshalMap(todo)
		if err != nil {
			return nil, myerrors.NewSystemError(err, message.E_EX_9001)
		}
		input := &dynamodb.PutItemInput{
			Item:      av,
			TableName: aws.String(tr.tableName),
		}
		// Itemの登録
		_, err = tr.accessor.PutItemSdk(input)
		if err != nil {
			return nil, myerrors.NewSystemError(err, message.E_EX_9001)
		}
	*/
	// DynamoDBTemplateを使ったコード
	err := tr.dynamodbTemplate.CreateOne(tr.tableName, todo)
	if err != nil {
		if errors.Is(err, mydynamodb.ErrKeyDuplicaiton) {
			// キーの重複の場合
			return nil, myerrors.NewBusinessError(message.W_EX_8003, todoId)
		}
		return nil, myerrors.NewSystemError(err, message.E_EX_9001)
	}
	return todo, nil

}

// CreateOneTx implements TodoRepository.
func (tr *todoRepositoryImplByDynamoDB) CreateOneTx(todo *entity.Todo) (*entity.Todo, error) {
	// ID採番
	todoId := id.GenerateId()
	todo.ID = todoId
	// 従来のDynamoDBAccessorを使ったコード
	/*
		av, err := attributevalue.MarshalMap(todo)
		if err != nil {
			return nil, myerrors.NewSystemError(err, message.E_EX_9001)
		}
		put := &types.Put{
			Item:      av,
			TableName: aws.String(tr.tableName),
		}
		// TransactWriteItemの追加
		input := &types.TransactWriteItem{Put: put}
		tr.accessor.AppendTransactWriteItem(input)
	*/
	// DynamoDBTemplateを使ったコード
	tr.dynamodbTemplate.CreateOneWithTransaction(tr.tableName, todo)

	return todo, nil
}
