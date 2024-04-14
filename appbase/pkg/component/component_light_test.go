/*
component パッケージはフレームワークのコンポーネントのインスタンスをDIし管理するパッケージです。
*/

package component

import (
	"fmt"
	"testing"
	"time"

	"example.com/appbase/pkg/env"
)

func TestNewApplicationContextSpeed(t *testing.T) {
	env.SetTestEnv(t)

	start := time.Now()
	NewApplicationContext()
	fmt.Println("NewApplicationContext()処理時間:", time.Since(start))
	start = time.Now()
	NewLightWeightApplicationContext()
	fmt.Println("NewLightWeightApplicationContext()処理時間:", time.Since(start))

	// 不要なインスタンス生成は避けると、処理時間がかなり変わり、効果あり
	// NewApplicationContext()処理時間: 193.6718ms
	// NewLightWeightApplicationContext()処理時間: 541.1µs
}
