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

func BenchmarkNewApplicationContext(b *testing.B) {
	env.SetTestEnvForBechMark(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewApplicationContext() //　285           4359328 ns/op          209031 B/op       2145 allocs/op
	}
}

func BenchmarkNewStatisticsApplicationContext(b *testing.B) {
	env.SetTestEnvForBechMark(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewStatisticsApplicationContext() //  1988            588965 ns/op           51407 B/op        597 allocs/op
	}
}

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
	// →ひょっとしたら、RDB関連の初期化処理（初回、_ "github.com/lib/pq"で動作する処理）が重いかもしれないので
	//  ベンチマークで何回も呼んでると、平均が短くなるのかもしれない
	//
	// ほとんどの不要な機能のインタンス化を減らすと、処理時間がかなり変わり効果あり
	// NewStatisticsApplicationContext()処理時間: 548.6µs
	// NewAuthorizerApplicationContext()処理時間: 1.9484ms
}
