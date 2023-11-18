// repositoryのパッケージ
package repository

import "os"

var (
	// DynamoDBのユーザテーブルの名称
	userTable = os.Getenv("USERS_TABLE_NAME")

	// DynamoDBのTODOテーブルの名称
	//todoTable = os.Getenv("TODO_TABLE_NAME")

	// TODO: 暫定的にテストに対処（本来はテスト用の設定情報を読めるConfigの作りに変えないといけない）
	todoTable = func() string {
		todoTableName := os.Getenv("TODO_TABLE_NAME")
		if todoTableName == "" {
			return "todo"
		}
		return todoTableName
	}()
)
