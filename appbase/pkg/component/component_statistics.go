/*
component パッケージはフレームワークのコンポーネントのインスタンスをDIし管理するパッケージです。
*/
package component

import (
	"example.com/appbase/pkg/async"
	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/date"
	"example.com/appbase/pkg/handler"
	"example.com/appbase/pkg/httpclient"
	"example.com/appbase/pkg/id"
	"example.com/appbase/pkg/logging"
	"example.com/appbase/pkg/message"
	"example.com/appbase/pkg/objectstorage"
	"example.com/appbase/pkg/rdb"
	"example.com/appbase/pkg/transaction"
	"example.com/appbase/pkg/validator"
)

// NewStatisticsApplicationContext は、ログ分割・振分等の統計分析のログ形式変換向けの軽量なApplicationContextを作成します。
// 重量級のAP基盤機能のインスタンス化を行わず、必要最低限の軽量なAP基盤機能のみを提供します。
// Lambdaコールドスタート時の初期化処理時間を極力短くすることで、コスト削減、バーストリミットのリスクを減らし
// ライトウエイトなApplicationContextに差し替えるために利用します。
// 提供するインタフェースは、以下のみです。
// - GetConfig
// - GetMessageSource
// - GetLogger
// - GetDateManager
// - GetInterceptor
// - GetSimpleLambdaHandler
func NewStatisticsApplicationContext() ApplicationContext {
	// 各種AP基盤の構造体を作成
	messageSource := createMessageSource()
	logger := createLogger(messageSource)
	config := createConfig(logger)
	dateManager := createDateManager(config, logger)
	interceptor := createHanderInterceptor(config, logger)
	simpleLambdaHandler := createSimpleLambdaHandler(config, logger)

	return &statisticsApplicationContext{
		config:              config,
		messageSource:       messageSource,
		logger:              logger,
		dateManager:         dateManager,
		interceptor:         interceptor,
		simpleLambdaHandler: simpleLambdaHandler,
	}
}

// statisticsApplicationContext は統計分析処理方式向けの軽量なApplicationContext実装です。
type statisticsApplicationContext struct {
	config              config.Config
	messageSource       message.MessageSource
	logger              logging.Logger
	dateManager         date.DateManager
	interceptor         handler.HandlerInterceptor
	simpleLambdaHandler *handler.SimpleLambdaHandler
}

// 本ApplicatonContextを利用するLambdaでは不要な機能とし未実装のため、panicします。
func (l *statisticsApplicationContext) GetAPILambdaHandler() *handler.APILambdaHandler {
	panic("unimplemented")
}

// 本ApplicatonContextを利用するLambdaでは不要な機能とし未実装のため、panicします。
func (l *statisticsApplicationContext) GetAsyncLambdaHandler() *handler.AsyncLambdaHandler {
	panic("unimplemented")
}

// GetConfig implements ApplicationContext.
func (l *statisticsApplicationContext) GetConfig() config.Config {
	return l.config
}

// GetDateManager implements ApplicationContext.
func (l *statisticsApplicationContext) GetDateManager() date.DateManager {
	return l.dateManager
}

// 本ApplicatonContextを利用するLambdaでは不要な機能とし未実装のため、panicします。
func (l *statisticsApplicationContext) GetDynamoDBAccessor() transaction.TransactionalDynamoDBAccessor {
	panic("unimplemented")
}

// 本ApplicatonContextを利用するLambdaでは不要な機能とし未実装のため、panicします。
func (l *statisticsApplicationContext) GetDynamoDBTemplate() transaction.TransactionalDynamoDBTemplate {
	panic("unimplemented")
}

// 本ApplicatonContextを利用するLambdaでは不要な機能とし未実装のため、panicします。
func (l *statisticsApplicationContext) GetDynamoDBTransactionManager() transaction.TransactionManager {
	panic("unimplemented")
}

// 本ApplicatonContextを利用するLambdaでは不要な機能とし未実装のため、panicします。
func (l *statisticsApplicationContext) GetDynamoDBTransactionManagerForDBOnly() transaction.TransactionManager {
	panic("unimplemented")
}

// 本ApplicatonContextを利用するLambdaでは不要な機能とし未実装のため、panicします。
func (l *statisticsApplicationContext) GetHttpClient() httpclient.HttpClient {
	panic("unimplemented")
}

// 本ApplicatonContextを利用するLambdaでは不要な機能とし未実装のため、panicします。
func (l *statisticsApplicationContext) GetIDGenerator() id.IDGenerator {
	panic("unimplemented")
}

// GetInterceptor implements ApplicationContext.
func (l *statisticsApplicationContext) GetInterceptor() handler.HandlerInterceptor {
	return l.interceptor
}

// GetLogger implements ApplicationContext.
func (l *statisticsApplicationContext) GetLogger() logging.Logger {
	return l.logger
}

// GetMessageSource implements ApplicationContext.
func (l *statisticsApplicationContext) GetMessageSource() message.MessageSource {
	return l.messageSource
}

// 本ApplicatonContextを利用するLambdaでは不要な機能とし未実装のため、panicします。
func (l *statisticsApplicationContext) GetObjectStorageAccessor() objectstorage.ObjectStorageAccessor {
	panic("unimplemented")
}

// 本ApplicatonContextを利用するLambdaでは不要な機能とし未実装のため、panicします。
func (l *statisticsApplicationContext) GetRDBAccessor() rdb.RDBAccessor {
	panic("unimplemented")
}

// 本ApplicatonContextを利用するLambdaでは不要な機能とし未実装のため、panicします。
func (l *statisticsApplicationContext) GetRDBTransactionManager() rdb.TransactionManager {
	panic("unimplemented")
}

// 本ApplicatonContextを利用するLambdaでは不要な機能とし未実装のため、panicします。
func (l *statisticsApplicationContext) GetSQSAccessor() transaction.TransactionalSQSAccessor {
	panic("unimplemented")
}

// 本ApplicatonContextを利用するLambdaでは不要な機能とし未実装のため、panicします。
func (l *statisticsApplicationContext) GetSQSTemplate() async.SQSTemplate {
	panic("unimplemented")
}

// GetSimpleLambdaHandler implements ApplicationContext.
func (l *statisticsApplicationContext) GetSimpleLambdaHandler() *handler.SimpleLambdaHandler {
	return l.simpleLambdaHandler
}

// 本ApplicatonContextを利用するLambdaでは不要な機能とし未実装のため、panicします。
func (l *statisticsApplicationContext) GetValidationManager() validator.ValidationManager {
	panic("unimplemented")
}
