// date パッケージは、システム日時を扱うための機能を提供します。
package date

import (
	"time"

	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/logging"
	"example.com/appbase/pkg/message"
)

const (
	// テスト日時
	TEST_DATE_NAME = "TEST_DATE"
	format         = "2006/01/02 15:04:05"
)

// DateManager は、システム日時を取得するためのインタフェースです。
type DateManager interface {
	// Now は、現在のシステム日時を取得します。
	GetSystemDate() time.Time
}

// defaultDateManager は、DateManagerのデフォルト実装の構造体です。
type defaultDateManager struct {
	cfg config.Config
	log logging.Logger
}

// NewDateManager は、DateManagerを作成します。
func NewDateManager(cfg config.Config, log logging.Logger) DateManager {
	return &defaultDateManager{
		cfg: cfg,
		log: log,
	}
}

// GetSystemDate implements DateManager
func (d *defaultDateManager) GetSystemDate() time.Time {
	// テスト用の日付が設定されている場合は、その日付を返す
	if now, ok := d.cfg.GetWithContains(TEST_DATE_NAME); ok {
		d.log.Debug("テスト時刻: %s", now)
		t, err := time.ParseInLocation(format, now, time.Local)
		if err != nil {
			d.log.WarnWithError(err, message.W_FW_8004, now)
		}
		return t
	}
	// 通常は、現在のローカル日時を返す
	return time.Now()
}
