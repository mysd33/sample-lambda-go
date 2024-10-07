/*
dynamodb パッケージは、DynamoDBアクセスに関する機能を提供するパッケージです。
*/
package dynamodb

import (
	"context"
	"strings"

	"example.com/appbase/pkg/apcontext"
	"example.com/appbase/pkg/dynamodb/input"
	"example.com/appbase/pkg/dynamodb/tables"
	"example.com/appbase/pkg/logging"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/cockroachdb/errors"
)

var (
	ErrRecordNotFound     = errors.New("対象レコードなし")
	ErrKeyDuplicaiton     = errors.New("プライマリキー重複エラー")
	ErrUpdateWithCondtion = errors.New("条件付き更新エラー")
	ErrDeleteWithCondtion = errors.New("条件付き削除エラー")
)

// DynamoDBTemplate は、DynamoDBアクセスを定型化した高次のインタフェースです。
type DynamoDBTemplate interface {
	// CreateOne は、DynamoDBに項目を1件登録します。
	CreateOne(tableName tables.DynamoDBTableName, inputEntity any, optFns ...func(*dynamodb.Options)) error
	// CreateOneWithContext は、goroutine向けに渡されたContextを利用して、DynamoDBに項目を1件登録します。
	CreateOneWithContext(ctx context.Context, tableName tables.DynamoDBTableName, inputEntity any, optFns ...func(*dynamodb.Options)) error
	// FindOneByTableKey は、ベーステーブルのプライマリキーの完全一致でDynamoDBから1件の項目を取得します。
	FindOneByTableKey(tableName tables.DynamoDBTableName, input input.PKOnlyQueryInput, outEntity any, optFns ...func(*dynamodb.Options)) error
	// FindOneByTableKeyWithContext は、goroutine向けに渡されたContextを利用して、ベーステーブルのプライマリキーの完全一致でDynamoDBから1件の項目を取得します。
	FindOneByTableKeyWithContext(ctx context.Context, tableName tables.DynamoDBTableName, input input.PKOnlyQueryInput, outEntity any, optFns ...func(*dynamodb.Options)) error
	// FindSomeByTableKey は、ベーステーブルのプライマリキーによる条件でDynamoDBから複数件の項目を取得します。
	FindSomeByTableKey(tableName tables.DynamoDBTableName, input input.PKQueryInput, outEntities any, optFns ...func(*dynamodb.Options)) error
	// FindSomeByTableKeyWithContext は、goroutine向けに渡されたContextを利用して、ベーステーブルのプライマリキーによる条件でDynamoDBから複数件の項目を取得します。
	FindSomeByTableKeyWithContext(ctx context.Context, tableName tables.DynamoDBTableName, input input.PKQueryInput, outEntities any, optFns ...func(*dynamodb.Options)) error
	// FindSomeByGSIKey は、GSIのプライマリキーによる条件でDynamoDBから項目を複数件取得します。
	FindSomeByGSIKey(tableName tables.DynamoDBTableName, input input.GsiQueryInput, outEntities any, optFns ...func(*dynamodb.Options)) error
	// FindSomeByGSIKeyWithContext は、goroutine向けに渡されたContextを利用して、GSIのプライマリキーによる条件でDynamoDBから項目を複数件取得します。
	FindSomeByGSIKeyWithContext(ctx context.Context, tableName tables.DynamoDBTableName, input input.GsiQueryInput, outEntities any, optFns ...func(*dynamodb.Options)) error
	// UpdateOne は、DynamoDBの項目を更新します。
	UpdateOne(tableName tables.DynamoDBTableName, input input.UpdateInput, optFns ...func(*dynamodb.Options)) error
	// UpdateOneWithContext は、goroutine向けに渡されたContextを利用して、DynamoDBの項目を更新します。
	UpdateOneWithContext(ctx context.Context, tableName tables.DynamoDBTableName, input input.UpdateInput, optFns ...func(*dynamodb.Options)) error
	// DeleteOne は、DynamoDBの項目を削除します。
	DeleteOne(tableName tables.DynamoDBTableName, input input.DeleteInput, optFns ...func(*dynamodb.Options)) error
	// DeleteOneWithContext は、goroutine向けに渡されたContextを利用して、DynamoDBの項目を削除します。
	DeleteOneWithContext(ctx context.Context, tableName tables.DynamoDBTableName, input input.DeleteInput, optFns ...func(*dynamodb.Options)) error
}

