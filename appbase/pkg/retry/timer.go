// retry パッケージは、リトライ処理を行うためのパッケージです。
package retry

import "time"

// defaultTimer は、タイマーを実装する構造体です。
type defaultTimer struct {
	timer *time.Timer
}

// C は、タイマーが発火すると現在時刻を受けとるチャネルを返します。
func (t *defaultTimer) C() <-chan time.Time {
	return t.timer.C
}

// durationで指定した期間の後、発火するタイマーを開始します。
func (t *defaultTimer) Start(duration time.Duration) {
	if t.timer == nil {
		t.timer = time.NewTimer(duration)
	} else {
		t.timer.Reset(duration)
	}

}

// タイマーを停止します。
func (t *defaultTimer) Stop() {
	if t.timer != nil {
		t.timer.Stop()
	}
}
