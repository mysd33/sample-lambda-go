/*
logging パッケージは、ログ出力に関する機能を提供するパッケージです。
*/
package logging

import (
	"fmt"
	"os"

	"example.com/appbase/pkg/env"
	"example.com/appbase/pkg/errors"
	"example.com/appbase/pkg/message"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	LOG_LEVEL_NAME  = "LOG_LEVEL"
	LOG_FORMAT_NAME = "LOG_FORMAT"
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
	// WarnWithError は、エラーおよび、メッセージID（エラーコードcode）、置き換え文字列argsに対応するメッセージで警告レベルのログを出力します。codeに対応するメッセージがない場合はそのまま出力します。
	WarnWithError(err error, code string, args ...any)
	// WarnWithCodableError は、エラーが持つメッセージID（エラーコード）、置き換え文字列に対応するメッセージで警告レベルのログを出力します。codeに対応するメッセージがない場合はそのまま出力します。
	WarnWithCodableError(err errors.CodableError)
	// WarnWithMultiError は、複数のエラーを持つエラーを警告レベルのログに出力します。
	WarnWithMultiCodableError(err errors.MultiCodableError)
	// Error は、メッセージID（エラーコードcode）、置き換え文字列argsに対応するメッセージでエラーレベルのログを出力します。codeに対応するメッセージがない場合はそのまま出力します。
	Error(code string, args ...any)
	// ErrorWithCodableError は、エラーおよび、メッセージID（エラーコード）、置き換え文字列に対応するメッセージでエラーレベルのログを出力します。codeに対応するメッセージがない場合はそのまま出力します。
	ErrorWithError(err error, code string, args ...any)
	// ErrorWithCodableError は、エラーが持つメッセージID（エラーコード）、置き換え文字列に対応するメッセージでエラーレベルのログを出力します。codeに対応するメッセージがない場合はそのまま出力します。
	ErrorWithCodableError(err errors.CodableError)
	// ErrorWithUnexpectedError は、予期せぬエラーをログに出力します。
	ErrorWithUnexpectedError(err error)
	// Sync は、バッファリングされたログをフラッシュします。Zapによるデフォルトの出力（標準エラー出力）ではバッファリングを実施していませんが、
	// ZapのAPIでカスタマイズすることで、バッファリングする出力に変更することも可能なので、呼び出さないとログが完全に出力されないことがあります。
	// プロセス終了の最後にこのメソッドを呼ぶことを推奨します。本ソフトウェアフレームワークを利用すると、
	// AP処理実行制御機能がLambdaのハンドラ関数の最後に自動的に必ず呼び出すようになっています。
	Sync() error
}

// NewLogger は、Loggerを作成します。
func NewLogger(messageSource message.MessageSource) (Logger, error) {
	var config zap.Config
	if env.IsStragingOrProd() {
		// 本番相当の環境の場合
		config = zap.NewProductionConfig()
		// ログの時刻をISO8601形式での出力に変更
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		// 開発環境の場合
		config = zap.NewDevelopmentConfig()
	}
	// 出力先を標準出力（バッファリング）に設定
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stdout"}

	// 個別にログレベルが設定されている場合はログレベル上書き
	if level, found := os.LookupEnv(LOG_LEVEL_NAME); found {
		if al, err := zap.ParseAtomicLevel(level); err == nil {
			config.Level = al
		}
	}
	// 個別のログフォーマットが設定されている場合はログフォーマット上書き
	if format, found := os.LookupEnv(LOG_FORMAT_NAME); found {
		if format == "json" || format == "console" {
			config.Encoding = format
		}
	}
	// ログのスタックトレースを無効化
	config.DisableStacktrace = true
	// ZapのLoggerをラップしているため、ログの呼び出し階層を調整
	z, err := config.Build(zap.AddCallerSkip(1))
	if err != nil {
		return nil, err
	}
	// SugarLoggerを使って、ログ出力を行う
	sugerredLogger := z.Sugar()
	return &zapLogger{
		originalLog:   sugerredLogger, // AddInfoで付加情報を追加した場合に元のロガーに戻すため保持
		log:           sugerredLogger, // 実際にログ出力するためのロガー
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

// WarnWithError implements Logger.
func (z *zapLogger) WarnWithError(err error, code string, args ...any) {
	message := z.messageSource.GetMessage(code, args...)
	// エラーのスタックトレース付きのWarnログ出力
	if message != "" {
		z.log.Warnf("[%s]%s, %+v", code, message, err)
		return
	}
	z.log.Warnf("メッセージ未取得：%s %v, %+v", code, args, err)
}

// WarnWithCodableError implements Logger.
func (z *zapLogger) WarnWithCodableError(err errors.CodableError) {
	code := err.ErrorCode()
	args := err.Args()
	z.WarnWithError(err, code, args...)
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

// ErrorWithError implements Logger.
func (z *zapLogger) ErrorWithError(err error, code string, args ...any) {
	message := z.messageSource.GetMessage(code, args...)
	// エラーのスタックトレース付きのErrorログ出力
	if message != "" {
		z.log.Errorf("[%s]%s, %+v", code, message, err)
		return
	}
	z.log.Errorf("メッセージ未取得：%s %v, %+v", code, args, err)
}

// ErrorWithCodableError implements Logger.
func (z *zapLogger) ErrorWithCodableError(err errors.CodableError) {
	code := err.ErrorCode()
	args := err.Args()
	z.ErrorWithError(err, code, args...)
}

// Error implements Logger.
func (z *zapLogger) ErrorWithUnexpectedError(err error) {
	message := z.messageSource.GetMessage(message.E_FW_9999)
	z.log.Errorf("%s, %+v", message, err)
}

// Sync implements Logger.
func (z *zapLogger) Sync() error {
	return z.log.Sync()
}
