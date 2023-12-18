// entityのパッケージ
package entity

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// Dummy は、ダミーのテーブルの構造体です。
type Dummy struct {
	// ID は、ダミーテーブルのIDです。
	ID string `json:"dummy_id" dynamodbav:"dummy_id"`
	// Value は、ダミーテーブルの値です。
	Value string `json:"dummy_value" dynamodbav:"dummy_value"`
}

// GetKey は、DynamoDBのキー情報を取得します。
func (d Dummy) GetKey() (map[string]types.AttributeValue, error) {
	id, err := attributevalue.Marshal(d.ID)
	if err != nil {
		return nil, err
	}

	return map[string]types.AttributeValue{"dummy_id": id}, nil
}
