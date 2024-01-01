/*
component パッケージはフレームワークのコンポーネントのインスタンスを管理するパッケージです。
*/
package component

import (
	"example.com/appbase/pkg/api"
	"example.com/appbase/pkg/async"
	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/handler"
	"example.com/appbase/pkg/httpclient"
	"example.com/appbase/pkg/logging"
	"example.com/appbase/pkg/message"
	"example.com/appbase/pkg/rdb"
	"example.com/appbase/pkg/transaction"
	"example.com/appbase/pkg/validator"
	"github.com/cockroachdb/errors"
)

// ApplicationContext は、フレームワークのコンポーネントを管理するインタフェースです。
type ApplicationContext interface {
	GetMessageSource() message.MessageSource
	GetLogger() logging.Logger
	GetConfig() config.Config
	GetDynamoDBAccessor() transaction.TransactionalDynamoDBAccessor
	GetDynamoDBTransactionManager() transaction.TransactionManager
	GetDynamoDBTransactionManagerForDBOnly() transaction.TransactionManager
	GetDynamoDBTemplate() transaction.TransactionalDynamoDBTemplate
	GetSQSAccessor() transaction.TransactionalSQSAccessor
	GetSQSTemplate() async.SQSTemplate
	GetRDBAccessor() rdb.RDBAccessor
	GetRDBTransactionManager() rdb.TransactionManager
	GetHttpClient() httpclient.HttpClient
	GetInterceptor() handler.HandlerInterceptor
	GetAPILambdaHandler() *handler.APILambdaHandler
	GetAsyncLambdaHandler() *handler.AsyncLambdaHandler
}

// NewApplicationContext は、デフォルトのApplicationContextを作成します。
func NewApplicationContext() ApplicationContext {
	// 各種AP基盤の構造体を作成
	config := createConfig()
	messageSource := createMessageSource()
	apiResponseFormatter := createApiResponseFormatter(messageSource)
	logger := createLogger(messageSource, config)
	dynamodbAccessor := createTransactionalDynamoDBAccessor(logger, config)
	dynamoDBTempalte := createDynamoDBTemplate(logger, dynamodbAccessor)
	queueMessageItemRepository := createQueueMessageItemRepository(config, logger, dynamoDBTempalte)
	messageRegisterer := createMessageRegisterer(queueMessageItemRepository)
	sqsAccessor := createTransactionalSQSAccessor(logger, config, messageRegisterer)
	sqsTemplate := createSQSTemplate(logger, sqsAccessor)
	dynamoDBTransactionManager := createDynamoDBTransactionManager(logger, dynamodbAccessor, sqsAccessor, messageRegisterer)
	dynamoDBTransactionManagerForDBOnly := createDynamoDBTransactionManagerForDBOnly(logger, dynamodbAccessor, messageRegisterer)
	rdbAccessor := createRDBAccessor()
	rdbTransactionManager := rdb.NewTransactionManager(logger, config, rdbAccessor)
	httpclient := createHttpClient(logger)
	interceptor := createHanderInterceptor(config, logger)
	apiLambdaHandler := createAPILambdaHandler(config, logger, messageSource, apiResponseFormatter)
	asyncLambdaHandler := createAsyncLambdaHandler(config, logger, queueMessageItemRepository)

	// Validatorの日本語化
	validator.Setup()

	return &defaultApplicationContext{
		config:                              config,
		messageSource:                       messageSource,
		logger:                              logger,
		dynamoDBAccessor:                    dynamodbAccessor,
		dynamoDBTransactionManager:          dynamoDBTransactionManager,
		dynamoDBTransactionManagerForDBOnly: dynamoDBTransactionManagerForDBOnly,
		dynamodbTempalte:                    dynamoDBTempalte,
		sqsAccessor:                         sqsAccessor,
		sqsTemplate:                         sqsTemplate,
		rdbAccessor:                         rdbAccessor,
		rdbTransactionManager:               rdbTransactionManager,
		httpClient:                          httpclient,
		interceptor:                         interceptor,
		apiLambdaHandler:                    apiLambdaHandler,
		asyncLambdaHandler:                  asyncLambdaHandler,
	}
}

type defaultApplicationContext struct {
	config                              config.Config
	messageSource                       message.MessageSource
	logger                              logging.Logger
	dynamoDBAccessor                    transaction.TransactionalDynamoDBAccessor
	dynamoDBTransactionManager          transaction.TransactionManager
	dynamoDBTransactionManagerForDBOnly transaction.TransactionManager
	dynamodbTempalte                    transaction.TransactionalDynamoDBTemplate
	sqsAccessor                         transaction.TransactionalSQSAccessor
	sqsTemplate                         async.SQSTemplate
	rdbAccessor                         rdb.RDBAccessor
	rdbTransactionManager               rdb.TransactionManager
	httpClient                          httpclient.HttpClient
	interceptor                         handler.HandlerInterceptor
	apiLambdaHandler                    *handler.APILambdaHandler
	asyncLambdaHandler                  *handler.AsyncLambdaHandler
}

