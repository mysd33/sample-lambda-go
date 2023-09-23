package entity

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type User struct {
	ID   string `json:"user_id" dynamodbav:"user_id"`
	Name string `json:"user_name" dynamodbav:"user_id"`
}

func (user User) GetKey() (map[string]types.AttributeValue, error) {
	id, err := attributevalue.Marshal(user.ID)
	if err != nil {
		return nil, err
	}

	return map[string]types.AttributeValue{"user_id": id}, nil
}
