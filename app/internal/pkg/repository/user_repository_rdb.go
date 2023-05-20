package repository

import (
	"app/internal/pkg/entity"

	"example.com/appbase/pkg/apcontext"
	"example.com/appbase/pkg/id"
)

func NewUserRepositoryForRDB() UserRepository {
	return &UserRepositoryImplByRDB{}
}

type UserRepositoryImplByRDB struct {
}

func (ur *UserRepositoryImplByRDB) GetUser(userId string) (*entity.User, error) {
	tx := apcontext.Tx
	ctx := apcontext.Context
	var user entity.User
	//row := tx.QueryRow("SELECT user_id, user_name FROM m_user WHERE user_id = $1", userId)
	//X-RayのSQLトレース対応
	row := tx.QueryRowContext(ctx, "SELECT user_id, user_name FROM m_user WHERE user_id = $1", userId)
	err := row.Scan(&user.ID, &user.Name)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (ur *UserRepositoryImplByRDB) PutUser(user *entity.User) (*entity.User, error) {
	//ID採番
	userId := id.GenerateId()
	user.ID = userId

	tx := apcontext.Tx
	ctx := apcontext.Context
	//_, err := tx.Exec("INSERT INTO m_user(user_id, user_name) VALUES($1, $2)", user.ID, user.Name)
	//X-RayのSQLトレース対応
	_, err := tx.ExecContext(ctx, "INSERT INTO m_user(user_id, user_name) VALUES($1, $2)", user.ID, user.Name)
	if err != nil {
		return nil, err
	}
	return user, nil
}
