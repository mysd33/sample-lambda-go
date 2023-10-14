/*
logging パッケージは、ログ出力に関する機能を提供するパッケージです。
*/
package logging

import (
	"go.uber.org/zap"
)

//TODO: ログレベルの設定

// Loggerは、ログ出力のインタフェースです
type Logger interface {
	// Debugは、デバッグレベルのログを出力します。
	Debug(template string, args ...interface{})
	// Infoは、情報レベルのログを出力します。
	Info(template string, args ...interface{})
	// Warnは、警告レベルのログを出力します。
	Warn(template string, args ...interface{})
	// Errorは、エラーレベルのログを出力します。
	Error(template string, args ...interface{})
	// Fatailは、致命的なエラーレベルのログを出力します。
	Fatal(template string, args ...interface{})
}

// NewLogger は、Loggerを作成します。
func NewLogger() Logger {
	z, _ := zap.NewProduction()
	return ZapLogger{log: z.Sugar()}
}

// ZapLoggerは、Zapを使ったLogger実装です。
type ZapLogger struct {
	log *zap.SugaredLogger
}

func (z ZapLogger) Debug(template string, args ...interface{}) {
	z.log.Debugf(template, args)
}

func (z ZapLogger) Info(template string, args ...interface{}) {
	z.log.Infof(template, args)
}

func (z ZapLogger) Warn(template string, args ...interface{}) {
	z.log.Warnf(template, args)
}

func (z ZapLogger) Error(template string, args ...interface{}) {
	z.log.Errorf(template, args)
}

func (z ZapLogger) Fatal(template string, args ...interface{}) {
	z.log.Fatalf(template, args)
}
