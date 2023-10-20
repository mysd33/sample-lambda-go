/*
logging パッケージは、ログ出力に関する機能を提供するパッケージです。
*/
package logging

import (
	"example.com/appbase/pkg/message"
	"go.uber.org/zap"
)

//TODO: ログレベルの設定

type any interface{}

// Loggerは、ログ出力のインタフェースです
type Logger interface {
	// Debugは、デバッグレベルのログを出力します。
	Debug(template string, args ...any)
	// Infoは、情報レベルのログを出力します。
	Info(code string, args ...any)
	// Warnは、警告レベルのログを出力します。
	Warn(code string, args ...any)
	// Errorは、エラーレベルのログを出力します。
	Error(code string, args ...any)
	// Fatailは、致命的なエラーレベルのログを出力します。
	Fatal(code string, args ...any)
}

// NewLogger は、Loggerを作成します。
func NewLogger() Logger {
	z, _ := zap.NewProduction()
	return ZapLogger{log: z.Sugar(), messageSource: message.NewMessageSource()}
}

// ZapLoggerは、Zapを使ったLogger実装です。
type ZapLogger struct {
	log           *zap.SugaredLogger
	messageSource message.MessageSource
}

func (z ZapLogger) Debug(template string, args ...any) {
	z.log.Debugf(template, args)
}

func (z ZapLogger) Info(code string, args ...any) {
	message := z.messageSource.GetMessage(code, args)
	if message != "" {
		z.log.Infof(message)
	}
	z.log.Infof(code, args)
}

func (z ZapLogger) Warn(code string, args ...any) {
	message := z.messageSource.GetMessage(code, args)
	if message != "" {
		z.log.Warnf(message)
	}
	z.log.Warnf(code, args)
}

func (z ZapLogger) Error(code string, args ...any) {
	message := z.messageSource.GetMessage(code, args)
	if message != "" {
		z.log.Errorf(message)
	}
	z.log.Errorf(code, args)
}

func (z ZapLogger) Fatal(code string, args ...any) {
	message := z.messageSource.GetMessage(code, args)
	if message != "" {
		z.log.Fatalf(message)
	}
	z.log.Fatalf(code, args)
}
