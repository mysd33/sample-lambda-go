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
	region    = os.Getenv("REGION")
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
	return &UserRepositoryImpl{Instance: dynamo}
}

type UserRepositoryImpl struct {
	Instance *dynamodb.DynamoDB
	Context  context.Context
}

func (d *UserRepositoryImpl) GetUser(userId string) (*entity.User, error) {
	return d.doGetUser(userId, apcontext.Context)
}

func (d *UserRepositoryImpl) doGetUser(userId string, ctx context.Context) (*entity.User, error) {
	//Itemの取得（X-Rayトレース）
	result, err := d.Instance.GetItemWithContext(ctx, &dynamodb.GetItemInput{
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

func (d *UserRepositoryImpl) PutUser(user *entity.User) (*entity.User, error) {
	return d.doPutUser(user, apcontext.Context)
}

func (d *UserRepositoryImpl) doPutUser(user *entity.User, ctx context.Context) (*entity.User, error) {
	//ID採番
	userId := id.GenerateId()
	user.ID = userId

	av, err := dynamodbattribute.MarshalMap(user)
	if err != nil {
		return nil, err
	}
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(userTable),
	}
	//Itemの登録（X-Rayトレース）
	_, err = d.Instance.PutItemWithContext(ctx, input)
	if err != nil {
		return nil, err
	}
	return user, nil
}
