/*
dynamodb パッケージは、DynamoDBアクセスに関する機能を提供するパッケージです。
*/
package dynamodb

import (
	"strings"

	"example.com/appbase/pkg/dynamodb/criteria"
	"example.com/appbase/pkg/dynamodb/tables"
	"example.com/appbase/pkg/logging"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/cockroachdb/errors"
)

var (
	ErrRecordNotFound     = errors.New("record not found")
	ErrKeyDuplicaiton     = errors.New("key duplication")
	ErrUpdateWithCondtion = errors.New("update with condition error")
	ErrDeleteWithCondtion = errors.New("delete with condition error")
)

// DynamoDBTemplate は、DynamoDBアクセスを定型化した高次のインタフェースです。
type DynamoDBTemplate interface {
	CreateOne(tableName tables.DynamoDBTableName, inputEntity any) error
	FindOneByPrimaryKey(tableName tables.DynamoDBTableName, input criteria.PkOnlyQueryInput, outEntity any) error
	FindSomeByPrimaryKey(tableName tables.DynamoDBTableName, input criteria.PkOnlyQueryInput, outEntities any) error
	FindSomeByGSI(tableName tables.DynamoDBTableName, input criteria.GsiQueryInput, outEntities any) error
	UpdateOne(tableName tables.DynamoDBTableName, input criteria.UpdateInput) error
	DeleteOne(tableName tables.DynamoDBTableName, input criteria.DeleteInput) error
}

// NewDynamoDBTemplate は、DynamoDBTemplateのインスタンスを生成します。
func NewDynamoDBTemplate(log logging.Logger, dynamodbAccessor DynamoDBAccessor) DynamoDBTemplate {
	return &defaultDynamoDBTemplate{
		log:              log,
		dynamodbAccessor: dynamodbAccessor,
	}
}

//TODO:　DynamoDBTemplateインタフェースの実装

type defaultDynamoDBTemplate struct {
	log              logging.Logger
	dynamodbAccessor DynamoDBAccessor
}

// CreateOne implements DynamoDBTemplate.
func (t *defaultDynamoDBTemplate) CreateOne(tableName tables.DynamoDBTableName, inputEntity any) error {
	item, err := attributevalue.MarshalMap(inputEntity)
	if err != nil {
		return errors.WithStack(err)
	}
	// パーティションキーの重複判定条件
	partitonkeyName := tables.GetPrimaryKey(tableName).PartitionKey
	conditionExpression := aws.String("attribute_not_exists(#partition_key)")
	expressionAttributeNames := map[string]string{
		"#partition_key": partitonkeyName,
	}
	input := &dynamodb.PutItemInput{
		TableName:                aws.String(string(tableName)),
		Item:                     item,
		ConditionExpression:      conditionExpression,
		ExpressionAttributeNames: expressionAttributeNames,
	}
	_, err = t.dynamodbAccessor.PutItemSdk(input)
	if err != nil {
		var condErr *types.ConditionalCheckFailedException
		if errors.As(err, &condErr) {
			return ErrKeyDuplicaiton
		}
		return errors.WithStack(err)
	}
	return nil
}

// FindOneByPrimaryKey implements DynamoDBTemplate.
func (t *defaultDynamoDBTemplate) FindOneByPrimaryKey(tableName tables.DynamoDBTableName, input criteria.PkOnlyQueryInput, outEntity any) error {
	// プライマリキーの検索条件
	keyMap, err := t.createPkAttributeValue(input.PrimarKey)
	if err != nil {
		return errors.WithStack(err)
	}
	// 取得項目
	var projection *string
	if len(input.SelectItems) > 0 {
		projection = aws.String(strings.Join(input.SelectItems, ","))
	}
	// GetItemInput
	getItemInput := &dynamodb.GetItemInput{
		TableName:            aws.String(string(tableName)),
		Key:                  keyMap,
		ProjectionExpression: projection,
		ConsistentRead:       aws.Bool(input.ConsitentRead),
	}
	// GetItemの実行
	getItemOutput, err := t.dynamodbAccessor.GetItemSdk(getItemInput)
	if err != nil {
		return errors.WithStack(err)
	}
	if len(getItemOutput.Item) == 0 {
		return ErrRecordNotFound
	}
	if err := attributevalue.UnmarshalMap(getItemOutput.Item, &outEntity); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// FindSomeByGSI implements DynamoDBTemplate.
func (t *defaultDynamoDBTemplate) FindSomeByGSI(tableName tables.DynamoDBTableName, input criteria.GsiQueryInput, outEntities any) error {
	panic("unimplemented")
}

// FindSomeByPrimaryKey implements DynamoDBTemplate.
func (t *defaultDynamoDBTemplate) FindSomeByPrimaryKey(tableName tables.DynamoDBTableName, input criteria.PkOnlyQueryInput, outEntities any) error {
	panic("unimplemented")
}

// UpdateOne implements DynamoDBTemplate.
func (t *defaultDynamoDBTemplate) UpdateOne(tableName tables.DynamoDBTableName, input criteria.UpdateInput) error {
	panic("unimplemented")
}

// DeleteOne implements DynamoDBTemplate.
func (t *defaultDynamoDBTemplate) DeleteOne(tableName tables.DynamoDBTableName, input criteria.DeleteInput) error {
	panic("unimplemented")
}

//TODO: transactionパッケージでも使えるように公開しないとダメかも

func (t *defaultDynamoDBTemplate) typeSwitch(keyValue criteria.KeyValue) (types.AttributeValue, error) {
	t.log.Debug("typeSwitch:%v", keyValue.Value)
	switch keyValue.Value.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		//TODO: ほんとに数値をそのままstringキャストでよい？
		return &types.AttributeValueMemberN{Value: keyValue.Value.(string)}, nil
	case string:
		return &types.AttributeValueMemberS{Value: keyValue.Value.(string)}, nil
	case bool:
		return &types.AttributeValueMemberBOOL{Value: keyValue.Value.(bool)}, nil
	case []byte:
		return &types.AttributeValueMemberB{Value: keyValue.Value.([]byte)}, nil
	default:
		return nil, errors.New("type not supported")
	}
}

// プライマリキーの完全一致による条件のAttributeValueのマップを生成します。
func (t *defaultDynamoDBTemplate) createPkAttributeValue(primaryKey criteria.KeyPair) (map[string]types.AttributeValue, error) {
	keymap := map[string]types.AttributeValue{}
	// パーティションキー
	partitionKey := primaryKey.PartitionKey
	pk, err := t.typeSwitch(partitionKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	keymap[partitionKey.Key] = pk

	// ソートキー
	sortKey := primaryKey.SortKey
	if sortKey != nil {
		sk, err := t.typeSwitch(*sortKey)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		keymap[sortKey.Key] = sk
	}
	return keymap, nil
}
