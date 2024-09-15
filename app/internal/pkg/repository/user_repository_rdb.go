// repositoryのパッケージ
package repository

import (
	"app/internal/pkg/entity"
	"app/internal/pkg/message"
	"database/sql"
	"fmt"

	"example.com/appbase/pkg/apcontext"
	"example.com/appbase/pkg/errors"
	"example.com/appbase/pkg/id"
	"example.com/appbase/pkg/logging"
	"example.com/appbase/pkg/rdb"
	"github.com/lib/pq"
)

// NewUserRepositoryForRDB は、RDB保存のためのUserRepository実装を作成します。
func NewUserRepositoryForRDB(accessor rdb.RDBAccessor, logger logging.Logger, id id.IDGenerator) UserRepository {
	return &UserRepositoryImplByRDB{accessor: accessor, logger: logger, id: id}
}

// UserRepositoryImplByRDB は、RDB保存のためのUserRepository実装です。
type UserRepositoryImplByRDB struct {
	accessor rdb.RDBAccessor
	logger   logging.Logger
	id       id.IDGenerator
}

func (ur *UserRepositoryImplByRDB) FindOne(userId string) (*entity.User, error) {

	// RDS Proxy経由で接続する場合、１つのトランザクション内での呼び出しは、同じコネクションを使用する
	// auto commit無効の場合は、トランザクションが終了（commit/rollback）するまで、接続の再利用は行われない
	// このため、後述のプリペアドステートメントによるピン留めを過度に気にする必要はないかもしれない。
	// https://pages.awscloud.com/rs/112-TZM-766/images/EV_amazon-rds-aws-lambda-update_Jul28-2020_RDS_Proxy.pdf
	// pp.12-13
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
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.NewBusinessErrorWithCause(err, message.W_EX_8009, userId)
		}
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}
	return &user, nil
}

func (ur *UserRepositoryImplByRDB) CreateOne(user *entity.User) (*entity.User, error) {
	//ID採番
	userId, err := ur.id.GenerateUUID()
	if err != nil {
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}
	user.ID = userId

	// RDS Proxy経由で接続する場合、１つのトランザクション内での呼び出しは、同じコネクションを使用する
	// auto commit無効の場合は、トランザクションが終了（commit/rollback）するまで、接続の再利用は行われない
	// このため、後述のプリペアドステートメントによるピン留めを過度に気にする必要はないかもしれない。
	// https://pages.awscloud.com/rs/112-TZM-766/images/EV_amazon-rds-aws-lambda-update_Jul28-2020_RDS_Proxy.pdf
	// pp.12-13
	tx := ur.accessor.GetTransaction()
	ctx := apcontext.Context

	// プリペアードステートメントによる例
	// X-RayのSQLトレース対応にも対応
	//_, err := tx.ExecContext(ctx, "INSERT INTO m_user(user_id, user_name) VALUES($1, $2)", user.ID, user.Name)

	// プリペアードステートメント未使用の例
	// X-RayのSQLトレース対応にも対応
	// RDS Proxy経由で接続する場合、プリペアードステートメントを使用すると、
	// ピン留め（RDSProxyはコネクションプール内のDB接続を特定のDBクライアントに対して固定）されてしまうこと回避
	// https://qiita.com/neruneruo/items/2313feed6d4ce28c2061
	// https://docs.aws.amazon.com/ja_jp/AmazonRDS/latest/UserGuide/rds-proxy-managing.html#rds-proxy-pinning
	// SQLインジェクション対策でQuoteLiteralメソッドでエスケープ
	userIdParam := pq.QuoteLiteral(user.ID)
	userNameParam := pq.QuoteLiteral(user.Name)
	_, err = tx.ExecContext(ctx, fmt.Sprintf("INSERT INTO m_user(user_id, user_name) VALUES(%s, %s)", userIdParam, userNameParam))

	if err != nil {
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}
	return user, nil
}
