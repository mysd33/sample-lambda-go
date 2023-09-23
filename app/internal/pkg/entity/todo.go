package entity

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type Todo struct {
	ID    string `json:"todo_id" dynamodbav:"todo_id"`
	Title string `json:"todo_title" dynamodbav:"todo_title"`
}

func (todo Todo) GetKey() (map[string]types.AttributeValue, error) {
	id, err := attributevalue.Marshal(todo.ID)
	if err != nil {
		return nil, err
	}

	return map[string]types.AttributeValue{"todo_id": id}, nil
}
