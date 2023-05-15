package repository

import (
	"app/internal/pkg/entity"

	"example.com/appbase/pkg/apcontext"
)

func NewUserRepositoryForRDB() UserRepository {
	return &UserRepositoryImplByRDB{}
}

type UserRepositoryImplByRDB struct {
}

func (ur *UserRepositoryImplByRDB) GetUser(userId string) (*entity.User, error) {
	db := apcontext.DB
	var user entity.User
	row := db.QueryRow("SELECT * FROM user WHEN user_id = ?", userId)
	err := row.Scan(&user.ID, &user.Name)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (ur *UserRepositoryImplByRDB) PutUser(user *entity.User) (*entity.User, error) {
	db := apcontext.DB
	_, err := db.Exec("INSERT INTO user (user_id, name) VALUES (?, ?)", user.ID, user.Name)
	if err != nil {
		return nil, err
	}
	return user, nil
}
