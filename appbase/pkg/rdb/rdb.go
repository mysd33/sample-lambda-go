package rdb

import (
	"database/sql"
	"fmt"
	"os"

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
