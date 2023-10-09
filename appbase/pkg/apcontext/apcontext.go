package apcontext

import (
	"context"
	"database/sql"
)

var Context context.Context

// TODO:相互参照になってしまうのでrdbパッケージへ分離
var DB *sql.DB
var Tx *sql.Tx
