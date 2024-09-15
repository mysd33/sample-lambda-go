// modelのパッケージ
package model

// Temp は、一時テーブルの構造体です。
type Temp struct {
	// ID は、ダミーテーブルのIDです。
	ID string `json:"id" dynamodbav:"id"`
	// Value は、ダミーテーブルの値です。
	Value string `json:"value" dynamodbav:"value"`
}

// 従来のDynamoDBAccessorを使ったコード
// GetKey は、DynamoDBのキー情報を取得します。
/*
func (t Temp) GetKey() (map[string]types.AttributeValue, error) {
	id, err := attributevalue.Marshal(t.ID)
	if err != nil {
		return nil, err
	}

	return map[string]types.AttributeValue{"id": id}, nil
}*/
