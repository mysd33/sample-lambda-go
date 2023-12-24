/*
dynamodb パッケージは、DynamoDBアクセスに関する機能を提供するパッケージです。
*/
package dynamodb

import (
	"strconv"

	"example.com/appbase/pkg/dynamodb/criteria"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/cockroachdb/errors"
)

// ユーティリティ関数

// CreatePkAttributeValue は、プライマリキーの完全一致による条件のAttributeValueのマップを作成します。
func CreatePkAttributeValue(primaryKey criteria.KeyPair) (map[string]types.AttributeValue, error) {
	keymap := map[string]types.AttributeValue{}
	// パーティションキー
	partitionKey := primaryKey.PartitionKey
	pk, err := typeSwitch(partitionKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	keymap[partitionKey.Key] = pk

	// ソートキー
	sortKey := primaryKey.SortKey
	if sortKey != nil {
		sk, err := typeSwitch(*sortKey)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		keymap[sortKey.Key] = sk
	}
	return keymap, nil
}

// CreateUpdateExpressionBuilder は、更新条件のExpressionを作成します。
func CreateUpdateExpressionBuilder(input criteria.UpdateInput) (*expression.Expression, error) {
	updCond, err := CreateFilterCondition(input.WhereKeys)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	// 更新項目の設定
	upd := expression.UpdateBuilder{}
	for _, attr := range input.UpdateAttributes {
		if attr != nil {
			upd = upd.Set(expression.Name(attr.Key), expression.Value(attr.Value))
		}
	}
	// Update表現の作成
	eb := expression.NewBuilder().WithUpdate(upd)
	if updCond != nil {
		eb = eb.WithCondition(*updCond)
	}
	expr, _ := eb.Build()
	return &expr, nil
}

// CreateDeleteExpressionBuilder は、削除条件のExpressionを作成します。
func CreateDeleteExpressionBuilder(input criteria.DeleteInput) (*expression.Expression, error) {
	delCond, err := CreateFilterCondition(input.WhereKeys)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	// Delete表現の作成
	eb := expression.NewBuilder().WithCondition(*delCond)
	expr, _ := eb.Build()
	return &expr, nil
}

// CreateFilterCondition は、フィルタ条件を作成します。
func CreateFilterCondition(whereClauses []*criteria.WhereClause) (*expression.ConditionBuilder, error) {
	fn := func(where criteria.WhereClause, cond *expression.ConditionBuilder) (*expression.ConditionBuilder, error) {
		var tmp expression.ConditionBuilder
		switch where.Operator {
		case criteria.WHERE_EQUAL:
			tmp = expression.Name(where.KeyValue.Key).Equal(expression.Value(where.KeyValue.Value))
		case criteria.WHERE_NOT_EQUAL:
			tmp = expression.Name(where.KeyValue.Key).NotEqual(expression.Value(where.KeyValue.Value))
		case criteria.WHERE_BEGINS_WITH:
			if v, ok := where.KeyValue.Value.(string); ok {
				tmp = expression.Name(where.KeyValue.Key).BeginsWith(v)
			} else {
				return nil, errors.New("type not supported")
			}
		case criteria.WHERE_GREATER_THAN:
			tmp = expression.Name(where.KeyValue.Key).GreaterThan(expression.Value(where.KeyValue.Value))
		case criteria.WHERE_GREATER_THAN_EQ:
			tmp = expression.Name(where.KeyValue.Key).GreaterThanEqual(expression.Value(where.KeyValue.Value))
		case criteria.WHERE_LESS_THAN:
			tmp = expression.Name(where.KeyValue.Key).LessThan(expression.Value(where.KeyValue.Value))
		case criteria.WHERE_LESS_THAN_EQ:
			tmp = expression.Name(where.KeyValue.Key).LessThanEqual(expression.Value(where.KeyValue.Value))
		default:
			return nil, errors.New("operator not supported")
		}
		if cond != nil {
			if where.AppendOperator == criteria.APPEND_OR {
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

func typeSwitch(keyValue criteria.KeyValue) (types.AttributeValue, error) {
	switch keyValue.Value.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		// TODO: 要確認
		//return &types.AttributeValueMemberN{Value: keyValue.Value.(string)}, nil
		return &types.AttributeValueMemberN{Value: strconv.Itoa(keyValue.Value.(int))}, nil
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
