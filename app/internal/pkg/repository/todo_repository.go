// repositoryのパッケージ
package repository

import (
	"app/internal/pkg/entity"
	"context"

	"example.com/appbase/pkg/apcontext"
	mydynamodb "example.com/appbase/pkg/dynamodb"
	"example.com/appbase/pkg/id"
	"example.com/appbase/pkg/logging"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/pkg/errors"
)

// TodoRepository は、Todoを管理するRepositoryインタフェースです。
type TodoRepository interface {
	// GetTodo は、todoIdが一致するTodoを取得します。
	GetTodo(todoId string) (*entity.Todo, error)
	// PutTodo は、指定されたTodoを登録します。
	PutTodo(todo *entity.Todo) (*entity.Todo, error)
}

// NewTodoRepository は、TodoRepositoryを作成します。
func NewTodoRepository(log logging.Logger) (TodoRepository, error) {
	dynamodbClient, err := mydynamodb.CreateDynamoDBClient()
	if err != nil {
		return nil, err
	}
	return &TodoRepositoryImpl{dynamodbClient: dynamodbClient, log: log}, nil
}

// TodoRepositoryImpl は、TodoRepositoryを実装する構造体です。
type TodoRepositoryImpl struct {
	dynamodbClient *dynamodb.Client
	log            logging.Logger
}

func (tr *TodoRepositoryImpl) GetTodo(todoId string) (*entity.Todo, error) {
	return tr.doGetTodo(todoId, apcontext.Context)
}

func (tr *TodoRepositoryImpl) doGetTodo(todoId string, ctx context.Context) (*entity.Todo, error) {
	// AWS SDK for Go v2 Migration
	// https://docs.aws.amazon.com/ja_jp/code-library/latest/ug/go_2_dynamodb_code_examples.html
	// https://github.com/awsdocs/aws-doc-sdk-examples/tree/main/gov2/dynamodb
	//Itemの取得（X-Rayトレース）
	todo := entity.Todo{ID: todoId}
	key, err := todo.GetKey()
	if err != nil {
		return nil, errors.Wrapf(err, "fail to get key")
	}
	result, err := tr.dynamodbClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(todoTable),
		Key:       key,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get item")
	}
	err = attributevalue.UnmarshalMap(result.Item, &todo)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to marshal item")
	}
	return &todo, nil
}

func (tr *TodoRepositoryImpl) PutTodo(todo *entity.Todo) (*entity.Todo, error) {
	return tr.doPutTodo(todo, apcontext.Context)
}

func (tr *TodoRepositoryImpl) doPutTodo(todo *entity.Todo, ctx context.Context) (*entity.Todo, error) {
	// AWS SDK for Go v2 Migration
	// https://docs.aws.amazon.com/ja_jp/code-library/latest/ug/go_2_dynamodb_code_examples.html
	// https://github.com/awsdocs/aws-doc-sdk-examples/tree/main/gov2/dynamodb

	//ID採番
	todoId := id.GenerateId()
	todo.ID = todoId

	av, err := attributevalue.MarshalMap(todo)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to marshal item")
	}
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(todoTable),
	}
	//Itemの登録（X-Rayトレース）
	_, err = tr.dynamodbClient.PutItem(ctx, input)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to put item")
	}
	return todo, nil

}
