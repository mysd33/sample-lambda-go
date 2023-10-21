// repositoryのパッケージ
package repository

import (
	"app/internal/pkg/entity"
)

// TodoRepository は、Todoを管理するRepositoryインタフェースです。
type TodoRepository interface {
	// GetTodo は、todoIdが一致するTodoを取得します。
	GetTodo(todoId string) (*entity.Todo, error)
	// PutTodo は、指定されたTodoを登録します。
	PutTodo(todo *entity.Todo) (*entity.Todo, error)
}
