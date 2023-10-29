// repositoryのパッケージ
package repository

import (
	"app/internal/pkg/code"
	"app/internal/pkg/entity"

	mydynamodb "example.com/appbase/pkg/dynamodb"
	"example.com/appbase/pkg/errors"
	"example.com/appbase/pkg/id"
	"example.com/appbase/pkg/logging"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// NewUserRepositoryForDynamoDB は、DynamoDB保存のためのUserRepository実装を作成します。
func NewUserRepositoryForDynamoDB(accessor mydynamodb.DynamoDBAccessor, log logging.Logger) UserRepository {
	return &UserRepositoryImplByDynamoDB{
		accessor: accessor,
		log:      log,
	}
}

// UserRepositoryImplByDynamoDB は、DynamoDB保存のためのUserRepository実装です。
type UserRepositoryImplByDynamoDB struct {
	accessor mydynamodb.DynamoDBAccessor
	log      logging.Logger
}

func (ur *UserRepositoryImplByDynamoDB) GetUser(userId string) (*entity.User, error) {
	// AWS SDK for Go v2 Migration
	// https://docs.aws.amazon.com/ja_jp/code-library/latest/ug/go_2_dynamodb_code_examples.html
	// https://github.com/awsdocs/aws-doc-sdk-examples/tree/main/gov2/dynamodb
	//Itemの取得（X-Rayトレース）
	user := entity.User{ID: userId}
	key, err := user.GetKey()
	if err != nil {
		// return nil, errors.Wrapf(err, "fail to get key")
		return nil, errors.NewSystemError(err, code.E_EX_9001)
	}
	result, err := ur.accessor.GetItemSdk(&dynamodb.GetItemInput{
		TableName: aws.String(userTable),
		Key:       key,
	})
	if err != nil {
		// return nil, errors.Wrapf(err, "failed to get item")
		return nil, errors.NewSystemError(err, code.E_EX_9001)
	}
	if result.Item == nil {
		return nil, nil
	}

	err = attributevalue.UnmarshalMap(result.Item, &user)
	if err != nil {
		// return nil, errors.Wrapf(err, "failed to marshal item")
		return nil, errors.NewSystemError(err, code.E_EX_9001)
	}
	return &user, nil
}

func (ur *UserRepositoryImplByDynamoDB) PutUser(user *entity.User) (*entity.User, error) {
	// ID採番
	userId := id.GenerateId()
	user.ID = userId

	av, err := attributevalue.MarshalMap(user)
	if err != nil {
		// return nil, errors.Wrapf(err, "failed to marshal item")
		return nil, errors.NewSystemError(err, code.E_EX_9001)
	}
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(userTable),
	}
	// Itemの登録
	_, err = ur.accessor.PutItemSdk(input)
	if err != nil {
		// return nil, errors.Wrapf(err, "failed to put item")
		return nil, errors.NewSystemError(err, code.E_EX_9001)
	}
	return user, nil
}
