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

// NewLightWeightApplicationContext は軽量なApplicationContextを作成します。
// 重量級のAP基盤機能のインスタンス化を行わず、軽量なAP基盤機能のみを提供します。
// Lambdaコールドスタート時の初期化処理時間を極力短くすることで、コスト削減、バーストリミットのリスクを減らし
// ライトウエイトなApplicationContextに差し替えたい場合に利用します。
// 提供するインタフェースは、以下のみです。
// - GetIDGenerator
// - GetConfig
// - GetMessageSource
// - GetLogger
// - GetDateManager
// - GetHttpClient
// - GetInterceptor
// - GetSimpleLambdaHandler
func NewLightWeightApplicationContext() ApplicationContext {
	// 各種AP基盤の構造体を作成
	id := createIDGenerator()
	messageSource := createMessageSource()
	logger := createLogger(messageSource)
	config := createConfig(logger)
	dateManager := createDateManager(config, logger)
	httpclient := createHttpClient(config, logger)
	interceptor := createHanderInterceptor(config, logger)
	simpleLambdaHandler := createSimpleLambdaHandler(config, logger)

	return &lightWeightApplicationContext{
		id:                  id,
		config:              config,
		messageSource:       messageSource,
		logger:              logger,
		dateManager:         dateManager,
		httpClient:          httpclient,
		interceptor:         interceptor,
		simpleLambdaHandler: simpleLambdaHandler,
	}
}

// lightWeightApplicationContext は軽量なApplicationContext実装です。
type lightWeightApplicationContext struct {
	id                  id.IDGenerator
	config              config.Config
	messageSource       message.MessageSource
	logger              logging.Logger
	dateManager         date.DateManager
	httpClient          httpclient.HttpClient
	interceptor         handler.HandlerInterceptor
	simpleLambdaHandler *handler.SimpleLambdaHandler
}

// 本ApplicatonContextを利用するLambdaでは不要な機能とし未実装のため、panicします。
func (l *lightWeightApplicationContext) GetAPILambdaHandler() *handler.APILambdaHandler {
	panic("unimplemented")
}

// 本ApplicatonContextを利用するLambdaでは不要な機能とし未実装のため、panicします。
func (l *lightWeightApplicationContext) GetAsyncLambdaHandler() *handler.AsyncLambdaHandler {
	panic("unimplemented")
}

// GetConfig implements ApplicationContext.
func (l *lightWeightApplicationContext) GetConfig() config.Config {
	return l.config
}

// GetDateManager implements ApplicationContext.
func (l *lightWeightApplicationContext) GetDateManager() date.DateManager {
	return l.dateManager
}

// 本ApplicatonContextを利用するLambdaでは不要な機能とし未実装のため、panicします。
func (l *lightWeightApplicationContext) GetDynamoDBAccessor() transaction.TransactionalDynamoDBAccessor {
	panic("unimplemented")
}

// 本ApplicatonContextを利用するLambdaでは不要な機能とし未実装のため、panicします。
func (l *lightWeightApplicationContext) GetDynamoDBTemplate() transaction.TransactionalDynamoDBTemplate {
	panic("unimplemented")
}

// 本ApplicatonContextを利用するLambdaでは不要な機能とし未実装のため、panicします。
func (l *lightWeightApplicationContext) GetDynamoDBTransactionManager() transaction.TransactionManager {
	panic("unimplemented")
}

// 本ApplicatonContextを利用するLambdaでは不要な機能とし未実装のため、panicします。
func (l *lightWeightApplicationContext) GetDynamoDBTransactionManagerForDBOnly() transaction.TransactionManager {
	panic("unimplemented")
}

// 本ApplicatonContextを利用するLambdaでは不要な機能とし未実装のため、panicします。
func (l *lightWeightApplicationContext) GetHttpClient() httpclient.HttpClient {
	return l.httpClient
}

// GetIDGenerator implements ApplicationContext.
func (l *lightWeightApplicationContext) GetIDGenerator() id.IDGenerator {
	return l.id
}

// GetInterceptor implements ApplicationContext.
func (l *lightWeightApplicationContext) GetInterceptor() handler.HandlerInterceptor {
	return l.interceptor
}

// GetLogger implements ApplicationContext.
func (l *lightWeightApplicationContext) GetLogger() logging.Logger {
	return l.logger
}

// GetMessageSource implements ApplicationContext.
func (l *lightWeightApplicationContext) GetMessageSource() message.MessageSource {
	return l.messageSource
}

// 本ApplicatonContextを利用するLambdaでは不要な機能とし未実装のため、panicします。
func (l *lightWeightApplicationContext) GetObjectStorageAccessor() objectstorage.ObjectStorageAccessor {
	panic("unimplemented")
}

// 本ApplicatonContextを利用するLambdaでは不要な機能とし未実装のため、panicします。
func (l *lightWeightApplicationContext) GetRDBAccessor() rdb.RDBAccessor {
	panic("unimplemented")
}

// 本ApplicatonContextを利用するLambdaでは不要な機能とし未実装のため、panicします。
func (l *lightWeightApplicationContext) GetRDBTransactionManager() rdb.TransactionManager {
	panic("unimplemented")
}

// 本ApplicatonContextを利用するLambdaでは不要な機能とし未実装のため、panicします。
func (l *lightWeightApplicationContext) GetSQSAccessor() transaction.TransactionalSQSAccessor {
	panic("unimplemented")
}

// 本ApplicatonContextを利用するLambdaでは不要な機能とし未実装のため、panicします。
func (l *lightWeightApplicationContext) GetSQSTemplate() async.SQSTemplate {
	panic("unimplemented")
}

// GetSimpleLambdaHandler implements ApplicationContext.
func (l *lightWeightApplicationContext) GetSimpleLambdaHandler() *handler.SimpleLambdaHandler {
	return l.simpleLambdaHandler
}

// 本ApplicatonContextを利用するLambdaでは不要な機能とし未実装のため、panicします。
func (l *lightWeightApplicationContext) GetValidationManager() validator.ValidationManager {
	panic("unimplemented")
}
