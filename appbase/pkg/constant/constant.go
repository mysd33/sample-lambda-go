/*
constantパッケージ は定数を扱うパッケージです
*/
package constant

import "example.com/appbase/pkg/apcontext"

const (
	RDB_USER_NAME                = "RDB_USER"
	RDB_PASSWORD_NAME            = "RDB_PASSWORD"
	RDB_ENDPOINT_NAME            = "RDB_ENDPOINT"
	RDB_PORT_NAME                = "RDB_PORT"
	RDB_DB_NAME_NAME             = "RDB_DB_NAME"
	RDB_SSL_MODE_NAME            = "RDB_SSL_MODE"
	DYNAMODB_LOCAL_ENDPOINT_NAME = "DYNAMODB_LOCAL_ENDPOINT"
	SQS_LOCAL_ENDPOINT_NAME      = "SQS_LOCAL_ENDPOINT"
	APPCONFIG_EXTENSION_URL_NAME = "APPCONFIG_EXTENSION_URL"
	LOG_LEVEL_NAME               = "LOG_LEVEL"
	GIN_DEBUG_NAME               = "GIN_DEBUG"
	IS_TABLE_CHECK_NAME          = "is_table_check"
	DELETE_TIME_NAME             = "delete_time"
	MESSAGE_DEDUPLICATION_ID     = "message_deduplication_id"
	ASYNC_HANDLER_INFO_CTX_KEY   = apcontext.ContextKey("ASYNC_HANDLER_INFO")
)
