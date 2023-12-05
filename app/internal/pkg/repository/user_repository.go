// repositoryのパッケージ
package repository

import (
	"app/internal/pkg/entity"
)

// UserRepository は、ユーザを管理するRepositoryインタフェースです。
type UserRepository interface {
	// FindOne は、userIdが一致するユーザを取得します。
	FindOne(userId string) (*entity.User, error)
	// CreateOne は、指定されたユーザを登録します。
	CreateOne(user *entity.User) (*entity.User, error)
}
