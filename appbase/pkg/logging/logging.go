/*
logging パッケージは、ログ出力に関する機能を提供するパッケージです。
*/
package logging

import (
	"os"

	"example.com/appbase/pkg/errors"
	"example.com/appbase/pkg/message"
	"go.uber.org/zap"
)

const (
	PROFILE_PRODUCTION = "production"
)

// Loggerは、ログ出力のインタフェースです
type Logger interface {
	// Debug は、メッセージのテンプレートtemplate, 置き換え文字列argsに対してfmt.Sprintfしたメッセージでデバッグレベルのログを出力します。
	Debug(template string, args ...interface{})
	// Info は、メッセージID（messages）、置き換え文字列argsに対応するメッセージで、情報レベルのログを出力します。codeに対応するメッセージがない場合はそのまま出力します。
	Info(code string, args ...interface{})
	// Warn は、メッセージID（エラーコードcode）、置き換え文字列argsに対応するメッセージで警告レベルのログを出力します。codeに対応するメッセージがない場合はそのまま出力します。
	Warn(code string, args ...interface{})
	// WarnWithCodableError は、エラーが持つメッセージID（エラーコード）、置き換え文字列に対応するメッセージで警告レベルのログを出力します。codeに対応するメッセージがない場合はそのまま出力します。
	WarnWithCodableError(err errors.CodableError)
	// Error は、メッセージID（エラーコードcode）、置き換え文字列argsに対応するメッセージでエラーレベルのログを出力します。codeに対応するメッセージがない場合はそのまま出力します。
	Error(code string, args ...interface{})
	// ErrorWithCodableError は、エラーが持つメッセージID（エラーコード）、置き換え文字列に対応するメッセージでエラーレベルのログを出力します。codeに対応するメッセージがない場合はそのまま出力します。
	ErrorWithCodableError(err errors.CodableError)
	// 	ErrorWithUnexpectedError は、予期せぬエラーをログに出力します。
	ErrorWithUnexpectedError(err error)
}

// NewLogger は、Loggerを作成します。
func NewLogger(messageSource message.MessageSource) (Logger, error) {
	profile := os.Getenv("LOG_PROFILE")
	var config zap.Config
	// プロファイルの切り替え
	if profile == PROFILE_PRODUCTION {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
	}
	z, err := config.Build(zap.AddCallerSkip(1))
	if err != nil {
		return nil, err
	}
	return &zapLogger{log: z.Sugar(), messageSource: messageSource}, nil
}

// zapLoggerは、Zapを使ったLogger実装です。
type zapLogger struct {
	log           *zap.SugaredLogger
	messageSource message.MessageSource
}

// Debug implements Logger.
func (z *zapLogger) Debug(template string, args ...interface{}) {
	z.log.Debugf(template, args...)
}

// Info implements Logger.
func (z *zapLogger) Info(code string, args ...interface{}) {
	message := z.messageSource.GetMessage(code, args...)
	if message != "" {
		z.log.Infof(message)
		return
	}
	z.log.Info(code, args)
}

// Warn implements Logger.
func (z *zapLogger) Warn(code string, args ...interface{}) {
	message := z.messageSource.GetMessage(code, args...)
	if message != "" {
		z.log.Warnf(message)
		return
	}
	z.log.Warn(code, args)
}

// WarnWithCodableError implements Logger.
func (z *zapLogger) WarnWithCodableError(err errors.CodableError) {
	code := err.ErrorCode()
	args := err.Args()
	message := z.messageSource.GetMessage(code, args...)
	// エラーのスタックトレース付きのWarnログ出力
	if message != "" {
		z.log.Warnf("%s, %+v", message, err)
		return
	}
	z.log.Warnf("%s[%v], %+v", code, args, err)
}

// Error implements Logger.
func (z *zapLogger) Error(code string, args ...interface{}) {
	message := z.messageSource.GetMessage(code, args...)
	if message != "" {
		z.log.Errorf(message)
		return
	}
	z.log.Error(code, args)
}

// ErrorWithCodableError implements Logger.
func (z *zapLogger) ErrorWithCodableError(err errors.CodableError) {
	code := err.ErrorCode()
	args := err.Args()
	message := z.messageSource.GetMessage(code, args...)
	// エラーのスタックトレース付きのErrorログ出力
	if message != "" {
		z.log.Errorf("%s, %+v", message, err)
		return
	}
	z.log.Errorf("%s[%v], %+v", code, args, err)
}

// Error implements Logger.
func (z *zapLogger) ErrorWithUnexpectedError(err error) {
	message := z.messageSource.GetMessage(message.E_FW_9999)
	z.log.Errorf("%s, %+v", message, err)
}
