/*
rdb パッケージは、RDBアクセスに関する機能を提供するパッケージです。
*/
package rdb

import (
	"database/sql"
)

type RDBAccessor interface {
	// トランザクションを取得する
	GetTransaction() *sql.Tx
	SetTransaction(tx *sql.Tx)
}

func NewRDBAccessor() RDBAccessor {
	return &defaultRDBAccessor{}
}

type defaultRDBAccessor struct {
	tx *sql.Tx
}

// GetTransaction implements RDBAccessor.
func (ra *defaultRDBAccessor) GetTransaction() *sql.Tx {
	return ra.tx
}

// SetTransaction implements RDBAccessor.
func (ra *defaultRDBAccessor) SetTransaction(tx *sql.Tx) {
	ra.tx = tx
}
