package repository

import (
	"app/internal/pkg/entity"
)

type UserRepository interface {
	GetUser(userId string) (*entity.User, error)
	PutUser(user *entity.User) (*entity.User, error)
}
