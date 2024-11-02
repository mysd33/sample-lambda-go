/*
model パッケージは、キューメッセージ管理テーブルに関連するエンティティを提供します。
*/
package model

// QueueMessageItem は、キューメッセージ管理テーブルのアイテムを表す構造体です。
type QueueMessageItem struct {
	MessageId      string `dynamodbav:"message_id"`
	MessageGroupId string `dynamodbav:"message_group_id"`
	DeleteTime     int    `dynamodbav:"delete_time"`
	Status         string `dynamodbav:"status"`
}

/*
func (m QueueMessageItem) GetKey() (map[string]types.AttributeValue, error) {
	id, err := attributevalue.Marshal(m.MessageId)
	if err != nil {
		return nil, err
	}

	return map[string]types.AttributeValue{"message_id": id}, nil
}*/
