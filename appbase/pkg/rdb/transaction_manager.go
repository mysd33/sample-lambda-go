/*
rdb パッケージは、RDBアクセスに関する機能を提供するパッケージです。
*/
package rdb

import (
	"database/sql"
	"fmt"

	"example.com/appbase/pkg/apcontext"
	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/domain"
	"example.com/appbase/pkg/logging"
	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/cockroachdb/errors"

	_ "github.com/lib/pq"
)

const (
	RDB_USER_NAME     = "username"
	RDB_PASSWORD_NAME = "password"
	RDB_ENDPOINT_NAME = "RDB_ENDPOINT"
	RDB_PORT_NAME     = "RDB_PORT"
	RDB_DB_NAME_NAME  = "RDB_DB_NAME"
	RDB_SSL_MODE_NAME = "RDB_SSL_MODE"
)

// TransactionManager はトランザクションを管理するインタフェースです
type TransactionManager interface {
	// ExecuteTransaction は、Serviceの関数serviceFuncの実行前後でDynamoDBトランザクション実行します。
	ExecuteTransaction(serviceFunc domain.ServiceFunc) (any, error)
}

// NewTransactionManager は、TransactionManagerを作成します
func NewTransactionManager(log logging.Logger, config config.Config, rdbAccessor RDBAccessor) TransactionManager {
	return &defaultTransactionManager{log: log, config: config, rdbAccessor: rdbAccessor}
}

// defaultTransactionManager は、TransactionManagerを実装する構造体です。
type defaultTransactionManager struct {
	log         logging.Logger
	config      config.Config
	tx          *sql.Tx
	rdbAccessor RDBAccessor
}

// ExecuteTransaction implements TransactionManager.
func (tm *defaultTransactionManager) ExecuteTransaction(serviceFunc domain.ServiceFunc) (any, error) {
	db, err := tm.rdbConnect()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	// 終了時にRDBコネクションの切断
	defer db.Close()
	// RDBトランザクション開始
	tm.tx, err = tm.startTransaction(db)
	if err != nil {
		return nil, err
	}
	// RDBアクセッサにTransactionをセット
	tm.rdbAccessor.SetTransaction(tm.tx)
	// サービスの実行
	result, err := serviceFunc()
	// RDBトランザクション終了
	err = tm.endTransaction(err)
	if err != nil {
		return nil, err
	}
	return result, nil

}

// rdbConnectは、RDBに接続します。
func (tm *defaultTransactionManager) rdbConnect() (*sql.DB, error) {
	// TODO: IAM認証でのDB接続の実装
	// https://qiita.com/k-sasaki-hisys-biz/items/12f680f9a97998322cc0#5-lambda%E3%81%AE%E4%BD%9C%E6%88%90
	// https://docs.aws.amazon.com/ja_jp/AmazonRDS/latest/UserGuide/UsingWithRDS.IAMDBAuth.Connecting.Go.html#UsingWithRDS.IAMDBAuth.Connecting.GoV2

	// AppConfig/SecretsManagerを利用してDB接続情報を取得する実装例
	// DBの認証情報は、SecretsManagerに管理されたものから取得されるが
	// AppConfigを用いており、AppConfigAgentによりキャッシュされたものを取得するので
	// APIのスロットリングの問題を防止できている

	// RDBに作成したユーザ名
	rdbUser, found := tm.config.GetWithContains(RDB_USER_NAME)
	if !found {
		return nil, errors.New("RDB_USER_NAMEが設定されていません")
	}
	tm.log.Debug("RDB_USER_NAME: %s", rdbUser)
	// RDBユーザのパスワード
	rdbPassword, found := tm.config.GetWithContains(RDB_PASSWORD_NAME)
	if !found {
		return nil, errors.New("RDB_PASSWORD_NAMEが設定されていません")
	}
	tm.log.Debug("RDB_PASSWORD_NAME: %s", rdbPassword)
	// RDS Proxyのエンドポイント
	rdbEndpoint, found := tm.config.GetWithContains(RDB_ENDPOINT_NAME)
	if !found {
		return nil, errors.New("RDB_ENDPOINT_NAMEが設定されていません")
	}
	// RDS Proxyのポート
	rdbPort, found := tm.config.GetWithContains(RDB_PORT_NAME)
	if !found {
		return nil, errors.New("RDB_PORT_NAMEが設定されていません")
	}
	// DB名
	rdbName, found := tm.config.GetWithContains(RDB_DB_NAME_NAME)
	if !found {
		return nil, errors.New("RDB_DB_NAME_NAMEが設定されていません")
	}
	// SSLMode
	rdbSslMode := tm.config.Get(RDB_SSL_MODE_NAME, "require")
	// X-Rayを使ったDB接続をすると、プリペアドステートメントを使用していなくても、RDS Proxyでのピン留めが起きてしまう
	// ただし、ピン留めは短時間のため影響は少ない
	// X-RayのSQLトレースに対応したDB接続の取得
	connectStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		rdbUser,
		rdbPassword,
		rdbEndpoint,
		rdbPort,
		rdbName,
		rdbSslMode)
	db, err := xray.SQLContext("postgres", connectStr)

	// X-Rayを使わない場合のDB接続取得の実装例
	/*
		connectStr := fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			rdbEndpoint,
			rdbPort,
			rdbUser,
			rdbPassword,
			rdbName)
		db, err := sql.Open("postgres", connectStr)*/

	if err != nil {
		return nil, errors.WithStack(err)
	}
	return db, nil
}

// startTransaction はトランザクションを開始します。
func (tm *defaultTransactionManager) startTransaction(db *sql.DB) (*sql.Tx, error) {
	tm.log.Debug("トランザクション開始")
	tx, err := db.BeginTx(apcontext.Context, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return tx, nil
}

// endTransaction は、トランザクションを
func (tm *defaultTransactionManager) endTransaction(err error) error {
	if err != nil {
		// トランザクションロールバック
		tm.log.Debug("トランザクションロールバック")
		err2 := tm.tx.Rollback()
		if err2 != nil {
			tm.log.Debug("トランザクションロールバックに失敗")
			//元のエラー、ロールバックに失敗したエラーまとめて返却する
			return errors.Join(err, err2)
		}
		// ロールバックに成功したら元のエラーオブジェクトを返却
		return err
	}
	// トランザクションコミット
	tm.log.Debug("トランザクションコミット")
	return tm.tx.Commit()
}
