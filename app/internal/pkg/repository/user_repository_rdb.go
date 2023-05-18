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
	tx := apcontext.Tx
	var user entity.User
	row := tx.QueryRow("SELECT user_id, user_name FROM m_user WHERE user_id = $1", userId)
	err := row.Scan(&user.ID, &user.Name)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (ur *UserRepositoryImplByRDB) PutUser(user *entity.User) (*entity.User, error) {
	tx := apcontext.Tx
	_, err := tx.Exec("INSERT INTO m_user(user_id, user_name) VALUES($1, $2)", user.ID, user.Name)

	if err != nil {
		return nil, err
	}
	return user, nil
}
