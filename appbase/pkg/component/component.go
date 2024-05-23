/*
component パッケージはフレームワークのコンポーネントのインスタンスをDIし管理するパッケージです。
*/
package component

import (
	"example.com/appbase/pkg/api"
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

// ApplicationContext は、フレームワークのコンポーネントを管理するインタフェースです。
type ApplicationContext interface {
	// GetIDGenerator は、ID生成機能のインタフェースIDGeneratorを取得します。
	GetIDGenerator() id.IDGenerator
	// GetMessageSource は、メッセージ管理機能のインタフェースMessageSourceを取得します。
	GetMessageSource() message.MessageSource
	// GetLogger は、ロギング機能のインタフェースLoggerを取得します。
	GetLogger() logging.Logger
	// GetConfig は、プロパティ管理機能のインタフェースConfigを取得します。
	GetConfig() config.Config
	// GetDynamoDBAccessor は、DynamoDBアクセス機能のインタフェースTransactionalDynamoDBAccessorを取得します。
	GetDynamoDBAccessor() transaction.TransactionalDynamoDBAccessor
	// GetDynamoDBTransactionManager は、DynamoDBトランザクション管理機能のインタフェースTransactionManagerを取得します。
	GetDynamoDBTransactionManager() transaction.TransactionManager
	// GetDynamoDBTransactionManagerForDBOnly は、DynamoDBトランザクション管理機能のインタフェースTransactionManagerを取得します。
	// DynamoDBのみのトランザクションのみを行う場合に利用します。
	GetDynamoDBTransactionManagerForDBOnly() transaction.TransactionManager
	// GetDynamoDBTemplate は、トランザクション対応のDynamoDBアクセス機能の汎用インタフェースTransactionalDynamoDBTemplateを取得します。
	GetDynamoDBTemplate() transaction.TransactionalDynamoDBTemplate
	// GetSQSAccessor は、トランザクション対応の非同期実行依頼機能のインタフェースTransactionalSQSAccessorを取得します。
	GetSQSAccessor() transaction.TransactionalSQSAccessor
	// GetSQSTemplate は、非同期実行依頼機能の汎用インタフェースSQSTemplateを取得します。
	GetSQSTemplate() async.SQSTemplate
	// GetObjectStorageAccessor は、オブジェクトストレージアクセス機能のインタフェースObjectStorageAccessorを取得します。
	GetObjectStorageAccessor() objectstorage.ObjectStorageAccessor
	// GetRDBAccessor は、RDBアクセス機能のインタフェースRDBAccessorを取得します。
	GetRDBAccessor() rdb.RDBAccessor
	// GetRDBTransactionManager は、RDBトランザクション管理機能のインタフェースTransactionManagerを取得します。
	GetRDBTransactionManager() rdb.TransactionManager
	// GetHttpClient は、HTTPクライアント機能のインタフェースHttpClientを取得します。
	GetHttpClient() httpclient.HttpClient
	// GetInterceptor は、集約例外ハンドリング機能のインターセプタのインタフェースHandlerInterceptorを取得します。
	GetInterceptor() handler.HandlerInterceptor
	// GetAPILambdaHandler は、APIトリガのオンラインAP実行制御機能のインタフェースAPILambdaHandlerを取得します。
	GetAPILambdaHandler() *handler.APILambdaHandler
	// GetAsyncLambdaHandler は、SQSトリガの非同期AP実行制御機能のインタフェースAsyncLambdaHandlerを取得します。
	GetAsyncLambdaHandler() *handler.AsyncLambdaHandler
	// GetSimpleLambdaHandler は、その他トリガのAP実行制御機能のインタフェースSimpleLambdaHandlerを取得します。
	GetSimpleLambdaHandler() *handler.SimpleLambdaHandler
	// GetValidationManager は、入力チェック機能のインタフェースValidationManagerを取得します。
	GetValidationManager() validator.ValidationManager
	// GetDateManager は、日付管理機能のインタフェースDateManagerを取得します。
	GetDateManager() date.DateManager
}

// NewApplicationContext は、デフォルトのApplicationContextを作成します。
func NewApplicationContext() ApplicationContext {
	// 各種AP基盤の構造体を作成
	id := createIDGenerator()
	messageSource := createMessageSource()
	logger := createLogger(messageSource)
	config := createConfig(logger)
	dateManager := createDateManager(config, logger)
	apiResponseFormatter := createApiResponseFormatter(logger, messageSource)
	dynamodbAccessor := createTransactionalDynamoDBAccessor(logger, config)
	dynamoDBTempalte := createDynamoDBTemplate(logger, dynamodbAccessor)
	queueMessageItemRepository := createQueueMessageItemRepository(config, logger, dynamoDBTempalte)
	messageRegisterer := createMessageRegisterer(queueMessageItemRepository)
	sqsAccessor := createTransactionalSQSAccessor(logger, config, messageRegisterer)
	sqsTemplate := createSQSTemplate(logger, config, id, sqsAccessor)
	objectStorageAccessor := createObjectStorageAccessor(config, logger)
	dynamoDBTransactionManager := createDynamoDBTransactionManager(logger, dynamodbAccessor, sqsAccessor, messageRegisterer)
	dynamoDBTransactionManagerForDBOnly := createDynamoDBTransactionManagerForDBOnly(logger, dynamodbAccessor, messageRegisterer)
	rdbAccessor := createRDBAccessor()
	rdbTransactionManager := rdb.NewTransactionManager(logger, config, rdbAccessor)
	httpclient := createHttpClient(config, logger)
	interceptor := createHanderInterceptor(config, logger)
	apiLambdaHandler := createAPILambdaHandler(config, logger, messageSource, apiResponseFormatter)
	asyncLambdaHandler := createAsyncLambdaHandler(config, logger, queueMessageItemRepository)
	simpleLambdaHandler := createSimpleLambdaHandler(config, logger)
	validationManager := createValidationManager(logger)

	return &defaultApplicationContext{
		id:                                  id,
		config:                              config,
		messageSource:                       messageSource,
		logger:                              logger,
		dateManager:                         dateManager,
		dynamoDBAccessor:                    dynamodbAccessor,
		dynamoDBTransactionManager:          dynamoDBTransactionManager,
		dynamoDBTransactionManagerForDBOnly: dynamoDBTransactionManagerForDBOnly,
		dynamodbTempalte:                    dynamoDBTempalte,
		sqsAccessor:                         sqsAccessor,
		sqsTemplate:                         sqsTemplate,
		objectStorageAccessor:               objectStorageAccessor,
		rdbAccessor:                         rdbAccessor,
		rdbTransactionManager:               rdbTransactionManager,
		httpClient:                          httpclient,
		interceptor:                         interceptor,
		apiLambdaHandler:                    apiLambdaHandler,
		asyncLambdaHandler:                  asyncLambdaHandler,
		simpleLambdaHandler:                 simpleLambdaHandler,
		validationManager:                   validationManager,
	}
}

type defaultApplicationContext struct {
	id                                  id.IDGenerator
	config                              config.Config
	messageSource                       message.MessageSource
	logger                              logging.Logger
	dateManager                         date.DateManager
	dynamoDBAccessor                    transaction.TransactionalDynamoDBAccessor
	dynamoDBTransactionManager          transaction.TransactionManager
	dynamoDBTransactionManagerForDBOnly transaction.TransactionManager
	dynamodbTempalte                    transaction.TransactionalDynamoDBTemplate
	sqsAccessor                         transaction.TransactionalSQSAccessor
	sqsTemplate                         async.SQSTemplate
	objectStorageAccessor               objectstorage.ObjectStorageAccessor
	rdbAccessor                         rdb.RDBAccessor
	rdbTransactionManager               rdb.TransactionManager
	httpClient                          httpclient.HttpClient
	interceptor                         handler.HandlerInterceptor
	apiLambdaHandler                    *handler.APILambdaHandler
	asyncLambdaHandler                  *handler.AsyncLambdaHandler
	simpleLambdaHandler                 *handler.SimpleLambdaHandler
	validationManager                   validator.ValidationManager
}

// GetIDGenerator implements ApplicationContext.
func (ac *defaultApplicationContext) GetIDGenerator() id.IDGenerator {
	return ac.id
}

// GetConfig implements ApplicationContext.
func (ac *defaultApplicationContext) GetConfig() config.Config {
	return ac.config
}

// GetMessageSource implements ApplicationContext.
func (ac *defaultApplicationContext) GetMessageSource() message.MessageSource {
	return ac.messageSource
}

// GetLogger implements ApplicationContext.
func (ac *defaultApplicationContext) GetLogger() logging.Logger {
	return ac.logger
}

// GetDateManager implements ApplicationContext.
func (ac *defaultApplicationContext) GetDateManager() date.DateManager {
	return ac.dateManager
}

// GetDynamoDBAccessor implements ApplicationContext.
func (ac *defaultApplicationContext) GetDynamoDBAccessor() transaction.TransactionalDynamoDBAccessor {
	return ac.dynamoDBAccessor
}

// GetDynamoDBTransactionManager implements ApplicationContext.
func (ac *defaultApplicationContext) GetDynamoDBTransactionManager() transaction.TransactionManager {
	return ac.dynamoDBTransactionManager
}

// GetDynamoDBTransactionManagerForDBOnly implements ApplicationContext.
func (ac *defaultApplicationContext) GetDynamoDBTransactionManagerForDBOnly() transaction.TransactionManager {
	return ac.dynamoDBTransactionManagerForDBOnly
}

// GetDynamoDBTemplate implements ApplicationContext.
func (ac *defaultApplicationContext) GetDynamoDBTemplate() transaction.TransactionalDynamoDBTemplate {
	return ac.dynamodbTempalte
}

// GetRDBAccessor implements ApplicationContext.
func (ac *defaultApplicationContext) GetRDBAccessor() rdb.RDBAccessor {
	return ac.rdbAccessor
}

// GetRDBTransactionManager implements ApplicationContext.
func (ac *defaultApplicationContext) GetRDBTransactionManager() rdb.TransactionManager {
	return ac.rdbTransactionManager
}

// GetSQSAccessor implements ApplicationContext.
func (ac *defaultApplicationContext) GetSQSAccessor() transaction.TransactionalSQSAccessor {
	return ac.sqsAccessor
}

// GetSQSTemplate implements ApplicationContext.
func (ac *defaultApplicationContext) GetSQSTemplate() async.SQSTemplate {
	return ac.sqsTemplate
}

// GetObjectStorageAccessor implements ApplicationContext.
func (ac *defaultApplicationContext) GetObjectStorageAccessor() objectstorage.ObjectStorageAccessor {
	return ac.objectStorageAccessor
}

// GetHttpClient implements ApplicationContext.
func (ac *defaultApplicationContext) GetHttpClient() httpclient.HttpClient {
	return ac.httpClient
}

// GetInterceptor implements ApplicationContext.
func (ac *defaultApplicationContext) GetInterceptor() handler.HandlerInterceptor {
	return ac.interceptor
}

// GetAPILambdaHandler implements ApplicationContext.
func (ac *defaultApplicationContext) GetAPILambdaHandler() *handler.APILambdaHandler {
	return ac.apiLambdaHandler
}

// GetAsyncLambdaHandler implements ApplicationContext.
func (ac *defaultApplicationContext) GetAsyncLambdaHandler() *handler.AsyncLambdaHandler {
	return ac.asyncLambdaHandler
}

// GetSimpleLambdaHandler implements ApplicationContext.
func (ac *defaultApplicationContext) GetSimpleLambdaHandler() *handler.SimpleLambdaHandler {
	return ac.simpleLambdaHandler
}

// GetValidationManager implements ApplicationContext.
func (ac *defaultApplicationContext) GetValidationManager() validator.ValidationManager {
	return ac.validationManager
}

func createIDGenerator() id.IDGenerator {
	return id.NewIDGenerator()
}

func createMessageSource() message.MessageSource {
	messageSource, err := message.NewMessageSource()
	if err != nil {
		// 異常終了
		panic(err)
	}
	return messageSource
}

func createLogger(messageSource message.MessageSource) logging.Logger {
	logger, err := logging.NewLogger(messageSource)
	if err != nil {
		// 異常終了
		panic(err)
	}
	return logger
}

func createConfig(logger logging.Logger) config.Config {
	cfg, err := config.NewConfig(logger)
	if err != nil {
		// 異常終了
		panic(err)
	}
	return cfg
}

func createDateManager(config config.Config, logger logging.Logger) date.DateManager {
	return date.NewDateManager(config, logger)
}

func createTransactionalDynamoDBAccessor(logger logging.Logger, config config.Config) transaction.TransactionalDynamoDBAccessor {
	accessor, err := transaction.NewTransactionalDynamoDBAccessor(logger, config)
	if err != nil {
		// 異常終了
		panic(err)
	}
	return accessor
}

func createTransactionalSQSAccessor(logger logging.Logger, config config.Config, messageRegisterer transaction.MessageRegisterer) transaction.TransactionalSQSAccessor {
	accessor, err := transaction.NewTransactionalSQSAccessor(logger, config, messageRegisterer)
	if err != nil {
		// 異常終了
		panic(err)
	}
	return accessor
}

func createSQSTemplate(logger logging.Logger, config config.Config, id id.IDGenerator, sqsAccessor transaction.TransactionalSQSAccessor) async.SQSTemplate {
	return transaction.NewSQSTemplate(logger, config, id, sqsAccessor)
}

func createDynamoDBTransactionManager(logger logging.Logger,
	dynamodbAccessor transaction.TransactionalDynamoDBAccessor,
	sqsAccessor transaction.TransactionalSQSAccessor,
	messageRegigsterer transaction.MessageRegisterer) transaction.TransactionManager {
	return transaction.NewTransactionManager(logger, dynamodbAccessor, sqsAccessor, messageRegigsterer)
}

func createDynamoDBTransactionManagerForDBOnly(logger logging.Logger,
	dynamodbAccessor transaction.TransactionalDynamoDBAccessor,
	messageRegigsterer transaction.MessageRegisterer) transaction.TransactionManager {
	return transaction.NewTransactionManagerForDBOnly(logger, dynamodbAccessor, messageRegigsterer)
}

func createDynamoDBTemplate(logger logging.Logger, dynamodbAccessor transaction.TransactionalDynamoDBAccessor) transaction.TransactionalDynamoDBTemplate {
	return transaction.NewTransactionalDynamoDBTemplate(logger, dynamodbAccessor)
}

func createObjectStorageAccessor(config config.Config, logger logging.Logger) objectstorage.ObjectStorageAccessor {
	accessor, err := objectstorage.NewObjectStorageAccessor(config, logger)
	if err != nil {
		// 異常終了
		panic(err)
	}
	return accessor
}

func createRDBAccessor() rdb.RDBAccessor {
	return rdb.NewRDBAccessor()
}

func createHttpClient(config config.Config, logger logging.Logger) httpclient.HttpClient {
	return httpclient.NewHttpClient(config, logger)
}

func createHanderInterceptor(config config.Config, logger logging.Logger) handler.HandlerInterceptor {
	return handler.NewHandlerInterceptor(config, logger)
}

func createApiResponseFormatter(logger logging.Logger, messageSource message.MessageSource) api.ApiResponseFormatter {
	return api.NewApiResponseFormatter(logger, messageSource)
}

func createAPILambdaHandler(config config.Config, logger logging.Logger, messageSource message.MessageSource, apiResponseFormatter api.ApiResponseFormatter) *handler.APILambdaHandler {
	return handler.NewAPILambdaHandler(config, logger, messageSource, apiResponseFormatter)
}

func createAsyncLambdaHandler(config config.Config, logger logging.Logger, queueMessageItemRepository transaction.QueueMessageItemRepository) *handler.AsyncLambdaHandler {
	return handler.NewAsyncLambdaHandler(config, logger, queueMessageItemRepository)
}

func createSimpleLambdaHandler(config config.Config, logger logging.Logger) *handler.SimpleLambdaHandler {
	return handler.NewSimpleLambdaHandler(config, logger)
}

func createQueueMessageItemRepository(config config.Config, logger logging.Logger, dynamodbTemplate transaction.TransactionalDynamoDBTemplate) transaction.QueueMessageItemRepository {
	return transaction.NewQueueMessageItemRepository(config, logger, dynamodbTemplate)
}

func createMessageRegisterer(queueMessageItemRepository transaction.QueueMessageItemRepository) transaction.MessageRegisterer {
	return transaction.NewMessageRegisterer(queueMessageItemRepository)
}

func createValidationManager(logger logging.Logger) validator.ValidationManager {
	return validator.NewValidationManager(logger.Debug, logger.Warn)
}
