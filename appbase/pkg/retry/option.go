// retry パッケージは、リトライ処理を行うためのパッケージです。
package retry

import "time"

// エクスポネンシャルバックオフのオプションを設定するための関数です。
type Option func(*ExponentialBackoff)

// MaxRetryTimesは、最大のリトライ回数を設定します。
// デフォルトは3回です。
func MaxRetryTimes(mt uint) Option {
	return func(b *ExponentialBackoff) {
		b.maxRetryTimes = mt
	}
}

// MaxIntervalは、最大のリトライ間隔を設定します。
// デフォルトは500msです。
func MaxInterval(mi time.Duration) Option {
	return func(b *ExponentialBackoff) {
		b.maxInterval = mi
	}
}

// Intervalは、初回のリトライ間隔を設定します。
// デフォルトは100msです。
func Interval(i time.Duration) Option {
	return func(b *ExponentialBackoff) {
		b.nextInterval = i
	}
}

// JitterIntervalは、ジッター間隔の最大幅を設定します。
// デフォルトは30msです。
func JitterInterval(ji time.Duration) Option {
	return func(b *ExponentialBackoff) {
		b.maxJitterInterval = ji
	}
}

// Multiplierは、エクスポネンシャルバックオフの乗数を設定します。
func Multiplier(m float64) Option {
	return func(b *ExponentialBackoff) {
		b.multiplier = m
	}
}

// MaxElapsedTimeは、処理開始からの最大経過時間を設定します。
func MaxElapsedTime(et time.Duration) Option {
	return func(b *ExponentialBackoff) {
		b.maxElapsedTime = et
	}
}
