/*
constantパッケージ は、パッケージ間で横断的に扱う定数を管理するパッケージです
*/
package constant

import "example.com/appbase/pkg/apcontext"

const (
	QUEUE_MESSAGE_NEEDS_TABLE_CHECK_NAME = "needs_table_check"
	QUEUE_MESSAGE_DELETE_TIME_NAME       = "delete_time"
	QUEUE_MESSAGE_STATUS                 = "status"
	QUEUE_MESSAGE_STATUS_COMPLETE        = "complete"
	ASYNC_HANDLER_INFO_CTX_KEY           = apcontext.ContextKey("ASYNC_HANDLER_INFO")
)
