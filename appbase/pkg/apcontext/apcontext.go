package apcontext

import (
	"context"
	"database/sql"
)

var Context context.Context

var DB *sql.DB

var Tx *sql.Tx
