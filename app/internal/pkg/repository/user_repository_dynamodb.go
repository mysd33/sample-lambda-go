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

func NewForDynamoDB() (UserRepository, error) {
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
	return &UserRepositoryImplByDynamoDB{instance: dynamo}, nil
}

type UserRepositoryImplByDynamoDB struct {
	instance *dynamodb.Client
}

func (ur *UserRepositoryImplByDynamoDB) GetUser(userId string) (*entity.User, error) {
	return ur.doGetUser(userId, apcontext.Context)
}

func (ur *UserRepositoryImplByDynamoDB) doGetUser(userId string, ctx context.Context) (*entity.User, error) {
	// AWS SDK for Go v2 Migration
	// https://docs.aws.amazon.com/ja_jp/code-library/latest/ug/go_2_dynamodb_code_examples.html
	// https://github.com/awsdocs/aws-doc-sdk-examples/tree/main/gov2/dynamodb
	//Itemの取得（X-Rayトレース）
	user := entity.User{ID: userId}
	key, err := user.GetKey()
	if err != nil {
		return nil, errors.Wrapf(err, "fail to get key")
	}
	result, err := ur.instance.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(userTable),
		Key:       key,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get item")
	}
	if result.Item == nil {
		return nil, nil
	}

	err = attributevalue.UnmarshalMap(result.Item, &user)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to marshal item")
	}
	return &user, nil
}

func (ur *UserRepositoryImplByDynamoDB) PutUser(user *entity.User) (*entity.User, error) {
	return ur.doPutUser(user, apcontext.Context)
}

func (ur *UserRepositoryImplByDynamoDB) doPutUser(user *entity.User, ctx context.Context) (*entity.User, error) {
	//ID採番
	userId := id.GenerateId()
	user.ID = userId

	av, err := attributevalue.MarshalMap(user)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to marshal item")
	}
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(userTable),
	}
	//Itemの登録（X-Rayトレース）
	_, err = ur.instance.PutItem(ctx, input)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to put item")
	}
	return user, nil
}
