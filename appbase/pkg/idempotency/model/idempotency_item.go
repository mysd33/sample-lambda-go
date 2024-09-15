/*
model パッケージは、冪等性管理テーブルに関連するエンティティを提供します。
*/
package model

// IdempotencyItem は、冪等性管理テーブルのアイテムを表す構造体です。
type IdempotencyItem struct {
	//
	IdempotencyKey   string `dynamodbav:"idempotency_key"`
	Expiry           int64  `dynamodbav:"expiry"`
	InprogressExpiry int64  `dynamodbav:"inprogress_expiry"`
	Status           string `dynamodbav:"status"`
}
