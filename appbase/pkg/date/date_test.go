// date パッケージは、システム日時を扱うための機能を提供します。

package date

import (
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"
	_ "unsafe"

	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/logging"
	"example.com/appbase/pkg/message"
	"github.com/stretchr/testify/assert"
)

func Test_GetSystemDate_TestDate(t *testing.T) {
	setUp(t)
	testDateStr := "2021/01/01 00:00:00"
	config := config.NewTestConfig(map[string]string{
		"TEST_DATE": testDateStr,
	})
	msg, _ := message.NewMessageSource()
	logger, _ := logging.NewLogger(msg, config)
	d := NewDateManager(config, logger)
	expected, _ := time.ParseInLocation(format, testDateStr, time.Local)
	actual := d.GetSystemDate()
	fmt.Println(actual)
	assert.WithinDuration(t, expected, actual, 1*time.Second)
}

func Test_GetSystemDate_Now(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Setenv("TZ", "Asia/Tokyo")
		t.Cleanup(resetLocalOnce)
		fmt.Printf("%s\n", time.Local.String())
	}
	config := config.NewTestConfig(map[string]string{})
	msg, _ := message.NewMessageSource()
	logger, _ := logging.NewLogger(msg, config)
	d := NewDateManager(config, logger)

	expected := time.Now()
	actual := d.GetSystemDate()
	fmt.Println(actual)
	assert.WithinDuration(t, expected, actual, 1*time.Second)
}

func setUp(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Setenv("TZ", "Asia/Tokyo")
		t.Cleanup(resetLocalOnce)
		fmt.Printf("%s\n", time.Local.String())
	}
}

//go:linkname localOnce sync.localOnce
var localOnce sync.Once

func resetLocalOnce() {
	localOnce = sync.Once{}
}
