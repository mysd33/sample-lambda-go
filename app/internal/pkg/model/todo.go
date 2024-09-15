// entityのパッケージ
package model

// Todo はやることリスト（Todo）のEntityです。
type Todo struct {
	// ID は、TodoのIDです。
	ID string `json:"todo_id" dynamodbav:"todo_id"`
	// Title は、Todoのタイトルです。
	Title string `json:"todo_title" dynamodbav:"todo_title"`
}

// 従来のDynamoDBAccessorを使ったコード
// GetKey は、DynamoDBのキー情報を取得します。
/*
func (todo Todo) GetKey() (map[string]types.AttributeValue, error) {
	id, err := attributevalue.Marshal(todo.ID)
	if err != nil {
		return nil, err
	}

	return map[string]types.AttributeValue{"todo_id": id}, nil
}*/
