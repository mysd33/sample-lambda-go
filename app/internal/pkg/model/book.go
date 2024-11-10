// modelのパッケージ
package model

import "go.mongodb.org/mongo-driver/bson/primitive"

// Book は、書籍の情報を表す構造体です。
type Book struct {
	// （参考） bsonの構造タグを利用した、MongoDBのフィールド名の指定
	// https://www.mongodb.com/ja-jp/docs/drivers/go/current/usage-examples/insertOne/
	// https://www.mongodb.com/ja-jp/docs/drivers/go/current/fundamentals/bson/#struct-tags

	// ObjectID（_id）は、書籍のIDです。DocumentDB(MongoDB)のObjectIDを利用します。
	ObjectID primitive.ObjectID `json:"object_id,omitempty" bson:"_id,omitempty"`
	// Title は、書籍のタイトルです。
	Title string `json:"title" bson:"title"`
	// Author は、書籍の著者です。
	Author string `json:"author" bson:"author"`
	// Publisher は、書籍の出版社です。
	Publisher string `json:"publisher,omitempty" bson:"publisher,omitempty"`
	// PublishedDate は、書籍の発売日です。
	PublishedDate string `json:"published_date,omitempty" bson:"published_date,omitempty"`
	// ISBNは、書籍のISBNです。
	ISBN string `json:"isbn,omitempty" bson:"isbn,omitempty"`
}
