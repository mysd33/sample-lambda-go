/*
dynamodb パッケージは、DynamoDBアクセスに関する機能を提供するパッケージです。
*/
package dynamodb

import (
	"strconv"

	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/dynamodb/input"
	"example.com/appbase/pkg/env"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/cockroachdb/errors"
)

// ユーティリティ関数

const (
	DYNAMODB_RETURN_CONSUMED_CAPACITY_NAME = "DYNAMODB_RETURN_CONSUMED_CAPACITY"
)

// ReturnConsumedCapacity は、ConsumedCapacityを返却するかどうかを返します。
func ReturnConsumedCapacity(config config.Config) bool {
	// 環境変数が存在する場合はその値を返却
	if value, found := config.GetBoolWithContains(DYNAMODB_RETURN_CONSUMED_CAPACITY_NAME); found {
		return value
	}
	// デフォルトでは開発時のみConsumedCapacityを返却する
	return !env.IsStragingOrProd()
}

// CreatePkAttributeValue は、プライマリキーの完全一致による条件のAttributeValueのマップを作成します。
func CreatePkAttributeValue(primaryKey input.PrimaryKey) (map[string]types.AttributeValue, error) {
	keymap := map[string]types.AttributeValue{}
	// パーティションキー
	partitionKey := primaryKey.PartitionKey
	pk, err := typeSwitch(partitionKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	keymap[partitionKey.Name] = pk

	// ソートキー
	sortKey := primaryKey.SortKey
	if sortKey != nil {
		sk, err := typeSwitch(*sortKey)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		keymap[sortKey.Name] = sk
	}
	return keymap, nil
}

// CreateQueryExpressionForTable は、 ベーステーブルに対するクエリの条件のExpressionを作成します。
func CreateQueryExpressionForTable(input input.PKQueryInput) (*expression.Expression, error) {
	primaryKey := &input.PrimaryKey
	return createQueryExpression(primaryKey, input.SelectAttributes, input.WhereClauses)
}

// CreateQueryExpressionForGSI は、 GSIに対するクエリの条件のExpressionを作成します。
func CreateQueryExpressionForGSI(input input.GsiQueryInput) (*expression.Expression, error) {
	primaryKey := &input.IndexKey
	return createQueryExpression(primaryKey, input.SelectAttributes, input.WhereClauses)
}

func createQueryExpression(primaryKey *input.PrimaryKey, attributes []string, whereCauses []*input.WhereClause) (*expression.Expression, error) {
	keyCond, err := CreateKeyCondition(primaryKey)
	if err != nil {
		return nil, err
	}
	// キーによる検索条件の指定
	eb := expression.NewBuilder().WithKeyCondition(*keyCond)
	// 取得項目の設定
	proj := CreateProjection(attributes)
	if proj != nil {
		eb = eb.WithProjection(*proj)
	}
	// フィルタ条件の設定
	filterCond, err := CreateWhereCondition(whereCauses)
	if err != nil {
		return nil, err
	}
	if filterCond != nil {
		eb.WithFilter(*filterCond)
	}
	// クエリ表現の作成
	expr, err := eb.Build()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &expr, nil
}

// CreateUpdateExpression は、更新条件のExpressionを作成します。
func CreateUpdateExpression(input input.UpdateInput) (*expression.Expression, error) {
	// 更新項目の設定
	upd := expression.UpdateBuilder{}
	for _, attr := range input.UpdateAttributes {
		if attr != nil {
			upd = upd.Set(expression.Name(attr.Name), expression.Value(attr.Value))
		}
	}
	// 削除項目の設定
	for _, name := range input.RemoveAttributeNames {
		upd = upd.Remove(expression.Name(name))
	}
	// Update表現の作成
	eb := expression.NewBuilder().WithUpdate(upd)
	// 更新条件の作成
	updCond, err := CreateWhereCondition(input.WhereClauses)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if updCond != nil {
		eb = eb.WithCondition(*updCond)
	}
	expr, err := eb.Build()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &expr, nil
}

// CreateDeleteExpression は、削除条件のExpressionを作成します。
func CreateDeleteExpression(input input.DeleteInput) (*expression.Expression, error) {
	// 削除条件の作成
	delCond, err := CreateWhereCondition(input.WhereClauses)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	// Delete表現の作成
	if delCond == nil {
		return nil, nil
	}
	eb := expression.NewBuilder().WithCondition(*delCond)
	expr, err := eb.Build()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &expr, nil
}

// CreateKeyCondition は、キー条件を作成します。
func CreateKeyCondition(primaryKeyCond *input.PrimaryKey) (*expression.KeyConditionBuilder, error) {
	// パーティションキーの条件
	keyCond := expression.Key(primaryKeyCond.PartitionKey.Name).Equal(expression.Value(primaryKeyCond.PartitionKey.Value))
	// ソートキーがある場合
	if primaryKeyCond.SortKey != nil {
		switch primaryKeyCond.SortKeyOp {
		case input.SORTKEY_BEGINS_WITH:
			if v, ok := primaryKeyCond.SortKey.Value.(string); ok {
				keyCond = keyCond.And(expression.Key(primaryKeyCond.SortKey.Name).BeginsWith(v))
			} else {
				return nil, errors.New("type not supported")
			}
		case input.SORTKEY_BETWEEN:
			// primaryKey.SortKey.Value[0] <= ソートキー <= primaryKey.SortKey.Value[1]
			if v, ok := primaryKeyCond.SortKey.Value.([2]any); ok {
				keyCond = keyCond.And(expression.Key(primaryKeyCond.SortKey.Name).Between(expression.Value(v[0]), expression.Value(v[1])))
			} else {
				return nil, errors.New("type not supported")
			}

		case input.SORTKEY_GREATER_THAN:
			keyCond = keyCond.And(expression.Key(primaryKeyCond.SortKey.Name).GreaterThan(expression.Value(primaryKeyCond.SortKey.Value)))
		case input.SORTKEY_GREATER_THAN_EQ:
			keyCond = keyCond.And(expression.Key(primaryKeyCond.SortKey.Name).GreaterThanEqual(expression.Value(primaryKeyCond.SortKey.Value)))
		case input.SORTKEY_LESS_THAN:
			keyCond = keyCond.And(expression.Key(primaryKeyCond.SortKey.Name).LessThan(expression.Value(primaryKeyCond.SortKey.Value)))
		case input.SORTKEY_LESS_THAN_EQ:
			keyCond = keyCond.And(expression.Key(primaryKeyCond.SortKey.Name).LessThanEqual(expression.Value(primaryKeyCond.SortKey.Value)))
		default:
			//それ以外はEqualで扱う
			keyCond = keyCond.And(expression.Key(primaryKeyCond.SortKey.Name).Equal(expression.Value(primaryKeyCond.SortKey.Value)))
		}
	}
	return &keyCond, nil
}

// CreateScanInput は、検索時の取得（射影）項目を作成します。
func CreateProjection(attributeNames []string) *expression.ProjectionBuilder {
	if len(attributeNames) == 0 {
		return nil
	}
	var proj expression.ProjectionBuilder
	for _, v := range attributeNames {
		proj = expression.AddNames(proj, expression.Name(v))
	}
	return &proj
}

// CreateWhereCondition は、Where句をもとにした条件を作成します。
func CreateWhereCondition(whereClauses []*input.WhereClause) (*expression.ConditionBuilder, error) {
	fn := func(where input.WhereClause, cond *expression.ConditionBuilder) (*expression.ConditionBuilder, error) {
		var tmp expression.ConditionBuilder
		switch where.WhereOp {
		case input.WHERE_EQUAL:
			tmp = expression.Name(where.Attribute.Name).Equal(expression.Value(where.Attribute.Value))
		case input.WHERE_NOT_EQUAL:
			tmp = expression.Name(where.Attribute.Name).NotEqual(expression.Value(where.Attribute.Value))
		case input.WHERE_BEGINS_WITH:
			if v, ok := where.Attribute.Value.(string); ok {
				tmp = expression.Name(where.Attribute.Name).BeginsWith(v)
			} else {
				return nil, errors.New("type not supported")
			}
		case input.WHERE_GREATER_THAN:
			tmp = expression.Name(where.Attribute.Name).GreaterThan(expression.Value(where.Attribute.Value))
		case input.WHERE_GREATER_THAN_EQ:
			tmp = expression.Name(where.Attribute.Name).GreaterThanEqual(expression.Value(where.Attribute.Value))
		case input.WHERE_LESS_THAN:
			tmp = expression.Name(where.Attribute.Name).LessThan(expression.Value(where.Attribute.Value))
		case input.WHERE_LESS_THAN_EQ:
			tmp = expression.Name(where.Attribute.Name).LessThanEqual(expression.Value(where.Attribute.Value))
		default:
			return nil, errors.New("operator not supported")
		}
		if cond != nil {
			if where.AppendOp == input.APPEND_OR {
				tmp = cond.Or(*cond, tmp)
			} else {
				tmp = cond.And(*cond, tmp)
			}
		}
		return &tmp, nil
	}
	var filterCond *expression.ConditionBuilder
	var conderr error
	for _, where := range whereClauses {
		if where != nil {
			filterCond, conderr = fn(*where, filterCond)
			if conderr != nil {
				return nil, conderr
			}
		}
	}
	return filterCond, nil
}

// StandardIndexForward は、インデックスの検索順序を返します。
func ScanIndexForward(orderby input.OrderBy) *bool {
	switch orderby {
	case input.ORDER_BY_DESC:
		return aws.Bool(false)
	default:
		return aws.Bool(true)
	}
}

func typeSwitch(attribute input.Attribute) (types.AttributeValue, error) {
	switch attribute.Value.(type) {
	case nil:
		return nil, errors.Errorf("cannot switch type because attribuite is nil")
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return &types.AttributeValueMemberN{Value: strconv.Itoa(attribute.Value.(int))}, nil
	case string:
		return &types.AttributeValueMemberS{Value: attribute.Value.(string)}, nil
	case bool:
		return &types.AttributeValueMemberBOOL{Value: attribute.Value.(bool)}, nil
	case []byte:
		return &types.AttributeValueMemberB{Value: attribute.Value.([]byte)}, nil
	default:
		return nil, errors.Errorf("type not supported: %T", attribute.Value)
	}
}
