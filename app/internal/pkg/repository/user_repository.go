package repository

import (
	"context"
	"os"

	"app/internal/pkg/entity"

	"example.com/appbase/pkg/apcontext"
	"example.com/appbase/pkg/id"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/pkg/errors"
)

var (
	userTable = os.Getenv("USERS_TABLE_NAME")
)

type UserRepository interface {
	GetUser(userId string) (*entity.User, error)
	PutUser(user *entity.User) (*entity.User, error)
}

func NewUserRepository() UserRepository {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region)}),
	)
	dynamo := dynamodb.New(sess)
	xray.AWS(dynamo.Client)
	return &UserRepositoryImpl{instance: dynamo}
}

type UserRepositoryImpl struct {
	instance *dynamodb.DynamoDB
}

func (ur *UserRepositoryImpl) GetUser(userId string) (*entity.User, error) {
	return ur.doGetUser(userId, apcontext.Context)
}

func (ur *UserRepositoryImpl) doGetUser(userId string, ctx context.Context) (*entity.User, error) {
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

func (ur *UserRepositoryImpl) PutUser(user *entity.User) (*entity.User, error) {
	return ur.doPutUser(user, apcontext.Context)
}

func (ur *UserRepositoryImpl) doPutUser(user *entity.User, ctx context.Context) (*entity.User, error) {
	//ID採番
	userId := id.GenerateId()
	user.ID = userId

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
