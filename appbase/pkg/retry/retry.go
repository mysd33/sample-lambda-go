// retry パッケージは、リトライ処理を行うためのパッケージです。
package retry

import (
	"context"
	"time"

	"example.com/appbase/pkg/apcontext"
	"example.com/appbase/pkg/logging"
)

// RetryableFunc は、リトライ処理可能な関数です。
type RetryableFunc[T any] func() (T, error)

// CheckRetryable は、リトライ処理可能なエラーかどうかを判定する関数です。
type CheckRetryable[T any] func(result T, err error) bool

// 　Retryer は リトライ処理を行うためのインターフェースです。
type Retryer[T any] interface {
	// Do は、リトライ処理を行います。
	Do(retryableFunc RetryableFunc[T], checkRetryable CheckRetryable[T], opts ...Option) (T, error)
	// DoWithContext は、goroutineを使用する場合のリトライ処理を行います。
	// context.Contextを監視してスリープし、context.Context.Done()を受信するとリトライ終了します。
	DoWithContext(ctx context.Context, retryableFunc RetryableFunc[T], checkRetryable CheckRetryable[T], opts ...Option) (T, error)
}

// defaultRetryer は、リトライ処理を行うためのRetryerのデフォルト実装です。
type defaultRetryer[T any] struct {
	log logging.Logger
}

// NewRetryer は、リトライ処理を行うためのRetryerを生成します。
func NewRetryer[T any](log logging.Logger) Retryer[T] {
	return &defaultRetryer[T]{
		log: log,
	}
}

// Do implements Retryer
func (r *defaultRetryer[T]) Do(retryableFunc RetryableFunc[T], checkRetryable CheckRetryable[T], opts ...Option) (T, error) {
	return r.DoWithContext(apcontext.Context, retryableFunc, checkRetryable, opts...)
}

// DoWithContext implements Retryer
func (r *defaultRetryer[T]) DoWithContext(ctx context.Context, retryableFunc RetryableFunc[T], checkRetryable CheckRetryable[T],
	opts ...Option) (T, error) {
	// エクスポネンシャルバックオフを作成しリトライ処理開始
	eb := NewExponentialBackoff()
	// 指定されたオプションでエクスポネンシャルバックオフ設定上書き
	for _, opt := range opts {
		opt(eb)
	}
	// Contextがnilの場合は、デフォルトのContextを使用
	if ctx == nil {
		ctx = apcontext.Context
	}
	// Context付きのバックオフの作成
	bc := withContext(eb, ctx)

	// タイマーの作成
	t := &defaultTimer{}
	defer t.Stop()

	var err error
	var interval time.Duration
	var result T // 処理結果（ゼロ値）
	// リトライ処理
	for {
		// 対象処理の実行
		if result, err = retryableFunc(); err != nil {
			// 正常終了
			return result, nil
		}
		if !checkRetryable(result, err) {
			// リトライ不能なエラーの場合は、エラーを返却し終了
			return result, err
		}
		// リトライ間隔分待機する
		if interval = bc.Next(); interval == STOP {
			// リトライ終了の場合は、エラーを返却し終了
			return result, err
		}
		// リトライ間隔分待機
		t.Start(interval)
		select {
		// 当該goroutineがキャンセルまたはタイムアウトの場合は、その理由を説明するエラーを返却し終了
		case <-ctx.Done():
			return result, ctx.Err()
		// リトライ間隔の時間経過誤、最初に戻り処理を継続
		case <-t.C():
			r.log.Debug("リトライ回数:%d", eb.retryTimes)
		}
	}
}
