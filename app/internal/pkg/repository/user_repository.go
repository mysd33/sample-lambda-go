// repositoryのパッケージ
package repository

import (
	"app/internal/pkg/entity"
)

// UserRepository は、ユーザを管理するRepositoryインタフェースです。
type UserRepository interface {
	// GetUser は、userIdが一致するユーザを取得します。
	GetUser(userId string) (*entity.User, error)
	// PutUser は、指定されたユーザを登録します。
	PutUser(user *entity.User) (*entity.User, error)
}
