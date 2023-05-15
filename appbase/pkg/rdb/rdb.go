package rdb

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

var (
	// RDBに作成したユーザ名
	rdbUser = os.Getenv("RDB_USER")
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
	connectStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		rdbEndpoint,
		rdbPort,
		rdbUser,
		rdbPassword,
		rdbName)
	db, err := sql.Open("postgres", connectStr)
	if err != nil {
		panic(err.Error())
	}
	return db, nil
}
