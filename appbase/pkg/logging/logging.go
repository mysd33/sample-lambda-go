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
	// Debugは、メッセージのテンプレートtemplate, 置き換え文字列argsに対してfmt.Sprintfしたメッセージでデバッグレベルのログを出力します。
	Debug(template string, args ...any)
	// Infoは、メッセージID（code）、置き換え文字列argsに対応するメッセージで、情報レベルのログを出力します。codeに対応するメッセージがない場合はそのまま出力します。
	Info(code string, args ...any)
	// Warnは、メッセージID（エラーコードcode）、置き換え文字列argsに対応するメッセージで警告レベルのログを出力します。codeに対応するメッセージがない場合はそのまま出力します。
	Warn(code string, args ...any)
	// Errorは、メッセージID（エラーコードcode）、置き換え文字列argsに対応するメッセージでエラーレベルのログを出力します。codeに対応するメッセージがない場合はそのまま出力します。
	Error(code string, args ...any)
	// Fatailは、メッセージID（エラーコードcode）、置き換え文字列argsに対応するメッセージで致命的なエラーレベルのログを出力します。codeに対応するメッセージがない場合はそのまま出力します。
	Fatal(code string, args ...any)
}

// NewLogger は、Loggerを作成します。
func NewLogger() Logger {
	z, _ := zap.NewProduction()
	return &zapLogger{log: z.Sugar(), messageSource: message.NewMessageSource()}
}

// zapLoggerは、Zapを使ったLogger実装です。
type zapLogger struct {
	log           *zap.SugaredLogger
	messageSource message.MessageSource
}

func (z *zapLogger) Debug(template string, args ...any) {
	z.log.Debugf(template, args)
}

func (z *zapLogger) Info(code string, args ...any) {
	message := z.messageSource.GetMessage(code, args)
	if message != "" {
		z.log.Infof(message)
	}
	z.log.Info(code, args)
}

func (z *zapLogger) Warn(code string, args ...any) {
	message := z.messageSource.GetMessage(code, args)
	if message != "" {
		z.log.Warnf(message)
	}
	z.log.Warn(code, args)
}

func (z *zapLogger) Error(code string, args ...any) {
	message := z.messageSource.GetMessage(code, args)
	if message != "" {
		z.log.Errorf(message)
	}
	z.log.Error(code, args)
}

func (z *zapLogger) Fatal(code string, args ...any) {
	message := z.messageSource.GetMessage(code, args)
	if message != "" {
		z.log.Fatalf(message)
	}
	z.log.Fatal(code, args)
}
