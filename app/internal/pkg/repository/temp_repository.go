// repositoryのパッケージ
package repository

import (
	"app/internal/pkg/message"
	"app/internal/pkg/model"
	mytables "app/internal/pkg/repository/tables"

	"example.com/appbase/pkg/config"
	mydynamodb "example.com/appbase/pkg/dynamodb"
	"example.com/appbase/pkg/dynamodb/input"
	"example.com/appbase/pkg/dynamodb/tables"
	"example.com/appbase/pkg/errors"
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
	FindOne(id string) (*model.Temp, error)
	// CreateOneTx は、指定された値をトランザクションを使って登録します。
	CreateOneTx(temp *model.Temp) (*model.Temp, error)
	// CreateOne は、指定された値を登録します。
	CreateOne(temp *model.Temp) (*model.Temp, error)
}

// NewTempRepository は、TempRepositoryを生成します。
func NewTempRepository(dynamoDBTempalte transaction.TransactionalDynamoDBTemplate,
	accessor transaction.TransactionalDynamoDBAccessor,
	logger logging.Logger, config config.Config,
	id id.IDGenerator) TempRepository {
	// テーブル名の取得
	tableName := tables.DynamoDBTableName(config.Get(TEMP_TABLE_NAME, "temp"))
	// テーブル定義の設定
	mytables.Temp{}.InitPK(tableName)

	return &tempRepositoryImpl{
		dynamodbTemplate: dynamoDBTempalte,
		accessor:         accessor,
		logger:           logger,
		config:           config,
		tableName:        tableName,
		id:               id,
	}
}

// tempRepositoryImpl は、TempRepositoryを実装する構造体です。
type tempRepositoryImpl struct {
	dynamodbTemplate transaction.TransactionalDynamoDBTemplate
	accessor         transaction.TransactionalDynamoDBAccessor
	logger           logging.Logger
	config           config.Config
	tableName        tables.DynamoDBTableName
	id               id.IDGenerator
}

// FindOne implements TempRepository.
func (r *tempRepositoryImpl) FindOne(id string) (*model.Temp, error) {
	// DynamoDBTemplateを使ったコード
	input := input.PKOnlyQueryInput{
		PrimaryKey: input.PrimaryKey{
			PartitionKey: input.Attribute{
				Name:  tables.GetPrimaryKey(r.tableName).PartitionKey,
				Value: id,
			},
		},
		//ConsitentRead: true,
	}
	var temp model.Temp
	// Itemの取得
	err := r.dynamodbTemplate.FindOneByTableKey(r.tableName, input, &temp)
	if err != nil {
		if errors.Is(err, mydynamodb.ErrRecordNotFound) {
			// レコード未取得の場合
			return nil, errors.NewBusinessError(message.W_EX_8006, id)
		}
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}

	// 従来のDynamoDBAccessorを使ったコード
	// Itemの取得
	/*
		temp := model.Temp{ID: id}
		key, err := temp.GetKey()
		if err != nil {
			return nil, errors.NewSystemError(err, message.E_EX_9001)
		}
		result, err := r.accessor.GetItemSdk(&dynamodb.GetItemInput{
			TableName: aws.String("temp"),
			Key:       key,
		})
		if err != nil {
			return nil, errors.NewSystemError(err, message.E_EX_9001)
		}
		err = attributevalue.UnmarshalMap(result.Item, &temp)
		if err != nil {
			return nil, errors.NewSystemError(err, message.E_EX_9001)
		}*/
	return &temp, nil
}

// CreateOneTx implements TempRepository.
func (r *tempRepositoryImpl) CreateOneTx(temp *model.Temp) (*model.Temp, error) {
	// ID採番
	id, err := r.id.GenerateUUID()
	if err != nil {
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}
	temp.ID = id
	r.logger.Debug("CreateOneTx Table name: %s", r.tableName)
	r.logger.Debug("CreateOneTx Temp id: %s", id)

	// DynamoDBTemplateを使ったコード
	err = r.dynamodbTemplate.CreateOneWithTransaction(r.tableName, temp)
	if err != nil {
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}

	// 従来のDynamoDBAccessorを使ったコード
	/*
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
		err := r.accessor.AppendTransactWriteItem(input)
		if err != nil {
			return nil, errors.NewSystemError(err, message.E_EX_9001)
		}
	*/
	return temp, nil
}

// CreateOne implements TempRepository.
func (r *tempRepositoryImpl) CreateOne(temp *model.Temp) (*model.Temp, error) {
	// ID採番
	id, err := r.id.GenerateUUID()
	if err != nil {
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}
	temp.ID = id
	r.logger.Debug("CreateOne Table name: %s", r.tableName)
	r.logger.Debug("CreateOne Temp id: %s", id)

	// DynamoDBTemplateを使ったコード
	err = r.dynamodbTemplate.CreateOne(r.tableName, temp)
	if err != nil {
		if errors.Is(err, mydynamodb.ErrKeyDuplicaiton) {
			return nil, errors.NewBusinessError(message.W_EX_8007, id)
		}
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}
	return temp, nil
}
