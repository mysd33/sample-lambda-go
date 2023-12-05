// repositoryのパッケージ
package repository

import (
	"app/internal/pkg/entity"
	"fmt"

	"example.com/appbase/pkg/apcontext"
	"example.com/appbase/pkg/id"
	"example.com/appbase/pkg/logging"
	"example.com/appbase/pkg/rdb"
	"github.com/lib/pq"
)

// NewUserRepositoryForRDB は、RDB保存のためのUserRepository実装を作成します。
func NewUserRepositoryForRDB(accessor rdb.RDBAccessor, log logging.Logger) UserRepository {
	return &UserRepositoryImplByRDB{accessor: accessor, log: log}
}

// UserRepositoryImplByRDB は、RDB保存のためのUserRepository実装です。
type UserRepositoryImplByRDB struct {
	accessor rdb.RDBAccessor
	log      logging.Logger
}

func (ur *UserRepositoryImplByRDB) FindOne(userId string) (*entity.User, error) {
	tx := ur.accessor.GetTransaction()
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

func (ur *UserRepositoryImplByRDB) CreateOne(user *entity.User) (*entity.User, error) {
	//ID採番
	userId := id.GenerateId()
	user.ID = userId

	tx := ur.accessor.GetTransaction()
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
