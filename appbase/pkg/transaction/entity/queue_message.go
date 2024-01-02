package entity

// QueueMessageItem は、QueueMessageテーブルのアイテムを表す構造体です。
type QueueMessageItem struct {
	MessageId              string `dynamodbav:"message_id"`
	DeleteTime             int    `dynamodbav:"delete_time"`
	MessageDeduplicationId string `dynamodbav:"message_deduplication_id"`
}

/*
func (m QueueMessageItem) GetKey() (map[string]types.AttributeValue, error) {
	id, err := attributevalue.Marshal(m.MessageId)
	if err != nil {
		return nil, err
	}

	return map[string]types.AttributeValue{"message_id": id}, nil
}*/