// NewDynamoDBTemplate は、DynamoDBTemplateのインスタンスを生成します。
func NewDynamoDBTemplate(logger logging.Logger, dynamodbAccessor DynamoDBAccessor) DynamoDBTemplate {
	return &defaultDynamoDBTemplate{
		logger:           logger,
		dynamodbAccessor: dynamodbAccessor,
	}
}

// defaultDynamoDBTemplate は、DynamoDBTemplateを実装する構造体です。
type defaultDynamoDBTemplate struct {
	logger           logging.Logger
	dynamodbAccessor DynamoDBAccessor
}

// CreateOne implements DynamoDBTemplate.
func (t *defaultDynamoDBTemplate) CreateOne(tableName tables.DynamoDBTableName, inputEntity any, optFns ...func(*dynamodb.Options)) error {
	return t.CreateOneWithContext(apcontext.Context, tableName, inputEntity, optFns...)
}

// CreateOneWithContext implements DynamoDBTemplate.
func (t *defaultDynamoDBTemplate) CreateOneWithContext(ctx context.Context, tableName tables.DynamoDBTableName, inputEntity any, optFns ...func(*dynamodb.Options)) error {
	attributes, err := attributevalue.MarshalMap(inputEntity)
	if err != nil {
		return errors.Wrap(err, "CreateOneで構造体をAttributeValueのMap変換時にエラー")
	}
	// パーティションキーの重複判定条件
	partitonkeyName := tables.GetPrimaryKey(tableName).PartitionKey
	conditionExpression := aws.String("attribute_not_exists(#partition_key)")
	expressionAttributeNames := map[string]string{
		"#partition_key": partitonkeyName,
	}
	item := &dynamodb.PutItemInput{
		TableName:                aws.String(string(tableName)),
		Item:                     attributes,
		ConditionExpression:      conditionExpression,
		ExpressionAttributeNames: expressionAttributeNames,
	}
	_, err = t.dynamodbAccessor.PutItemSdkWithContext(ctx, item, optFns...)
	if err != nil {
		var condErr *types.ConditionalCheckFailedException
		if errors.As(err, &condErr) {
			return ErrKeyDuplicaiton
		}
		return errors.Wrap(err, "CreateOneで登録実行時エラー")
	}
	return nil
}

// FindOneByTableKey implements DynamoDBTemplate.
func (t *defaultDynamoDBTemplate) FindOneByTableKey(tableName tables.DynamoDBTableName, input input.PKOnlyQueryInput, outEntity any, optFns ...func(*dynamodb.Options)) error {
	return t.FindOneByTableKeyWithContext(apcontext.Context, tableName, input, outEntity, optFns...)
}

// FindOneByTableKeyWithContext implements DynamoDBTemplate.
func (t *defaultDynamoDBTemplate) FindOneByTableKeyWithContext(ctx context.Context, tableName tables.DynamoDBTableName, input input.PKOnlyQueryInput, outEntity any, optFns ...func(*dynamodb.Options)) error {
	// プライマリキーの条件
	keyMap, err := CreatePkAttributeValue(input.PrimaryKey)
	if err != nil {
		return errors.Wrap(err, "FindOneByTableKeyで検索条件生成時エラー")
	}
	// 取得項目
	var projection *string
	if len(input.SelectAttributes) > 0 {
		projection = aws.String(strings.Join(input.SelectAttributes, ","))
	}
	// GetItemInput
	getItemInput := &dynamodb.GetItemInput{
		TableName:            aws.String(string(tableName)),
		Key:                  keyMap,
		ProjectionExpression: projection,
		ConsistentRead:       aws.Bool(input.ConsitentRead),
	}
	// GetItemの実行
	getItemOutput, err := t.dynamodbAccessor.GetItemSdkWithContext(ctx, getItemInput, optFns...)
	if err != nil {
		return errors.Wrap(err, "FindOneByTableKeyで検索実行時エラー")
	}
	if len(getItemOutput.Item) == 0 {
		return ErrRecordNotFound
	}
	if err := attributevalue.UnmarshalMap(getItemOutput.Item, &outEntity); err != nil {
		return errors.Wrap(err, "FindOneByTableKeyで検索結果を構造体にアンマーシャル時エラー")
	}

	return nil
}

// FindSomeByTableKey implements DynamoDBTemplate.
func (t *defaultDynamoDBTemplate) FindSomeByTableKey(tableName tables.DynamoDBTableName, input input.PKQueryInput, outEntities any, optFns ...func(*dynamodb.Options)) error {
	return t.FindSomeByTableKeyWithContext(apcontext.Context, tableName, input, outEntities, optFns...)
}

