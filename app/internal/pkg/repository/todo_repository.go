package repository

import (
	"app/internal/pkg/entity"
	"context"

	"example.com/appbase/pkg/apcontext"
	"example.com/appbase/pkg/id"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/pkg/errors"
)

type TodoRepository interface {
	GetTodo(todoId string) (*entity.Todo, error)
	PutTodo(todo *entity.Todo) (*entity.Todo, error)
}

func NewTodoRepository() TodoRepository {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region)}),
	)
	dynamo := dynamodb.New(sess)
	xray.AWS(dynamo.Client)
	return &TodoRepositoryImpl{instance: dynamo}
}

type TodoRepositoryImpl struct {
	instance *dynamodb.DynamoDB
}

func (tr *TodoRepositoryImpl) GetTodo(todoId string) (*entity.Todo, error) {
	return tr.doGetTodo(todoId, apcontext.Context)
}

func (tr *TodoRepositoryImpl) doGetTodo(todoId string, ctx context.Context) (*entity.Todo, error) {
	//Itemの取得（X-Rayトレース）
	result, err := tr.instance.GetItemWithContext(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(todoTable),
		Key: map[string]*dynamodb.AttributeValue{
			"todo_id": {
				S: aws.String(todoId),
			},
		},
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get item")
	}
	if result.Item == nil {
		return nil, nil
	}
	todo := entity.Todo{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &todo)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to marshal item")
	}
	return &todo, nil
}

func (tr *TodoRepositoryImpl) PutTodo(todo *entity.Todo) (*entity.Todo, error) {
	return tr.doPutTodo(todo, apcontext.Context)
}

func (tr *TodoRepositoryImpl) doPutTodo(todo *entity.Todo, ctx context.Context) (*entity.Todo, error) {
	//ID採番
	todoId := id.GenerateId()
	todo.ID = todoId

	av, err := dynamodbattribute.MarshalMap(todo)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to marshal item")
	}
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(todoTable),
	}
	//Itemの登録（X-Rayトレース）
	_, err = tr.instance.PutItemWithContext(ctx, input)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to put item")
	}
	return todo, nil

}
