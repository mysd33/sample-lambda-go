// repositoryのパッケージ
package repository

import "app/internal/pkg/model"

// TodoRepository は、Todoを管理するRepositoryインタフェースです。
type TodoRepository interface {
	// FindOne は、todoIdが一致するTodoを取得します。
	FindOne(todoId string) (*model.Todo, error)
	// CreateOne は、指定されたTodoを登録します。
	CreateOne(todo *model.Todo) (*model.Todo, error)
	// CreateOneTx は、指定されたTodoをトランザクションを使って登録します。
	CreateOneTx(todo *model.Todo) (*model.Todo, error)
}
