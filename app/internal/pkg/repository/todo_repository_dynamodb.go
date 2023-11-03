// repositoryのパッケージ
package repository

import (
	"app/internal/pkg/entity"
	"app/internal/pkg/message"

	mydynamodb "example.com/appbase/pkg/dynamodb"
	"example.com/appbase/pkg/errors"
	"example.com/appbase/pkg/id"
	"example.com/appbase/pkg/logging"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// NewTodoRepositoryForDynamoDB は、TodoRepositoryを作成します。
func NewTodoRepositoryForDynamoDB(accessor mydynamodb.DynamoDBAccessor, log logging.Logger) TodoRepository {
	return &todoRepositoryImplByDynamoDB{
		accessor: accessor,
		log:      log,
	}
}

// todoRepositoryImplByDynamoDB は、TodoRepositoryを実装する構造体です。
type todoRepositoryImplByDynamoDB struct {
	accessor mydynamodb.DynamoDBAccessor
	log      logging.Logger
}

func (tr *todoRepositoryImplByDynamoDB) GetTodo(todoId string) (*entity.Todo, error) {
	// AWS SDK for Go v2 Migration
	// https://docs.aws.amazon.com/ja_jp/code-library/latest/ug/go_2_dynamodb_code_examples.html
	// https://github.com/awsdocs/aws-doc-sdk-examples/tree/main/gov2/dynamodb
	//Itemの取得
	todo := entity.Todo{ID: todoId}
	key, err := todo.GetKey()
	if err != nil {
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}
	result, err := tr.accessor.GetItemSdk(&dynamodb.GetItemInput{
		TableName: aws.String(todoTable),
		Key:       key,
	})
	if err != nil {
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}
	err = attributevalue.UnmarshalMap(result.Item, &todo)
	if err != nil {
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}
	return &todo, nil
}

func (tr *todoRepositoryImplByDynamoDB) PutTodo(todo *entity.Todo) (*entity.Todo, error) {
	// AWS SDK for Go v2 Migration
	// https://docs.aws.amazon.com/ja_jp/code-library/latest/ug/go_2_dynamodb_code_examples.html
	// https://github.com/awsdocs/aws-doc-sdk-examples/tree/main/gov2/dynamodb

	//ID採番
	todoId := id.GenerateId()
	todo.ID = todoId

	av, err := attributevalue.MarshalMap(todo)
	if err != nil {
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(todoTable),
	}
	// Itemの登録
	_, err = tr.accessor.PutItemSdk(input)
	if err != nil {
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}
	return todo, nil

}
