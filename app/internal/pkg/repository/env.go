// repositoryのパッケージ
package repository

import "os"

var (
	// DynamoDBのユーザテーブルの名称
	userTable = os.Getenv("USERS_TABLE_NAME")
	// DynamoDBのTODOテーブルの名称
	todoTable = os.Getenv("TODO_TABLE_NAME")
)
