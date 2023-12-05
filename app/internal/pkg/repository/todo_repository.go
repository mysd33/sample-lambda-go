// repositoryのパッケージ
package repository

import (
	"app/internal/pkg/entity"
)

// TodoRepository は、Todoを管理するRepositoryインタフェースです。
type TodoRepository interface {
	// FindOne は、todoIdが一致するTodoを取得します。
	FindOne(todoId string) (*entity.Todo, error)
	// CreateOne は、指定されたTodoを登録します。
	CreateOne(todo *entity.Todo) (*entity.Todo, error)
	// CreateOneTx は、指定されたTodoをトランザクションを使って登録します。
	CreateOneTx(todo *entity.Todo) (*entity.Todo, error)
}
