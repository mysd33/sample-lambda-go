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

// ライトウェイトなApplicationContextを作成する処理時間の測定テスト
func TestNewApplicationContextSpeed(t *testing.T) {
	env.SetTestEnv(t)

	start := time.Now()
	NewApplicationContext()
	fmt.Println("NewApplicationContext()処理時間:", time.Since(start))
	start = time.Now()
	NewStatisticsApplicationContext()
	fmt.Println("NewStatisticsApplicationContext()処理時間:", time.Since(start))
	start = time.Now()
	NewAuthorizerApplicationContext()
	fmt.Println("NewAuthorizerApplicationContext()処理時間:", time.Since(start))

	// NewApplicationContext()処理時間: 171.5675ms
	// 統計向けは、ほとんどの機能が不要なので、処理時間がかなり変わり、効果あり
	// NewStatisticsApplicationContext()処理時間: 535.3µs
	// API認可向けだと、あまり減らせる機能はないので、あまりかわらなそう、効果薄い
	// NewAuthorizerApplicationContext()処理時間: 1.6448ms
}
