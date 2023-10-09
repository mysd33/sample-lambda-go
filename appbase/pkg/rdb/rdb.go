package rdb

import (
	"database/sql"
	"fmt"
	"os"

	"example.com/appbase/pkg/apcontext"
	"example.com/appbase/pkg/domain"
	"github.com/aws/aws-xray-sdk-go/xray"
	_ "github.com/lib/pq"
)

var (
	// RDBに作成したユーザ名
	rdbUser = os.Getenv("RDB_USER")
	// TODO: IAM認証でトークン取得
	// RDBユーザのパスワード
	rdbPassword = os.Getenv("RDB_PASSWORD")
	// RDS Proxyのエンドポイント
	rdbEndpoint = os.Getenv("RDB_ENDPOINT")
	// RDS Proxyのポート
	rdbPort = os.Getenv("RDB_PORT")
	// DB名
	rdbName = os.Getenv("RDB_DB_NAME")
)

func RDSConnect() (*sql.DB, error) {
	// X-Rayを使ったDB接続をすると、プリペアドステートメントを使用していなくても、RDS Proxyでのピン留めが起きてしまう
	// ただし、ピン留めは短時間のため影響は少ない
	// X-RayのSQLトレースに対応したDB接続の取得
	connectStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		rdbUser,
		rdbPassword,
		rdbEndpoint,
		rdbPort,
		rdbName)
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
		panic(err.Error())
	}
	return db, nil
}

func HandleTransaction(serviceFunc domain.ServiceFunc) (interface{}, error) {
	// RDBトランザクション開始
	err := startTransaction()
	if err != nil {
		return nil, err
	}
	// サービスの実行
	result, err := serviceFunc()
	// RDBトランザクション終了
	err = endTransaction(err)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func startTransaction() error {
	// RDBコネクションの確立
	db, err := RDSConnect()
	if err != nil {
		return err
	}
	// DBコネクションをコンテキスト領域に格納
	apcontext.DB = db
	// トランザクション開始
	tx, err := apcontext.DB.BeginTx(apcontext.Context, nil)
	if err != nil {
		return err
	}
	// トランザクションをコンテキスト領域に格納
	apcontext.Tx = tx
	return nil
}

func endTransaction(err error) error {
	// 終了時にRDBコネクションの切断
	db := apcontext.DB
	if db == nil {
		return nil
	}
	defer db.Close()
	// トランザクション取得
	tx := apcontext.Tx
	if tx == nil {
		return nil
	}
	if err != nil {
		// トランザクションロールバック
		err2 := tx.Rollback()
		if err2 != nil {
			//TODO: ロールバックに失敗した旨と、元のエラーをログ出力
			return err2
		}
		// ロールバックに成功したら元のエラーオブジェクトを返却
		return err
	}
	// トランザクションコミット
	return tx.Commit()
}
