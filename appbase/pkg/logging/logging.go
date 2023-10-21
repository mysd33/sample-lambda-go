/*
logging パッケージは、ログ出力に関する機能を提供するパッケージです。
*/
package logging

import (
	"example.com/appbase/pkg/message"
	"go.uber.org/zap"
)

// Loggerは、ログ出力のインタフェースです
type Logger interface {
	// Debugは、メッセージのテンプレートtemplate, 置き換え文字列argsに対してfmt.Sprintfしたメッセージでデバッグレベルのログを出力します。
	Debug(template string, args ...interface{})
	// Infoは、メッセージID（code）、置き換え文字列argsに対応するメッセージで、情報レベルのログを出力します。codeに対応するメッセージがない場合はそのまま出力します。
	Info(code string, args ...interface{})
	// Warnは、メッセージID（エラーコードcode）、置き換え文字列argsに対応するメッセージで警告レベルのログを出力します。codeに対応するメッセージがない場合はそのまま出力します。
	Warn(code string, args ...interface{})
	// Errorは、メッセージID（エラーコードcode）、置き換え文字列argsに対応するメッセージでエラーレベルのログを出力します。codeに対応するメッセージがない場合はそのまま出力します。
	Error(code string, args ...interface{})
	// Fatailは、メッセージID（エラーコードcode）、置き換え文字列argsに対応するメッセージで致命的なエラーレベルのログを出力します。codeに対応するメッセージがない場合はそのまま出力します。
	Fatal(code string, args ...interface{})
}

// NewLogger は、Loggerを作成します。
func NewLogger() Logger {
	// TODO: ログレベルの設定
	//z, _ := zap.NewProduction()
	z, _ := zap.NewDevelopment()
	return &zapLogger{log: z.Sugar(), messageSource: message.NewMessageSource()}
}

// zapLoggerは、Zapを使ったLogger実装です。
type zapLogger struct {
	log           *zap.SugaredLogger
	messageSource message.MessageSource
}

func (z *zapLogger) Debug(template string, args ...interface{}) {
	z.log.Debugf(template, args...)
}

func (z *zapLogger) Info(code string, args ...interface{}) {
	message := z.messageSource.GetMessage(code, args)
	if message != "" {
		z.log.Infof(message)
	}
	z.log.Info(code, args)
}

func (z *zapLogger) Warn(code string, args ...interface{}) {
	message := z.messageSource.GetMessage(code, args)
	if message != "" {
		z.log.Warnf(message)
	}
	z.log.Warn(code, args)
}

func (z *zapLogger) Error(code string, args ...interface{}) {
	message := z.messageSource.GetMessage(code, args)
	if message != "" {
		z.log.Errorf(message)
	}
	z.log.Error(code, args)
}

func (z *zapLogger) Fatal(code string, args ...interface{}) {
	message := z.messageSource.GetMessage(code, args)
	if message != "" {
		z.log.Fatalf(message)
	}
	z.log.Fatal(code, args)
}
