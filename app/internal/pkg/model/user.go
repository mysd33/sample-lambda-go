// modelのパッケージ
package model

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// User ユーザ情報のEntityです。
type User struct {
	// ID は、ユーザのIDです。
	ID string `json:"user_id" dynamodbav:"user_id"`
	// Nameは、ユーザ名です。
	Name string `json:"user_name" dynamodbav:"user_name`
}

// GetKey DynamoDBのキー情報を取得します。
func (user User) GetKey() (map[string]types.AttributeValue, error) {
	id, err := attributevalue.Marshal(user.ID)
	if err != nil {
		return nil, err
	}

	return map[string]types.AttributeValue{"user_id": id}, nil
}
