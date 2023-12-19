/*
component パッケージはフレームワークのコンポーネントのインスタンスを管理するパッケージです。
*/
package component

import (
	"log"

	"example.com/appbase/pkg/api"
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
	GetSQSAccessor() transaction.TransactionalSQSAccessor
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
	sqsAccessor := createTransactionalSQSAccessor(logger, config)
	dynamodbAccessor := createTransactionalDynamoDBAccessor(logger, config)
	dynamoDBTransactionManager := createDynamoDBTransactionManager(logger, dynamodbAccessor, sqsAccessor)
	rdbAccessor := createRDBAccessor()
	rdbTransactionManager := rdb.NewTransactionManager(logger, config, rdbAccessor)
	httpclient := createHttpClient(logger)
	interceptor := createHanderInterceptor(config, logger)
	apiLambdaHandler := createAPILambdaHandler(config, logger, messageSource, apiResponseFormatter)
	asyncLambdaHandler := createAsyncLambdaHandler(config, logger)

	// Validatorの日本語化
	validator.Setup()

	return &defaultApplicationContext{
		config:                     config,
		messageSource:              messageSource,
		logger:                     logger,
		dynamoDBAccessor:           dynamodbAccessor,
		dynamoDBTransactionManager: dynamoDBTransactionManager,
		sqsAccessor:                sqsAccessor,
		rdbAccessor:                rdbAccessor,
		rdbTransactionManager:      rdbTransactionManager,
		httpClient:                 httpclient,
		interceptor:                interceptor,
		apiLambdaHandler:           apiLambdaHandler,
		asyncLambdaHandler:         asyncLambdaHandler,
	}
}

type defaultApplicationContext struct {
	config                     config.Config
	messageSource              message.MessageSource
	logger                     logging.Logger
	dynamoDBAccessor           transaction.TransactionalDynamoDBAccessor
	dynamoDBTransactionManager transaction.TransactionManager
	sqsAccessor                transaction.TransactionalSQSAccessor
	rdbAccessor                rdb.RDBAccessor
	rdbTransactionManager      rdb.TransactionManager
	httpClient                 httpclient.HttpClient
	interceptor                handler.HandlerInterceptor
	apiLambdaHandler           *handler.APILambdaHandler
	asyncLambdaHandler         *handler.AsyncLambdaHandler
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
		log.Fatalf("初期化処理エラー:%+v", errors.WithStack(err))
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
		log.Fatalf("初期化処理エラー:%+v", errors.WithStack(err))
	}
	return logger
}

func createConfig() config.Config {
	cfg, err := config.NewConfig()
	if err != nil {
		// 異常終了
		log.Fatalf("初期化処理エラー:%+v", errors.WithStack(err))
	}
	return cfg
}

func createTransactionalDynamoDBAccessor(logger logging.Logger, config config.Config) transaction.TransactionalDynamoDBAccessor {
	accessor, err := transaction.NewTransactionalDynamoDBAccessor(logger, config)
	if err != nil {
		// 異常終了
		log.Fatalf("初期化処理エラー:%+v", errors.WithStack(err))
	}
	return accessor
}

func createTransactionalSQSAccessor(logger logging.Logger, config config.Config) transaction.TransactionalSQSAccessor {
	accessor, err := transaction.NewTransactionalSQSAccessor(logger, config)
	if err != nil {
		// 異常終了
		log.Fatalf("初期化処理エラー:%+v", errors.WithStack(err))
	}
	return accessor
}

func createDynamoDBTransactionManager(logger logging.Logger,
	dynamodbAccessor transaction.TransactionalDynamoDBAccessor,
	sqsAccessor transaction.TransactionalSQSAccessor) transaction.TransactionManager {
	return transaction.NewTransactionManager(logger, dynamodbAccessor, sqsAccessor)
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

func createAsyncLambdaHandler(config config.Config, logger logging.Logger) *handler.AsyncLambdaHandler {
	return handler.NewAsyncLambdaHandler(config, logger)
}