// FindSomeByTableKeyWithContext implements DynamoDBTemplate.
func (t *defaultDynamoDBTemplate) FindSomeByTableKeyWithContext(ctx context.Context, tableName tables.DynamoDBTableName, input input.PKQueryInput, outEntities any, optFns ...func(*dynamodb.Options)) error {
	// クエリ表現の作成
	expr, err := CreateQueryExpressionForTable(input)
	if err != nil {
		return errors.Wrap(err, "FindSomeByTableKeyで検索条件生成時エラー")
	}
	// 最終的な検索結果
	var resultItems []map[string]types.AttributeValue
	// 検索開始キー
	var sKey map[string]types.AttributeValue

	loopCnt := 0
	for {
		loopCnt += 1
		t.logger.Debug("検索回数: %d, 検索開始キー: %v", loopCnt, sKey)
		queryInput := &dynamodb.QueryInput{
			TableName:                 aws.String(string(tableName)),
			KeyConditionExpression:    expr.KeyCondition(),
			ProjectionExpression:      expr.Projection(),
			ExpressionAttributeNames:  expr.Names(),
			ExpressionAttributeValues: expr.Values(),
			FilterExpression:          expr.Filter(),
			ExclusiveStartKey:         sKey,
			ConsistentRead:            aws.Bool(input.ConsitentRead),
			ScanIndexForward:          ScanIndexForward(input.PrimaryKey.SortkeyOrderBy),
		}
		// Queryの実行
		result, err := t.dynamodbAccessor.QuerySDKWithContext(ctx, queryInput, optFns...)
		if err != nil {
			return errors.Wrap(err, "FindSomeByTableKeyで検索実行時エラー")
		}
		// 検索結果の追加
		resultItems = append(resultItems, result.Items...)
		if len(result.LastEvaluatedKey) == 0 {
			// 検索終了
			break
		} else {
			sKey = result.LastEvaluatedKey
		}
	}
	if len(resultItems) == 0 {
		return ErrRecordNotFound
	}
	if err := attributevalue.UnmarshalListOfMaps(resultItems, &outEntities); err != nil {
		return errors.Wrap(err, "FindSomeByTableKeyで検索結果を構造体にアンマーシャル時エラー")
	}
	return nil
}

// FindSomeByGSIKey implements DynamoDBTemplate.
func (t *defaultDynamoDBTemplate) FindSomeByGSIKey(tableName tables.DynamoDBTableName, input input.GsiQueryInput, outEntities any, optFns ...func(*dynamodb.Options)) error {
	return t.FindSomeByGSIKeyWithContext(apcontext.Context, tableName, input, outEntities, optFns...)
}

