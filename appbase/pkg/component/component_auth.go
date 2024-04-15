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

// NewAuthorizerApplicationContext は、API認可処理向けに軽量なApplicationContextを作成します。
// 重量級のAP基盤機能のインスタンス化を行わず、必要最低限の軽量なAP基盤機能のみを提供します。
// Lambdaコールドスタート時の初期化処理時間を極力短くすることで、コスト削減、バーストリミットのリスクを減らし
// ライトウエイトなApplicationContextに差し替えるために利用します。
// 提供するインタフェースは、以下のみです。
// - GetConfig
// - GetMessageSource
// - GetLogger
// - GetDateManager
// - GetDynamoDBAccessor
// - GetDynamoDBTemplate
// - GetDynamoDBTransactionManager（ただし、DBのみのトランザクション管理機能を返却します）
// - GetHttpClient
// - GetInterceptor
// - GetSimpleLambdaHandler
func NewAuthorizerApplicationContext() ApplicationContext {
	// 各種AP基盤の構造体を作成
	messageSource := createMessageSource()
	logger := createLogger(messageSource)
	config := createConfig(logger)
	dateManager := createDateManager(config, logger)
	dynamodbAccessor := createTransactionalDynamoDBAccessor(logger, config)
	dynamoDBTempalte := createDynamoDBTemplate(logger, dynamodbAccessor)
	queueMessageItemRepository := createQueueMessageItemRepository(config, logger, dynamoDBTempalte)
	messageRegisterer := createMessageRegisterer(queueMessageItemRepository)
	dynamoDBTransactionManagerForDBOnly := createDynamoDBTransactionManagerForDBOnly(logger, dynamodbAccessor, messageRegisterer)
	httpclient := createHttpClient(config, logger)
	interceptor := createHanderInterceptor(config, logger)
	simpleLambdaHandler := createSimpleLambdaHandler(config, logger)

	return &authorizerApplicationContext{
		config:                              config,
		messageSource:                       messageSource,
		logger:                              logger,
		dateManager:                         dateManager,
		dynamoDBAccessor:                    dynamodbAccessor,
		dynamoDBTransactionManagerForDBOnly: dynamoDBTransactionManagerForDBOnly,
		dynamodbTempalte:                    dynamoDBTempalte,
		httpClient:                          httpclient,
		interceptor:                         interceptor,
		simpleLambdaHandler:                 simpleLambdaHandler,
	}
}

type authorizerApplicationContext struct {
	config                              config.Config
	messageSource                       message.MessageSource
	logger                              logging.Logger
	dateManager                         date.DateManager
	dynamoDBAccessor                    transaction.TransactionalDynamoDBAccessor
	dynamoDBTransactionManagerForDBOnly transaction.TransactionManager
	dynamodbTempalte                    transaction.TransactionalDynamoDBTemplate
	httpClient                          httpclient.HttpClient
	interceptor                         handler.HandlerInterceptor
	simpleLambdaHandler                 *handler.SimpleLambdaHandler
}

// 本ApplicatonContextを利用するLambdaでは不要な機能とし未実装のため、panicします。
func (a *authorizerApplicationContext) GetAPILambdaHandler() *handler.APILambdaHandler {
	panic("unimplemented")
}

// 本ApplicatonContextを利用するLambdaでは不要な機能とし未実装のため、panicします。
func (a *authorizerApplicationContext) GetAsyncLambdaHandler() *handler.AsyncLambdaHandler {
	panic("unimplemented")
}

// GetConfig implements ApplicationContext.
func (a *authorizerApplicationContext) GetConfig() config.Config {
	return a.config
}

// GetDateManager implements ApplicationContext.
func (a *authorizerApplicationContext) GetDateManager() date.DateManager {
	return a.dateManager
}

// GetDynamoDBAccessor implements ApplicationContext.
func (a *authorizerApplicationContext) GetDynamoDBAccessor() transaction.TransactionalDynamoDBAccessor {
	return a.dynamoDBAccessor
}

// GetDynamoDBTemplate implements ApplicationContext.
func (a *authorizerApplicationContext) GetDynamoDBTemplate() transaction.TransactionalDynamoDBTemplate {
	return a.dynamodbTempalte
}

// GetDynamoDBTransactionManager implements ApplicationContext.
// GetDynamoDBTransactionManagerもSQSのトランザクションを管理をサポートせず、DBのみのトランザクション管理機能を返却します。
func (a *authorizerApplicationContext) GetDynamoDBTransactionManager() transaction.TransactionManager {
	return a.GetDynamoDBTransactionManagerForDBOnly()
}

// GetDynamoDBTransactionManagerForDBOnly implements ApplicationContext.
func (a *authorizerApplicationContext) GetDynamoDBTransactionManagerForDBOnly() transaction.TransactionManager {
	return a.dynamoDBTransactionManagerForDBOnly
}

// GetHttpClient implements ApplicationContext.
func (a *authorizerApplicationContext) GetHttpClient() httpclient.HttpClient {
	return a.httpClient
}

// 本ApplicatonContextを利用するLambdaでは不要な機能とし未実装のため、panicします。
func (a *authorizerApplicationContext) GetIDGenerator() id.IDGenerator {
	panic("unimplemented")
}

// GetInterceptor implements ApplicationContext.
func (a *authorizerApplicationContext) GetInterceptor() handler.HandlerInterceptor {
	return a.interceptor
}

// GetLogger implements ApplicationContext.
func (a *authorizerApplicationContext) GetLogger() logging.Logger {
	return a.logger
}

// GetMessageSource implements ApplicationContext.
func (a *authorizerApplicationContext) GetMessageSource() message.MessageSource {
	return a.messageSource
}

// 本ApplicatonContextを利用するLambdaでは不要な機能とし未実装のため、panicします。
func (a *authorizerApplicationContext) GetObjectStorageAccessor() objectstorage.ObjectStorageAccessor {
	panic("unimplemented")
}

// 本ApplicatonContextを利用するLambdaでは不要な機能とし未実装のため、panicします。
func (a *authorizerApplicationContext) GetRDBAccessor() rdb.RDBAccessor {
	panic("unimplemented")
}

// 本ApplicatonContextを利用するLambdaでは不要な機能とし未実装のため、panicします。
func (a *authorizerApplicationContext) GetRDBTransactionManager() rdb.TransactionManager {
	panic("unimplemented")
}

// 本ApplicatonContextを利用するLambdaでは不要な機能とし未実装のため、panicします。
func (a *authorizerApplicationContext) GetSQSAccessor() transaction.TransactionalSQSAccessor {
	panic("unimplemented")
}

// 本ApplicatonContextを利用するLambdaでは不要な機能とし未実装のため、panicします。
func (a *authorizerApplicationContext) GetSQSTemplate() async.SQSTemplate {
	panic("unimplemented")
}

// GetSimpleLambdaHandler implements ApplicationContext.
func (a *authorizerApplicationContext) GetSimpleLambdaHandler() *handler.SimpleLambdaHandler {
	return a.simpleLambdaHandler
}

// 本ApplicatonContextを利用するLambdaでは不要な機能とし未実装のため、panicします。
func (a *authorizerApplicationContext) GetValidationManager() validator.ValidationManager {
	panic("unimplemented")
}
