// repositoryのパッケージ
package repository

import (
	"app/internal/pkg/entity"
	"app/internal/pkg/message"
	mytables "app/internal/pkg/repository/tables"
	"errors"

	"example.com/appbase/pkg/config"
	mydynamodb "example.com/appbase/pkg/dynamodb"
	"example.com/appbase/pkg/dynamodb/input"
	"example.com/appbase/pkg/dynamodb/tables"
	myerrors "example.com/appbase/pkg/errors"
	"example.com/appbase/pkg/id"
	"example.com/appbase/pkg/logging"
	"example.com/appbase/pkg/transaction"
)

const (
	TEMP_TABLE_NAME = "TEMP_TABLE_NAME"
)

// TempRepository は、一時テーブルを管理するRepositoryインタフェースです。
type TempRepository interface {
	// FindOne は、idが一致する値を取得します。
	FindOne(id string) (*entity.Temp, error)
	// CreateOneTx は、指定された値をトランザクションを使って登録します。
	CreateOneTx(temp *entity.Temp) (*entity.Temp, error)
}

// NewTempRepository は、TempRepositoryを生成します。
func NewTempRepository(dynamoDBTempalte transaction.TransactionalDynamoDBTemplate,
	accessor transaction.TransactionalDynamoDBAccessor,
	log logging.Logger, config config.Config) TempRepository {
	// テーブル名の取得
	tableName := tables.DynamoDBTableName(config.Get(TEMP_TABLE_NAME))
	// テーブル定義の設定
	mytables.Temp{}.InitPK(tableName)
	// プライマリキーの設定
	primaryKey := tables.GetPrimaryKey(tableName)

	return &tempRepositoryImpl{
		dynamodbTemplate: dynamoDBTempalte,
		accessor:         accessor,
		log:              log,
		config:           config,
		tableName:        tableName,
		primaryKey:       primaryKey,
	}
}

// tempRepositoryImpl は、TempRepositoryを実装する構造体です。
type tempRepositoryImpl struct {
	dynamodbTemplate transaction.TransactionalDynamoDBTemplate
	accessor         transaction.TransactionalDynamoDBAccessor
	log              logging.Logger
	config           config.Config
	tableName        tables.DynamoDBTableName
	primaryKey       *tables.PKKeyPair
}

// FindOne implements TempRepository.
func (r *tempRepositoryImpl) FindOne(id string) (*entity.Temp, error) {
	// DynamoDBTemplateを使ったコード
	input := input.PKOnlyQueryInput{
		PrimaryKey: input.PrimaryKey{
			PartitionKey: input.Attribute{
				Name:  r.primaryKey.PartitionKey,
				Value: id,
			},
		},
	}
	var temp entity.Temp
	// Itemの取得
	err := r.dynamodbTemplate.FindOneByTableKey(r.tableName, input, &temp)
	if err != nil {
		if errors.Is(err, mydynamodb.ErrRecordNotFound) {
			// レコード未取得の場合
			return nil, myerrors.NewBusinessError(message.W_EX_8002, id)
		}
		return nil, myerrors.NewSystemError(err, message.E_EX_9001)
	}

	// 従来のDynamoDBAccessorを使ったコード
	// Itemの取得
	/*
		temp := entity.Temp{ID: id}
		key, err := temp.GetKey()
		if err != nil {
			return nil, myerrors.NewSystemError(err, message.E_EX_9001)
		}
		result, err := r.accessor.GetItemSdk(&dynamodb.GetItemInput{
			TableName: aws.String("temp"),
			Key:       key,
		})
		if err != nil {
			return nil, myerrors.NewSystemError(err, message.E_EX_9001)
		}
		err = attributevalue.UnmarshalMap(result.Item, &temp)
		if err != nil {
			return nil, myerrors.NewSystemError(err, message.E_EX_9001)
		}*/
	return &temp, nil
}

// CreateOneTx implements TempRepository.
func (r *tempRepositoryImpl) CreateOneTx(temp *entity.Temp) (*entity.Temp, error) {
	// ID採番
	id := id.GenerateId()
	temp.ID = id
	r.log.Debug("CreateOneTx Table name: %s", r.tableName)
	r.log.Debug("CreateOneTx Temp id: %s", id)

	// DynamoDBTemplateを使ったコード
	err := r.dynamodbTemplate.CreateOneWithTransaction(r.tableName, temp)
	if err != nil {
		return nil, myerrors.NewSystemError(err, message.E_EX_9001)
	}

	// 従来のDynamoDBAccessorを使ったコード
	/*
		av, err := attributevalue.MarshalMap(temp)
		if err != nil {
			return nil, myerrors.NewSystemError(err, message.E_EX_9001)
		}
		put := &types.Put{
			Item:      av,
			TableName: aws.String("temp"),
		}
		// TransactWriteItemの追加
		input := &types.TransactWriteItem{Put: put}
		r.accessor.AppendTransactWriteItem(input)
	*/
	return temp, nil
}
