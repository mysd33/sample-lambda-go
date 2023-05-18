package repository

import (
	"app/internal/pkg/entity"
	"context"

	"example.com/appbase/pkg/apcontext"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/pkg/errors"
)

func NewUserRepositoryForDynamoDB() UserRepository {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region)}),
	)
	dynamo := dynamodb.New(sess)
	xray.AWS(dynamo.Client)
	return &UserRepositoryImplByDynamoDB{instance: dynamo}
}

type UserRepositoryImplByDynamoDB struct {
	instance *dynamodb.DynamoDB
}

func (ur *UserRepositoryImplByDynamoDB) GetUser(userId string) (*entity.User, error) {
	return ur.doGetUser(userId, apcontext.Context)
}

func (ur *UserRepositoryImplByDynamoDB) doGetUser(userId string, ctx context.Context) (*entity.User, error) {
	//Itemの取得（X-Rayトレース）
	result, err := ur.instance.GetItemWithContext(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(userTable),
		Key: map[string]*dynamodb.AttributeValue{
			"user_id": {
				S: aws.String(userId),
			},
		},
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get item")
	}
	if result.Item == nil {
		return nil, nil
	}
	user := entity.User{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &user)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to marshal item")
	}
	return &user, nil
}

func (ur *UserRepositoryImplByDynamoDB) PutUser(user *entity.User) (*entity.User, error) {
	return ur.doPutUser(user, apcontext.Context)
}

func (ur *UserRepositoryImplByDynamoDB) doPutUser(user *entity.User, ctx context.Context) (*entity.User, error) {
	av, err := dynamodbattribute.MarshalMap(user)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to marshal item")
	}
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(userTable),
	}
	//Itemの登録（X-Rayトレース）
	_, err = ur.instance.PutItemWithContext(ctx, input)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to put item")
	}
	return user, nil
}
