/*
logging パッケージは、ログ出力に関する機能を提供するパッケージです。
*/
package logging

import (
	"fmt"

	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/env"
	"example.com/appbase/pkg/errors"
	"example.com/appbase/pkg/message"
	"go.uber.org/zap"
)

const (
	LOG_LEVEL_NAME = "LOG_LEVEL"
)

// Loggerは、ログ出力のインタフェースです
type Logger interface {
	// AddInfo は、ログに付加情報を追加します。
	AddInfo(key string, value string)
	// ClearInfo は、ログに付加情報をクリアします。
	ClearInfo()
	// Debug は、メッセージのテンプレートtemplate, 置き換え文字列argsに対してfmt.Sprintfしたメッセージでデバッグレベルのログを出力します。
	Debug(template string, args ...any)
	// Info は、メッセージID（messages）、置き換え文字列argsに対応するメッセージで、情報レベルのログを出力します。codeに対応するメッセージがない場合はそのまま出力します。
	Info(code string, args ...any)
	// Warn は、メッセージID（エラーコードcode）、置き換え文字列argsに対応するメッセージで警告レベルのログを出力します。codeに対応するメッセージがない場合はそのまま出力します。
	Warn(code string, args ...any)
	// WarnWithCodableError は、エラーが持つメッセージID（エラーコード）、置き換え文字列に対応するメッセージで警告レベルのログを出力します。codeに対応するメッセージがない場合はそのまま出力します。
	WarnWithCodableError(err errors.CodableError)
	// WarnWithMultiError は、複数のエラーを持つエラーを警告レベルのログに出力します。
	WarnWithMultiCodableError(err errors.MultiCodableError)
	// Error は、メッセージID（エラーコードcode）、置き換え文字列argsに対応するメッセージでエラーレベルのログを出力します。codeに対応するメッセージがない場合はそのまま出力します。
	Error(code string, args ...any)
	// ErrorWithCodableError は、エラーが持つメッセージID（エラーコード）、置き換え文字列に対応するメッセージでエラーレベルのログを出力します。codeに対応するメッセージがない場合はそのまま出力します。
	ErrorWithCodableError(err errors.CodableError)
	// 	ErrorWithUnexpectedError は、予期せぬエラーをログに出力します。
	ErrorWithUnexpectedError(err error)
}

// NewLogger は、Loggerを作成します。
func NewLogger(messageSource message.MessageSource, mycfg config.Config) (Logger, error) {
	var config zap.Config
	if env.IsStragingOrProd() {
		// 本番相当の場合
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
	}
	// 個別にログレベルが設定されている場合はログレベル上書き
	level := mycfg.Get(LOG_LEVEL_NAME, "")
	if level != "" {
		al, err := zap.ParseAtomicLevel(level)
		if err == nil {
			config.Level = al
		}
	}
	z, err := config.Build(zap.AddCallerSkip(1))
	if err != nil {
		return nil, err
	}
	sugerredLogger := z.Sugar()
	return &zapLogger{originalLog: sugerredLogger,
		log:           sugerredLogger,
		messageSource: messageSource,
	}, nil
}

// zapLoggerは、Zapを使ったLogger実装です。
type zapLogger struct {
	originalLog   *zap.SugaredLogger
	log           *zap.SugaredLogger
	messageSource message.MessageSource
}

// AddInfo implements Logger.
func (z *zapLogger) AddInfo(key string, value string) {
	z.log = z.log.With(key, value)
}

// ClearInfo implements Logger.
func (z *zapLogger) ClearInfo() {
	z.log = z.originalLog
}

// Debug implements Logger.
func (z *zapLogger) Debug(template string, args ...any) {
	z.log.Debugf(template, args...)
}

// Info implements Logger.
func (z *zapLogger) Info(code string, args ...any) {
	message := z.messageSource.GetMessage(code, args...)
	if message != "" {
		z.log.Infof("[%s]%s", code, message)
		return
	}
	z.log.Infof("メッセージ未取得：%s %v", code, args)
}

// Warn implements Logger.
func (z *zapLogger) Warn(code string, args ...any) {
	message := z.messageSource.GetMessage(code, args...)
	if message != "" {
		z.log.Warnf("[%s]%s", code, message)
		return
	}
	z.log.Warnf("メッセージ未取得：%s %v", code, args)
}

// WarnWithCodableError implements Logger.
func (z *zapLogger) WarnWithCodableError(err errors.CodableError) {
	code := err.ErrorCode()
	args := err.Args()
	message := z.messageSource.GetMessage(code, args...)
	// エラーのスタックトレース付きのWarnログ出力
	if message != "" {
		z.log.Warnf("[%s]%s, %+v", code, message, err)
		return
	}
	z.log.Warnf("メッセージ未取得：%s %v, %+v", code, args, err)
}

// WarnWithMultiCodableError implements Logger.
func (z *zapLogger) WarnWithMultiCodableError(err errors.MultiCodableError) {
	var logStrs []any
	for _, e := range err.CodableErrors() {
		code := e.ErrorCode()
		args := e.Args()
		message := z.messageSource.GetMessage(code, args...)
		if message != "" {
			logStrs = append(logStrs, fmt.Sprintf("[%s]%s, %+v\n", code, message, e))
		} else {
			logStrs = append(logStrs, fmt.Sprintf("メッセージ未取得：%s %v, %+v\n", code, args, e))
		}
	}
	z.log.Warn(logStrs...)
}

// Error implements Logger.
func (z *zapLogger) Error(code string, args ...any) {
	message := z.messageSource.GetMessage(code, args...)
	if message != "" {
		z.log.Errorf("[%s]%s", code, message)
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
		z.log.Errorf("[%s]%s, %+v", code, message, err)
		return
	}
	z.log.Errorf("メッセージ未取得：%s %v, %+v", code, args, err)
}

// Error implements Logger.
func (z *zapLogger) ErrorWithUnexpectedError(err error) {
	message := z.messageSource.GetMessage(message.E_FW_9999)
	z.log.Errorf("%s, %+v", message, err)
}
