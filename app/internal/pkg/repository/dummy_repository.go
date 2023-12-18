// repositoryのパッケージ
package repository

import (
	"app/internal/pkg/entity"
	"app/internal/pkg/message"

	"example.com/appbase/pkg/config"
	mydynamodb "example.com/appbase/pkg/dynamodb"
	"example.com/appbase/pkg/errors"
	"example.com/appbase/pkg/id"
	"example.com/appbase/pkg/logging"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// DummyRepository は、ダミーテーブルを管理するRepositoryインタフェースです。
type DummyRepository interface {
	// FindOne は、dummyIdが一致する値を取得します。
	FindOne(dummyId string) (*entity.Dummy, error)
	// CreateOneTx は、指定された値をトランザクションを使って登録します。
	CreateOneTx(dummy *entity.Dummy) (*entity.Dummy, error)
}

// NewTodoRepositoryForDynamoDB は、TodoRepositoryを作成します。
func NewDummyRepository(accessor mydynamodb.DynamoDBAccessor, log logging.Logger, config config.Config) DummyRepository {
	return &dummyRepositoryImpl{
		accessor: accessor,
		log:      log,
		config:   config,
	}
}

// dummyRepositoryImpl は、DummyRepositoryを実装する構造体です。
type dummyRepositoryImpl struct {
	accessor mydynamodb.DynamoDBAccessor
	log      logging.Logger
	config   config.Config
}

// FindOne implements DummyRepository.
func (r *dummyRepositoryImpl) FindOne(dummyId string) (*entity.Dummy, error) {
	// Itemの取得
	dummy := entity.Dummy{ID: dummyId}
	key, err := dummy.GetKey()
	if err != nil {
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}
	result, err := r.accessor.GetItemSdk(&dynamodb.GetItemInput{
		TableName: aws.String("dummy"),
		Key:       key,
	})
	if err != nil {
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}
	err = attributevalue.UnmarshalMap(result.Item, &dummy)
	if err != nil {
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}
	return &dummy, nil
}

// CreateOneTx implements DummyRepository.
func (r *dummyRepositoryImpl) CreateOneTx(dummy *entity.Dummy) (*entity.Dummy, error) {
	// ID採番
	dummyId := id.GenerateId()
	dummy.ID = dummyId
	av, err := attributevalue.MarshalMap(dummy)
	if err != nil {
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}
	put := &types.Put{
		Item:      av,
		TableName: aws.String("dummy"),
	}
	// TransactWriteItemの追加
	input := &types.TransactWriteItem{Put: put}
	r.accessor.AppendTransactWriteItem(input)

	return dummy, nil
}
