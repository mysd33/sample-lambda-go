/*
component パッケージはフレームワークのコンポーネントのインスタンスを管理するパッケージです。
*/
package component

import (
	"log"

	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/dynamodb"
	"example.com/appbase/pkg/handler"
	"example.com/appbase/pkg/httpclient"
	"example.com/appbase/pkg/logging"
	"example.com/appbase/pkg/message"
)

// ApplicationContext は、フレームワークのコンポーネントを管理するインタフェースです。
type ApplicationContext interface {
	GetMessageSource() message.MessageSource
	GetLogger() logging.Logger
	GetConfig() *config.Config
	GetDynamoDBAccessor() dynamodb.DynamoDBAccessor
	GetHttpClient() httpclient.HttpClient
	GetInterceptor() handler.HandlerInterceptor
}

// NewApplicationContext は、デフォルトのApplicationContextを作成します。
func NewApplicationContext() ApplicationContext {
	messageSource := createMessageSource()
	logger := createLogger(messageSource)
	config := createConfig()
	dynamodbAccessor := createDynamoDBAccessor(logger)
	httpclient := createHttpClient(logger)
	interceptor := createHanderInterceptor(logger)
	return &defaultApplicationContext{
		config:           config,
		messageSource:    nil,
		logger:           logger,
		dynamoDBAccessor: dynamodbAccessor,
		httpClient:       httpclient,
		interceptor:      interceptor,
	}
}

type defaultApplicationContext struct {
	config           *config.Config
	messageSource    message.MessageSource
	logger           logging.Logger
	dynamoDBAccessor dynamodb.DynamoDBAccessor
	httpClient       httpclient.HttpClient
	interceptor      handler.HandlerInterceptor
}

// GetConfig implements ApplicationContext.
func (ac *defaultApplicationContext) GetConfig() *config.Config {
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
func (*defaultApplicationContext) GetMessageSource() message.MessageSource {
	panic("unimplemented")
}

func createMessageSource() message.MessageSource {
	return message.NewMessageSource()
}

func createLogger(messageSource message.MessageSource) logging.Logger {
	logger, err := logging.NewLogger(messageSource)
	if err != nil {
		// 異常終了
		log.Fatalf("初期化処理エラー:%s", err.Error())
	}
	return logger
}

func createConfig() *config.Config {
	cfg, err := config.LoadConfig()
	if err != nil {
		// 異常終了
		log.Fatalf("初期化処理エラー:%s", err.Error())
	}
	return cfg
}

func createDynamoDBAccessor(logger logging.Logger) dynamodb.DynamoDBAccessor {
	accessor, err := dynamodb.NewDynamoDBAccessor(logger)
	if err != nil {
		// 異常終了
		log.Fatalf("初期化処理エラー:%s", err.Error())
	}
	return accessor
}

func createHttpClient(logger logging.Logger) httpclient.HttpClient {
	return httpclient.NewHttpClient(logger)
}

func createHanderInterceptor(logger logging.Logger) handler.HandlerInterceptor {
	return handler.NewHandlerInterceptor(logger)
}