// GetConfig implements ApplicationContext.
func (ac *defaultApplicationContext) GetConfig() config.Config {
	return ac.config
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

// GetHttpClient implements ApplicationContext.
func (ac *defaultApplicationContext) GetHttpClient() httpclient.HttpClient {
	return ac.httpClient
}

// GetInterceptor implements ApplicationContext.
func (ac *defaultApplicationContext) GetInterceptor() handler.HandlerInterceptor {
	return ac.interceptor
}

// GetLogger implements ApplicationContext.
func (ac *defaultApplicationContext) GetLogger() logging.Logger {
	return ac.logger
}

// GetMessageSource implements ApplicationContext.
func (ac *defaultApplicationContext) GetMessageSource() message.MessageSource {
	return ac.messageSource
}

// GetAPILambdaHandler implements ApplicationContext.
func (ac *defaultApplicationContext) GetAPILambdaHandler() *handler.APILambdaHandler {
	return ac.apiLambdaHandler
}

// GetAsyncLambdaHandler implements ApplicationContext.
func (ac *defaultApplicationContext) GetAsyncLambdaHandler() *handler.AsyncLambdaHandler {
	return ac.asyncLambdaHandler
}

func createMessageSource() message.MessageSource {
	messageSource, err := message.NewMessageSource()
	if err != nil {
		// 異常終了
		panic(errors.Wrap(err, "初期化処理エラー"))
	}
	return messageSource
}

func createApiResponseFormatter(messageSource message.MessageSource) api.ApiResponseFormatter {
	return api.NewApiResponseFormatter(messageSource)
}

func createLogger(messageSource message.MessageSource, config config.Config) logging.Logger {
	logger, err := logging.NewLogger(messageSource, config)
	if err != nil {
		// 異常終了
		panic(errors.Wrap(err, "初期化処理エラー"))
	}
	return logger
}

func createConfig() config.Config {
	cfg, err := config.NewConfig()
	if err != nil {
		// 異常終了
		panic(errors.Wrap(err, "初期化処理エラー"))
	}
	return cfg
}

func createTransactionalDynamoDBAccessor(logger logging.Logger, config config.Config) transaction.TransactionalDynamoDBAccessor {
	accessor, err := transaction.NewTransactionalDynamoDBAccessor(logger, config)
	if err != nil {
		// 異常終了
		panic(errors.Wrap(err, "初期化処理エラー"))
	}
	return accessor
}

func createTransactionalSQSAccessor(logger logging.Logger, config config.Config, messageRegisterer transaction.MessageRegisterer) transaction.TransactionalSQSAccessor {
	accessor, err := transaction.NewTransactionalSQSAccessor(logger, config, messageRegisterer)
	if err != nil {
		// 異常終了
		panic(errors.Wrap(err, "初期化処理エラー"))
	}
	return accessor
}

func createSQSTemplate(logger logging.Logger, sqsAccessor transaction.TransactionalSQSAccessor) async.SQSTemplate {
	return transaction.NewSQSTemplate(logger, sqsAccessor)
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

func createRDBAccessor() rdb.RDBAccessor {
	return rdb.NewRDBAccessor()
}

func createHttpClient(logger logging.Logger) httpclient.HttpClient {
	return httpclient.NewHttpClient(logger)
}

func createHanderInterceptor(config config.Config, logger logging.Logger) handler.HandlerInterceptor {
	return handler.NewHandlerInterceptor(config, logger)
}

func createAPILambdaHandler(config config.Config, logger logging.Logger, messageSource message.MessageSource, apiResponseFormatter api.ApiResponseFormatter) *handler.APILambdaHandler {
	return handler.NewAPILambdaHandler(config, logger, messageSource, apiResponseFormatter)
}

func createAsyncLambdaHandler(config config.Config, logger logging.Logger, queueMessageItemRepository transaction.QueueMessageItemRepository) *handler.AsyncLambdaHandler {
	return handler.NewAsyncLambdaHandler(config, logger, queueMessageItemRepository)
}

func createQueueMessageItemRepository(config config.Config, logger logging.Logger, dynamodbTemplate transaction.TransactionalDynamoDBTemplate) transaction.QueueMessageItemRepository {
	return transaction.NewQueueMessageItemRepository(config, logger, dynamodbTemplate)
}

func createMessageRegisterer(queueMessageItemRepository transaction.QueueMessageItemRepository) transaction.MessageRegisterer {
	return transaction.NewMessageRegisterer(queueMessageItemRepository)
}
