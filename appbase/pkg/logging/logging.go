package logging

import (
	"go.uber.org/zap"
)

//TODO: ログレベルの設定

type Logger interface {
	Debug(template string, args ...interface{})
	Info(template string, args ...interface{})
	Warn(template string, args ...interface{})
	Error(template string, args ...interface{})
	Fatal(template string, args ...interface{})
}

func NewLogger() Logger {
	z, _ := zap.NewProduction()
	return ZapLogger{Log: z.Sugar()}
}

type ZapLogger struct {
	Log *zap.SugaredLogger
}

func (z ZapLogger) Debug(template string, args ...interface{}) {
	z.Log.Debugf(template, args)
}

func (z ZapLogger) Info(template string, args ...interface{}) {
	z.Log.Infof(template, args)
}

func (z ZapLogger) Warn(template string, args ...interface{}) {
	z.Log.Warnf(template, args)
}

func (z ZapLogger) Error(template string, args ...interface{}) {
	z.Log.Errorf(template, args)
}

func (z ZapLogger) Fatal(template string, args ...interface{}) {
	z.Log.Fatalf(template, args)
}
