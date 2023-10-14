// repositoryのパッケージ
package repository

import "os"

var (
	// リージョン名
	region = os.Getenv("REGION")
	// DynamoDBのユーザテーブルの名称
	userTable = os.Getenv("USERS_TABLE_NAME")
	// DynamoDBのTODOテーブルの名称
	todoTable = os.Getenv("TODO_TABLE_NAME")
)
