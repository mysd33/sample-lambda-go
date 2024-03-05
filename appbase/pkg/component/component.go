/*
component パッケージはフレームワークのコンポーネントのインスタンスを管理するパッケージです。
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
	GetIDGenerator() id.IDGenerator
	GetMessageSource() message.MessageSource
	GetLogger() logging.Logger
	GetConfig() config.Config
	GetDynamoDBAccessor() transaction.TransactionalDynamoDBAccessor
	GetDynamoDBTransactionManager() transaction.TransactionManager
	GetDynamoDBTransactionManagerForDBOnly() transaction.TransactionManager
	GetDynamoDBTemplate() transaction.TransactionalDynamoDBTemplate
	GetSQSAccessor() transaction.TransactionalSQSAccessor
	GetSQSTemplate() async.SQSTemplate
	GetObjectStorageAccessor() objectstorage.ObjectStorageAccessor
	GetRDBAccessor() rdb.RDBAccessor
	GetRDBTransactionManager() rdb.TransactionManager
	GetHttpClient() httpclient.HttpClient
	GetInterceptor() handler.HandlerInterceptor
	GetAPILambdaHandler() *handler.APILambdaHandler
	GetAsyncLambdaHandler() *handler.AsyncLambdaHandler
	GetSimpleLambdaHandler() *handler.SimpleLambdaHandler
	GetValidationManager() validator.ValidationManager
	GetDateManager() date.DateManager
}

// NewApplicationContext は、デフォルトのApplicationContextを作成します。
func NewApplicationContext() ApplicationContext {
	// 各種AP基盤の構造体を作成
	id := createIDGenerator()
	config := createConfig()
	messageSource := createMessageSource()
	logger := createLogger(messageSource, config)
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

func createLogger(messageSource message.MessageSource, config config.Config) logging.Logger {
	logger, err := logging.NewLogger(messageSource, config)
	if err != nil {
		// 異常終了
		panic(err)
	}
	return logger
}

func createConfig() config.Config {
	cfg, err := config.NewConfig()
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
	return validator.NewValidationManager(logger.Debug)
}
