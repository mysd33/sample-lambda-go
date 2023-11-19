package testing

import (
	"os"

	"example.com/appbase/pkg/constant"
)

// IsTestRunning はテスト実行中か判定します
func IsTestRunning() bool {
	return os.Getenv(constant.ENV_NAME) == ""

	// TODO: flag.Lookup("test.v")を使いたいが
	// flag.Parse()がmain.goのinitより先に実行されない？
	//return flag.Lookup("test.v") != nil
}
