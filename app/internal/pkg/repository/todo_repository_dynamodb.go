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
	TODO_TABLE_NAME = "TODO_TABLE_NAME"
)

// NewTodoRepositoryForDynamoDB は、TodoRepositoryを作成します。
func NewTodoRepositoryForDynamoDB(dynamoDBTempalte transaction.TransactionalDynamoDBTemplate,
	accessor transaction.TransactionalDynamoDBAccessor,
	logger logging.Logger, config config.Config,
	id id.IDGenerator) TodoRepository {
	// テーブル名の取得
	tableName := tables.DynamoDBTableName(config.Get(TODO_TABLE_NAME, "todo"))
	// テーブル定義の設定
	mytables.Todo{}.InitPK(tableName)

	return &todoRepositoryImplByDynamoDB{
		dynamodbTemplate: dynamoDBTempalte,
		accessor:         accessor,
		logger:           logger,
		config:           config,
		tableName:        tableName,
		id:               id,
	}
}

// todoRepositoryImplByDynamoDB は、TodoRepositoryを実装する構造体です。
type todoRepositoryImplByDynamoDB struct {
	dynamodbTemplate transaction.TransactionalDynamoDBTemplate
	accessor         transaction.TransactionalDynamoDBAccessor
	logger           logging.Logger
	config           config.Config
	tableName        tables.DynamoDBTableName
	id               id.IDGenerator
}

func (tr *todoRepositoryImplByDynamoDB) FindOne(todoId string) (*model.Todo, error) {
	// DynamoDBTemplateを使ったコード
	input := input.PKOnlyQueryInput{
		PrimaryKey: input.PrimaryKey{
			PartitionKey: input.Attribute{
				Name:  tables.GetPrimaryKey(tr.tableName).PartitionKey,
				Value: todoId,
			},
		},
	}
	var todo model.Todo
	// Itemの取得
	err := tr.dynamodbTemplate.FindOneByTableKey(tr.tableName, input, &todo)
	if err != nil {
		if errors.Is(err, mydynamodb.ErrRecordNotFound) {
			// レコード未取得の場合
			return nil, errors.NewBusinessError(message.W_EX_8002, todoId)
		}
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}

	// 従来のDynamoDBAccessorを使ったコード
	// AWS SDK for Go v2 Migration
	// https://docs.aws.amazon.com/ja_jp/code-library/latest/ug/go_2_dynamodb_code_examples.html
	// https://github.com/awsdocs/aws-doc-sdk-examples/tree/main/gov2/dynamodb
	/*
		todo := model.Todo{ID: todoId}
		key, err := todo.GetKey()
		if err != nil {
			return nil, errors.NewSystemError(err, message.E_EX_9001)
		}
		// Itemの取得
		result, err := tr.accessor.GetItemSdk(&dynamodb.GetItemInput{
			TableName: aws.String(tr.tableName),
			Key:       key,
		})
		if err != nil {
			return nil, errors.NewSystemError(err, message.E_EX_9001)
		}
		// レコード未取得の場合
		if len(result.Item) == 0 {
			return nil, errors.NewBusinessError(message.W_EX_8002, todoId)
		}
		err = attributevalue.UnmarshalMap(result.Item, &todo)
		if err != nil {
			return nil, errors.NewSystemError(err, message.E_EX_9001)
		}
	*/
	return &todo, nil
}

func (tr *todoRepositoryImplByDynamoDB) CreateOne(todo *model.Todo) (*model.Todo, error) {
	// ID採番
	todoId, err := tr.id.GenerateUUID()
	if err != nil {
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}
	//todoId := "dummy"
	todo.ID = todoId
	// DynamoDBTemplateを使ったコード
	err = tr.dynamodbTemplate.CreateOne(tr.tableName, todo)
	if err != nil {
		if errors.Is(err, mydynamodb.ErrKeyDuplicaiton) {
			// キーの重複の場合
			// 都度、処理実行時に生成するUUIDが重複するということは、DynamoDB登録後に、何らかの原因でエラーとなり
			// AWS SDKでのリトライによる重複実行が発生している場合なので、Warnログを出力しエラーとしない。
			tr.logger.Warn(message.W_EX_8003, todoId)
			return todo, nil

			// 業務によっては、ErrKeyDuplicaitonが返却された場合に、業務エラーとすることもある。
			//return nil, errors.NewBusinessError(message.W_EX_8003, todoId)
		}
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}
	// 従来のDynamoDBAccessorを使ったコード
	// AWS SDK for Go v2 Migration
	// https://docs.aws.amazon.com/ja_jp/code-library/latest/ug/go_2_dynamodb_code_examples.html
	// https://github.com/awsdocs/aws-doc-sdk-examples/tree/main/gov2/dynamodb
	/*
		av, err := attributevalue.MarshalMap(todo)
		if err != nil {
			return nil, errors.NewSystemError(err, message.E_EX_9001)
		}
		input := &dynamodb.PutItemInput{
			Item:      av,
			TableName: aws.String(tr.tableName),
		}
		// Itemの登録
		_, err = tr.accessor.PutItemSdk(input)
		if err != nil {
			return nil, errors.NewSystemError(err, message.E_EX_9001)
		}
	*/
	return todo, nil

}

// CreateOneTx implements TodoRepository.
func (tr *todoRepositoryImplByDynamoDB) CreateOneTx(todo *model.Todo) (*model.Todo, error) {
	// ID採番
	todoId, err := tr.id.GenerateUUID()
	if err != nil {
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}
	//todoId := "dummy"
	todo.ID = todoId
	// DynamoDBTemplateを使ったコード
	err = tr.dynamodbTemplate.CreateOneWithTransaction(tr.tableName, todo)
	if err != nil {
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}

	// 従来のDynamoDBAccessorを使ったコード
	/*
		av, err := attributevalue.MarshalMap(todo)
		if err != nil {
			return nil, errors.NewSystemError(err, message.E_EX_9001)
		}
		put := &types.Put{
			Item:      av,
			TableName: aws.String(tr.tableName),
		}
		// TransactWriteItemの追加
		input := &types.TransactWriteItem{Put: put}
		err := tr.accessor.AppendTransactWriteItem(input)
		if err != nil {
			return nil, errors.NewSystemError(err, message.E_EX_9001)
		}
	*/
	return todo, nil
}
