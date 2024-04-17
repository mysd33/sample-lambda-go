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

// NewApplicationContext()のベンチマークテスト
func BenchmarkNewApplicationContext(b *testing.B) {
	env.SetTestEnvForBechMark(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewApplicationContext() //　285           4359328 ns/op          209031 B/op       2145 allocs/op
	}
}

// NewStatisticsApplicationContext()のベンチマークテスト
func BenchmarkNewStatisticsApplicationContext(b *testing.B) {
	env.SetTestEnvForBechMark(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewStatisticsApplicationContext() //  1988            588965 ns/op           51407 B/op        597 allocs/op
	}
}

// NewAuthorizerApplicationContext()のベンチマークテスト
func BenchmarkNewAuthorizerApplicationContext(b *testing.B) {
	env.SetTestEnvForBechMark(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewAuthorizerApplicationContext() //  763           1503269 ns/op           82770 B/op        864 allocs/op
	}
}

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

	// NewApplicationContext()処理時間: 180.423ms
	// → ログが出力される際に、Zapの初回ログ出力の際に遅くなっているものと思われる
	// NewStatisticsApplicationContext()処理時間: 548.6µs
	// NewAuthorizerApplicationContext()処理時間: 1.9484ms
}
