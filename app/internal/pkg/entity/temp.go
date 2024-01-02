// entityのパッケージ
package entity

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// Temp は、一時テーブルの構造体です。
type Temp struct {
	// ID は、ダミーテーブルのIDです。
	ID string `json:"id" dynamodbav:"id"`
	// Value は、ダミーテーブルの値です。
	Value string `json:"value" dynamodbav:"value"`
}

// GetKey は、DynamoDBのキー情報を取得します。
func (t Temp) GetKey() (map[string]types.AttributeValue, error) {
	id, err := attributevalue.Marshal(t.ID)
	if err != nil {
		return nil, err
	}

	return map[string]types.AttributeValue{"id": id}, nil
}
