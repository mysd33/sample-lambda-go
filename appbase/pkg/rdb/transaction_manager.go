/*
rdb パッケージは、RDBアクセスに関する機能を提供するパッケージです。
*/
package rdb

import (
	"database/sql"
	"fmt"
	"net/url"

	"example.com/appbase/pkg/apcontext"
	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/domain"
	"example.com/appbase/pkg/logging"
	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/cockroachdb/errors"

	_ "github.com/lib/pq"
)

const (
	// TODO: SecretsManagerのシークレット名を含むので、定数ではなく設定変更できるようにする。
	RDB_USERNAME_NAME = "rds_smconfig_username"
	RDB_PASSWORD_NAME = "rds_smconfig_password"

	RDB_ENDPOINT_NAME = "RDB_ENDPOINT"
	RDB_PORT_NAME     = "RDB_PORT"
	RDB_DBNAME_NAME   = "RDB_DB_NAME"
	RDB_SSL_MODE_NAME = "RDB_SSL_MODE"
)

// TransactionManager はトランザクションを管理するインタフェースです
type TransactionManager interface {
	// ExecuteTransaction は、Serviceの関数serviceFuncの実行前後でDynamoDBトランザクション実行します。
	ExecuteTransaction(serviceFunc domain.ServiceFunc) (any, error)
}

// NewTransactionManager は、TransactionManagerを作成します
func NewTransactionManager(logger logging.Logger, config config.Config, rdbAccessor RDBAccessor) TransactionManager {
	return &defaultTransactionManager{logger: logger, config: config, rdbAccessor: rdbAccessor}
}

// defaultTransactionManager は、TransactionManagerを実装する構造体です。
type defaultTransactionManager struct {
	logger      logging.Logger
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

	// RDBユーザ名
	rdbUser, found := tm.config.GetWithContains(RDB_USERNAME_NAME)
	if !found {
		return nil, errors.Newf("%sが設定されていません", RDB_USERNAME_NAME)
	}
	// RDBユーザのパスワード
	rdbPassword, found := tm.config.GetWithContains(RDB_PASSWORD_NAME)
	if !found {
		return nil, errors.Newf("%sが設定されていません", RDB_PASSWORD_NAME)
	}
	// RDS Proxyのエンドポイント
	rdbEndpoint, found := tm.config.GetWithContains(RDB_ENDPOINT_NAME)
	if !found {
		return nil, errors.Newf("%sが設定されていません", RDB_ENDPOINT_NAME)
	}
	// RDS Proxyのポート（デフォルトは5432番）
	rdbPort := tm.config.Get(RDB_PORT_NAME, "5432")
	// DB名
	rdbName, found := tm.config.GetWithContains(RDB_DBNAME_NAME)
	if !found {
		return nil, errors.Newf("%sが設定されていません", RDB_DBNAME_NAME)
	}
	// SSLMode
	rdbSslMode := tm.config.Get(RDB_SSL_MODE_NAME, "require")
	// X-Rayを使ったDB接続をすると、プリペアドステートメントを使用していなくても、RDS Proxyでのピン留めが起きてしまう
	// ただし、ピン留めは短時間のため影響は少ない
	// X-RayのSQLトレースに対応したDB接続の取得
	connectStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		rdbUser,
		// パスワードに「[」等の特殊文字が入っている場合の対処
		url.PathEscape(rdbPassword),
		rdbEndpoint,
		rdbPort,
		rdbName,
		rdbSslMode)
	tm.logger.Debug("接続文字列: %s", connectStr)
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
		tm.logger.Debug("接続文字列: %s", connectStr)
		db, err := sql.Open("postgres", connectStr)*/

	if err != nil {
		return nil, errors.WithStack(err)
	}
	return db, nil
}

// startTransaction はトランザクションを開始します。
func (tm *defaultTransactionManager) startTransaction(db *sql.DB) (*sql.Tx, error) {
	tm.logger.Debug("トランザクション開始")
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
		tm.logger.Debug("トランザクションロールバック")
		err2 := tm.tx.Rollback()
		if err2 != nil {
			tm.logger.Debug("トランザクションロールバックに失敗")
			//元のエラー、ロールバックに失敗したエラーまとめて返却する
			return errors.Join(err, err2)
		}
		// ロールバックに成功したら元のエラーオブジェクトを返却
		return err
	}
	// トランザクションコミット
	tm.logger.Debug("トランザクションコミット")
	return tm.tx.Commit()
}
