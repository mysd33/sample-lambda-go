// repositoryのパッケージ
package repository

import (
	"app/internal/pkg/entity"
	"app/internal/pkg/message"

	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/errors"
	"example.com/appbase/pkg/id"
	"example.com/appbase/pkg/logging"
	"example.com/appbase/pkg/transaction"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// TempRepository は、一時テーブルを管理するRepositoryインタフェースです。
type TempRepository interface {
	// FindOne は、idが一致する値を取得します。
	FindOne(id string) (*entity.Temp, error)
	// CreateOneTx は、指定された値をトランザクションを使って登録します。
	CreateOneTx(temp *entity.Temp) (*entity.Temp, error)
}

// NewTempRepository は、TempRepositoryを生成します。
func NewTempRepository(accessor transaction.TransactionalDynamoDBAccessor, log logging.Logger, config config.Config) TempRepository {
	return &tempRepositoryImpl{
		accessor: accessor,
		log:      log,
		config:   config,
	}
}

// tempRepositoryImpl は、TempRepositoryを実装する構造体です。
type tempRepositoryImpl struct {
	accessor transaction.TransactionalDynamoDBAccessor
	log      logging.Logger
	config   config.Config
}

// FindOne implements TempRepository.
func (r *tempRepositoryImpl) FindOne(id string) (*entity.Temp, error) {
	// Itemの取得
	temp := entity.Temp{ID: id}
	key, err := temp.GetKey()
	if err != nil {
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}
	result, err := r.accessor.GetItemSdk(&dynamodb.GetItemInput{
		//TODO: テーブル名を設定外だし
		TableName: aws.String("temp"),
		Key:       key,
	})
	if err != nil {
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}
	err = attributevalue.UnmarshalMap(result.Item, &temp)
	if err != nil {
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}
	return &temp, nil
}

// CreateOneTx implements TempRepository.
func (r *tempRepositoryImpl) CreateOneTx(temp *entity.Temp) (*entity.Temp, error) {
	// ID採番
	id := id.GenerateId()
	temp.ID = id
	av, err := attributevalue.MarshalMap(temp)
	if err != nil {
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}
	put := &types.Put{
		Item:      av,
		TableName: aws.String("temp"),
	}
	// TransactWriteItemの追加
	input := &types.TransactWriteItem{Put: put}
	r.accessor.AppendTransactWriteItem(input)

	return temp, nil
}
