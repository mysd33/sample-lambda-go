// retry パッケージは、リトライ処理を行うためのパッケージです。
package retry

import (
	"context"
	"time"
)

// backoffWithContext は、Contextを持ったバックオフ情報を生成します。
type backoffWithContext interface {
	Backoff
	context() context.Context
}

// withContext は、Context付きのバックオフ情報を生成します。
func withContext(eb *ExponentialBackoff, ctx context.Context) backoffWithContext {
	return &defaultBackoffWithContext{
		eb:  eb,
		ctx: ctx,
	}
}

// defaultBackoffWithContext は、BackoffContextのデフォルト実装です。
type defaultBackoffWithContext struct {
	ctx context.Context
	eb  *ExponentialBackoff
}

// Next implements Backoff.
func (c *defaultBackoffWithContext) Next() time.Duration {
	select {
	// Contextがキャンセルされた場合は、リトライ処理を停止する
	case <-c.ctx.Done():
		return STOP
	default:
	}
	// 次のリトライ処理までの間隔を返す
	next := c.eb.Next()
	// Contextがタイムアウト設定されていて、次のリトライ処理までの間隔がない場合は、リトライ処理を停止する
	if deadline, ok := c.ctx.Deadline(); ok && time.Until(deadline) < next {
		return STOP
	}
	return next
}

// context implements backoffWithContext
func (c *defaultBackoffWithContext) context() context.Context {
	return c.ctx
}
