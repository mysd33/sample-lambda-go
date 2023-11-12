/*
component パッケージはフレームワークのコンポーネントのインスタンスを管理するパッケージです。
*/
package component

import (
	"log"

	"example.com/appbase/pkg/api"
	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/dynamodb"
	"example.com/appbase/pkg/handler"
	"example.com/appbase/pkg/httpclient"
	"example.com/appbase/pkg/logging"
	"example.com/appbase/pkg/message"
	"example.com/appbase/pkg/validator"
	"github.com/cockroachdb/errors"
)

// ApplicationContext は、フレームワークのコンポーネントを管理するインタフェースです。
type ApplicationContext interface {
	GetMessageSource() message.MessageSource
	GetLogger() logging.Logger
	GetConfig() config.Config
	GetDynamoDBAccessor() dynamodb.DynamoDBAccessor
	GetHttpClient() httpclient.HttpClient
	GetInterceptor() handler.HandlerInterceptor
}

// NewApplicationContext は、デフォルトのApplicationContextを作成します。
func NewApplicationContext() ApplicationContext {
	// 各種AP基盤の構造体を作成
	messageSource := createMessageSource()
	apiResponseFormatter := createApiResponseFormatter(messageSource)
	logger := createLogger(messageSource)
	config := createConfig()
	dynamodbAccessor := createDynamoDBAccessor(logger)
	httpclient := createHttpClient(logger)
	interceptor := createHanderInterceptor(config, logger, apiResponseFormatter)

	// Validatorの日本語化
	validator.Setup()

	return &defaultApplicationContext{
		config:           config,
		messageSource:    messageSource,
		logger:           logger,
		dynamoDBAccessor: dynamodbAccessor,
		httpClient:       httpclient,
		interceptor:      interceptor,
	}
}

type defaultApplicationContext struct {
	config           config.Config
	messageSource    message.MessageSource
	logger           logging.Logger
	dynamoDBAccessor dynamodb.DynamoDBAccessor
	httpClient       httpclient.HttpClient
	interceptor      handler.HandlerInterceptor
}

// GetConfig implements ApplicationContext.
func (ac *defaultApplicationContext) GetConfig() config.Config {
	return ac.config
}

// GetDynamoDBAccessor implements ApplicationContext.
func (ac *defaultApplicationContext) GetDynamoDBAccessor() dynamodb.DynamoDBAccessor {
	return ac.dynamoDBAccessor
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

func createLogger(messageSource message.MessageSource) logging.Logger {
	logger, err := logging.NewLogger(messageSource)
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

func createDynamoDBAccessor(logger logging.Logger) dynamodb.DynamoDBAccessor {
	accessor, err := dynamodb.NewDynamoDBAccessor(logger)
	if err != nil {
		// 異常終了
		log.Fatalf("初期化処理エラー:%+v", errors.WithStack(err))
	}
	return accessor
}

func createHttpClient(logger logging.Logger) httpclient.HttpClient {
	return httpclient.NewHttpClient(logger)
}

func createHanderInterceptor(config config.Config, logger logging.Logger, apiResponseFormatter api.ApiResponseFormatter) handler.HandlerInterceptor {
	return handler.NewHandlerInterceptor(config, logger, apiResponseFormatter)
}
