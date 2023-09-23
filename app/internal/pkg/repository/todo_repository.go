package repository

import (
	"app/internal/pkg/entity"
	"context"

	"example.com/appbase/pkg/apcontext"
	"example.com/appbase/pkg/id"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-xray-sdk-go/instrumentation/awsv2"
	"github.com/pkg/errors"
)

type TodoRepository interface {
	GetTodo(todoId string) (*entity.Todo, error)
	PutTodo(todo *entity.Todo) (*entity.Todo, error)
}

func NewTodoRepository() (TodoRepository, error) {
	// AWS SDK for Go v2 Migration
	// https://github.com/aws/aws-sdk-go-v2
	// https://aws.github.io/aws-sdk-go-v2/docs/migrating/
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region))
	if err != nil {
		return nil, err
	}
	// Instrumenting AWS SDK v2
	// https://github.com/aws/aws-xray-sdk-go
	awsv2.AWSV2Instrumentor(&cfg.APIOptions)
	dynamo := dynamodb.NewFromConfig(cfg)
	return &TodoRepositoryImpl{instance: dynamo}, nil
}

type TodoRepositoryImpl struct {
	instance *dynamodb.Client
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
	result, err := tr.instance.GetItem(ctx, &dynamodb.GetItemInput{
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
	_, err = tr.instance.PutItem(ctx, input)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to put item")
	}
	return todo, nil

}
