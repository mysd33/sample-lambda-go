// retry パッケージは、リトライ処理を行うためのパッケージです。
package retry

import (
	"math"
	"math/rand"
	"time"
)

const (
	// STOP は、リトライ処理を停止するための定数です。
	STOP time.Duration = -1
)

const (
	// デフォルトのリトライ回数
	DEFAULT_MAX_RETRY_TIMES = 3
	// デフォルトの初回リトライ間隔(100ms)
	DEFAULT_INIT_RETRY_INTERVAL = 100.0 * time.Millisecond
	// デフォルトの最大リトライ間隔(500ms)
	DEFAULT_MAX_RETRY_INTERVAL = 500.0 * time.Millisecond
	// デフォルトのジッター間隔の最大幅(30ms)
	DEFAULT_JITTER_INTERVAL = 30.0 * time.Millisecond
	// 処理開始からのデフォルトの最大経過時間(30秒)
	DEFAULT_MAX_ELAPSED_TIME = 30.0 * time.Second
	// デフォルトのエクスポネンシャルバックオフの乗数
	DEFUALT_MULTIPLIER = 2.0
)

// Backoff は、リトライ処理の間隔を決定するためのインターフェースです。
type Backoff interface {
	// Next は、次のリトライ処理までの間隔を返します。
	Next() time.Duration
}

// NewExponentialBackoff は、デフォルトのエクスポネンシャルバックオフを生成します。
func NewExponentialBackoff() *ExponentialBackoff {
	return &ExponentialBackoff{
		maxRetryTimes:     DEFAULT_MAX_RETRY_TIMES,
		maxInterval:       DEFAULT_MAX_RETRY_INTERVAL,
		maxJitterInterval: DEFAULT_JITTER_INTERVAL,
		maxElapsedTime:    DEFAULT_MAX_ELAPSED_TIME,
		multiplier:        DEFUALT_MULTIPLIER,
		retryTimes:        0,
		nextInterval:      DEFAULT_INIT_RETRY_INTERVAL,
		startTime:         time.Now(),
	}
}

// ExponentialBackoff は、エクスポネンシャルバックオフを実装するBackoffです。
type ExponentialBackoff struct {
	// 最大のリトライ回数
	maxRetryTimes uint
	// 最大のリトライ間隔
	maxInterval time.Duration
	// ジッター間隔の最大幅
	maxJitterInterval time.Duration
	// バックオフ開始からの最大経過時間
	maxElapsedTime time.Duration
	// エクスポネンシャルバックオフの乗数
	multiplier float64
	// 現在のリトライ回数
	retryTimes uint
	// 次のリトライ間隔
	nextInterval time.Duration
	// 処理開始時間
	startTime time.Time
}

// Next implements Backoff.
func (b *ExponentialBackoff) Next() time.Duration {
	// 最大リトライ回数を超えている場合は、リトライ処理を停止
	if b.retryTimes >= b.maxRetryTimes {
		return STOP
	}
	// ジッターを含めたリトライ間隔を取得
	interval := b.getRandomizedInterval(b.nextInterval)
	// リトライ回数をインクリメント
	b.retryTimes++
	// 次のリトライ間隔をあらかじめ計算して設定
	b.setNextInterval()
	// 最大経過時間を超えている場合は、リトライ処理を停止
	if b.isOverMaxElaspedTime() {
		return STOP
	}
	return interval
}

// setNextInterval は、次のリトライ間隔を設定します。
func (b *ExponentialBackoff) setNextInterval() {
	// 次の想定するリトライ間隔を計算（乗数で掛け算）
	calcInterval := time.Duration(float64(b.nextInterval) * b.multiplier)
	// 最大リトライ間隔を超えないように調整
	b.nextInterval = time.Duration(math.Min(float64(calcInterval), float64(b.maxInterval)))
}

// getElapsedTime は、処理開始後の経過時間が最大経過時間を超えているかを返します。
func (b *ExponentialBackoff) isOverMaxElaspedTime() bool {
	elasped := time.Since(b.startTime)
	return b.maxElapsedTime != 0 && elasped > b.maxElapsedTime
}

// getRandomizedInterval は、ジッターを含めたランダムなリトライ間隔を返します。
func (b *ExponentialBackoff) getRandomizedInterval(i time.Duration) time.Duration {
	// 乱数生成
	s := rand.New(rand.NewSource(time.Now().UnixNano()))
	// ジッターを含めた最大[nextInterval + maxJitter]、最小[nextInterval - maxJitter]の時間を計算
	max := float64(i) + float64(b.maxJitterInterval)
	min := float64(i) - float64(b.maxJitterInterval)
	// 乱数に基づきジッターの揺らぎを決定し、実際のリトライ間隔を計算
	return time.Duration(min + ((max - min) * s.Float64()))
}
