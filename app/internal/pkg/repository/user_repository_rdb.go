package repository

import (
	"app/internal/pkg/entity"
	"fmt"

	"example.com/appbase/pkg/apcontext"
	"example.com/appbase/pkg/id"
	"example.com/appbase/pkg/rdb"
	"github.com/lib/pq"
)

func NewUserRepositoryForRDB() UserRepository {
	return &UserRepositoryImplByRDB{}
}

type UserRepositoryImplByRDB struct {
}

func (ur *UserRepositoryImplByRDB) GetUser(userId string) (*entity.User, error) {
	tx := rdb.Tx
	ctx := apcontext.Context
	var user entity.User
	//プリペアードステートメントによる例
	//X-RayのSQLトレースにも対応
	//row := tx.QueryRowContext(ctx, "SELECT user_id, user_name FROM m_user WHERE user_id = $1", userId)

	//プリペアードステートメント未使用の例
	//X-RayのSQLトレースにも対応
	//RDS Proxy経由で接続する場合、プリペアードステートメントを使用すると、
	//ピン留め（RDSProxyはコネクションプール内のDB接続を特定のDBクライアントに対して固定）されてしまうことを回避
	//https://qiita.com/neruneruo/items/2313feed6d4ce28c2061
	row := tx.QueryRowContext(ctx, fmt.Sprintf("SELECT user_id, user_name FROM m_user WHERE user_id = %s", pq.QuoteLiteral(userId)))

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

	tx := rdb.Tx
	ctx := apcontext.Context
	//プリペアードステートメントによる例
	//X-RayのSQLトレース対応にも対応
	//_, err := tx.ExecContext(ctx, "INSERT INTO m_user(user_id, user_name) VALUES($1, $2)", user.ID, user.Name)

	//プリペアードステートメント未使用の例
	//X-RayのSQLトレース対応にも対応
	//RDS Proxy経由で接続する場合、プリペアードステートメントを使用すると、
	//ピン留め（RDSProxyはコネクションプール内のDB接続を特定のDBクライアントに対して固定）されてしまうこと回避
	//https://qiita.com/neruneruo/items/2313feed6d4ce28c2061
	//SQLインジェクション対策でQuoteLiteralメソッドでエスケープ
	userIdParam := pq.QuoteLiteral(user.ID)
	userNameParam := pq.QuoteLiteral(user.Name)
	_, err := tx.ExecContext(ctx, fmt.Sprintf("INSERT INTO m_user(user_id, user_name) VALUES(%s, %s)", userIdParam, userNameParam))

	if err != nil {
		return nil, err
	}
	return user, nil
}