// FindSomeByGSIKeyWithContext implements DynamoDBTemplate.
func (t *defaultDynamoDBTemplate) FindSomeByGSIKeyWithContext(ctx context.Context, tableName tables.DynamoDBTableName, input input.GsiQueryInput, outEntities any, optFns ...func(*dynamodb.Options)) error {
	// クエリ表現の作成
	expr, err := CreateQueryExpressionForGSI(input)
	if err != nil {
		return errors.Wrap(err, "FindSomeByGSIKeyで検索条件生成時エラー")
	}
	// 最終的な検索結果
	var resultItems []map[string]types.AttributeValue
	// 合計取得件数
	var totalCnt int32
	// ページング回数
	var pagingCnt int
	// ページングの処理
	handleFn := func(result *dynamodb.QueryOutput) bool {
		pagingCnt += 1
		totalCnt += result.Count
		t.logger.Debug("ページング回数: %d, 今回取得件数: %d, 合計取得件数: %d", pagingCnt, result.Count, totalCnt)
		if input.TotalLimit != nil && totalCnt > *input.TotalLimit {
			t.logger.Debug("合計件数: %d, 合計取得件数の上限値: %d, 切り捨て件数: %d", totalCnt, *input.TotalLimit, totalCnt-*input.TotalLimit)
			// 合計取得件数の上限の指定があり、合計取得件数が上限値を超えた場合は切り捨てて追加
			// 例）合計取得件数=120、 上限値=100、既存のリスト件数=95
			//     → 100 - 95 = 5件append可能
			delIdx := int(*input.TotalLimit) - len(resultItems)
			resultItems = append(resultItems, result.Items[:delIdx]...)
			// 上限値を超えたため検索終了
			return true
		}
		// 検索結果の追加
		resultItems = append(resultItems, result.Items...)
		// 最後のページなら検索終了
		return len(result.LastEvaluatedKey) == 0
	}
	// Limitの件数毎に取得する。
	queryInput := &dynamodb.QueryInput{
		TableName:                 aws.String(string(tableName)),
		IndexName:                 aws.String(string(input.GSIName)),
		KeyConditionExpression:    expr.KeyCondition(),
		ProjectionExpression:      expr.Projection(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		Limit:                     input.LimitPerQuery,
		ScanIndexForward:          ScanIndexForward(input.IndexKey.SortkeyOrderBy),
	}
	err = t.dynamodbAccessor.QueryPagesSdkWithContext(ctx, queryInput, handleFn, optFns...)
	if err != nil {
		return errors.Wrap(err, "FindSomeByGSIKeyで検索時エラー")
	}
	if len(resultItems) == 0 {
		return ErrRecordNotFound
	}
	if err := attributevalue.UnmarshalListOfMaps(resultItems, &outEntities); err != nil {
		return errors.Wrap(err, "FindSomeByGSIKeyで検索結果を構造体にアンマーシャル時エラー")
	}
	return nil
}

// UpdateOne implements DynamoDBTemplate.
func (t *defaultDynamoDBTemplate) UpdateOne(tableName tables.DynamoDBTableName, input input.UpdateInput, optFns ...func(*dynamodb.Options)) error {
	return t.UpdateOneWithContext(apcontext.Context, tableName, input, optFns...)
}

// UpdateOneWithContext implements DynamoDBTemplate.
func (t *defaultDynamoDBTemplate) UpdateOneWithContext(ctx context.Context, tableName tables.DynamoDBTableName, input input.UpdateInput, optFns ...func(*dynamodb.Options)) error {
	// プライマリキーの条件
	keyMap, err := CreatePkAttributeValue(input.PrimaryKey)
	if err != nil {
		return errors.Wrap(err, "UpdateOneで更新対象条件の生成時エラー")
	}
	// 更新表現
	expr, err := CreateUpdateExpression(input)
	if err != nil {
		return errors.Wrap(err, "UpdateOneで更新条件の生成時エラー")
	}
	// UpdateItemInput
	updateItemInput := &dynamodb.UpdateItemInput{
		TableName:                 aws.String(string(tableName)),
		Key:                       keyMap,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		UpdateExpression:          expr.Update(),
		ConditionExpression:       expr.Condition(),
		ReturnValues:              types.ReturnValueAllNew,
	}
	// UpdateItemの実行
	_, err = t.dynamodbAccessor.UpdateItemSdkWithContext(ctx, updateItemInput, optFns...)
	if err != nil {
		// 更新条件エラー
		var condErr *types.ConditionalCheckFailedException
		if errors.As(err, &condErr) {
			return ErrUpdateWithCondtion
		}
		return errors.Wrap(err, "UpdateOneで更新実行時エラー")
	}
	return nil
}

// DeleteOne implements DynamoDBTemplate.
func (t *defaultDynamoDBTemplate) DeleteOne(tableName tables.DynamoDBTableName, input input.DeleteInput, optFns ...func(*dynamodb.Options)) error {
	return t.DeleteOneWithContext(apcontext.Context, tableName, input, optFns...)
}

// DeleteOneWithContext implements DynamoDBTemplate.
func (t *defaultDynamoDBTemplate) DeleteOneWithContext(ctx context.Context, tableName tables.DynamoDBTableName, input input.DeleteInput, optFns ...func(*dynamodb.Options)) error {
	// プライマリキーの条件
	keyMap, err := CreatePkAttributeValue(input.PrimaryKey)
	if err != nil {
		return errors.Wrap(err, "DeleteOneで削除対象条件の生成時エラー")
	}
	// 削除表現
	expr, err := CreateDeleteExpression(input)
	if err != nil {
		return errors.Wrap(err, "DelteOneで削除条件の生成時エラー")
	}
	// DeleteItemInput
	deleteItemInput := &dynamodb.DeleteItemInput{
		TableName:                 aws.String(string(tableName)),
		Key:                       keyMap,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ConditionExpression:       expr.Condition(),
		ReturnValues:              types.ReturnValueNone,
	}
	// DeleteItemの実行
	_, err = t.dynamodbAccessor.DeleteItemSdkWithContext(ctx, deleteItemInput, optFns...)
	if err != nil {
		// 削除条件エラー
		var condErr *types.ConditionalCheckFailedException
		if errors.As(err, &condErr) {
			return ErrDeleteWithCondtion
		}
		return errors.Wrap(err, "DeleteOneで削除実行時エラー")
	}
	return nil
}
