// repositoryのパッケージ
package repository

import "app/internal/pkg/model"

// UserRepository は、ユーザを管理するRepositoryインタフェースです。
type UserRepository interface {
	// FindOne は、userIdが一致するユーザを取得します。
	FindOne(userId string) (*model.User, error)
	// CreateOne は、指定されたユーザを登録します。
	CreateOne(user *model.User) (*model.User, error)
}
