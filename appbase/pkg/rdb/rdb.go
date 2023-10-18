/*
rdb パッケージは、RDBアクセスに関する機能を提供するパッケージです。
*/
package rdb

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	"example.com/appbase/pkg/apcontext"
	"example.com/appbase/pkg/domain"
	"github.com/aws/aws-xray-sdk-go/xray"
	_ "github.com/lib/pq"
)

var (
	// DBコネクション
	DB *sql.DB
	// RDBトランザクション
	Tx *sql.Tx
	// RDBに作成したユーザ名
	rdbUser = os.Getenv("RDB_USER")
	// TODO: IAM認証でトークン取得による方法（スロットリングによる性能問題の恐れもあるので一旦様子見）
	// RDBユーザのパスワード
	rdbPassword = os.Getenv("RDB_PASSWORD")
	// RDS Proxyのエンドポイント
	rdbEndpoint = os.Getenv("RDB_ENDPOINT")
	// RDS Proxyのポート
	rdbPort = os.Getenv("RDB_PORT")
	// DB名
	rdbName = os.Getenv("RDB_DB_NAME")
	// SSLMode
	rdbSslMode = os.Getenv("RDB_SSL_MODE")
)

// ExecuteTransactionは、Serviceの関数serviceFuncの実行前後で、RDBトランザクションを実行します。
func ExecuteTransaction(serviceFunc domain.ServiceFunc) (interface{}, error) {
	// RDBコネクションの確立
	db, err := rdbConnect()
	if err != nil {
		return nil, err
	}
	// 終了時にRDBコネクションの切断
	defer db.Close()
	// RDBトランザクション開始
	tx, err := startTransaction(db)
	if err != nil {
		return nil, err
	}
	// トランザクションをコンテキスト領域に格納
	Tx = tx
	// サービスの実行
	result, err := serviceFunc()
	// RDBトランザクション終了
	err = endTransaction(tx, err)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// rdbConnectは、RDBに接続します。
func rdbConnect() (*sql.DB, error) {
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
		return nil, err
	}
	// DBコネクションをコンテキスト領域に格納
	DB = db
	return db, nil
}

func startTransaction(db *sql.DB) (*sql.Tx, error) {
	// トランザクション開始
	tx, err := db.BeginTx(apcontext.Context, nil)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func endTransaction(tx *sql.Tx, err error) error {
	// トランザクション取得
	if tx == nil {
		return nil
	}
	if err != nil {
		// トランザクションロールバック
		err2 := tx.Rollback()
		if err2 != nil {
			//元のエラー、ロールバックに失敗したエラーまとめて返却する
			return errors.Join(err, err2)
		}
		// ロールバックに成功したら元のエラーオブジェクトを返却
		return err
	}
	// トランザクションコミット
	return tx.Commit()
}
