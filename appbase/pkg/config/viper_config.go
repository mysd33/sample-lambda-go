/*
config パッケージは、設定ファイルを管理するパッケージです。
*/
package config

import (
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/spf13/viper"
)

// viperConfigは、spf13/viperによるConfig実装です。
type viperConfig struct {
	cfg map[string]string
}

// NewViperConfig は、設定ファイルをロードし、viperConfigを作成します。
func newViperConfig() (*viperConfig, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("configs/")
	// 環境変数がすでに指定されてる場合はそちらを優先させる
	viper.AutomaticEnv()
	// データ構造をキャメルケースに切り替える用の設定
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		return nil, errors.Errorf("設定ファイル読み込みエラー:%w", err)
	}
	var cfg map[string]string
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, errors.Errorf("設定ファイルアンマーシャルエラー:%w", err)
	}
	return &viperConfig{cfg: cfg}, nil
}

// Get implements Config.
func (c *viperConfig) Get(key string) string {
	v, found := c.cfg[key]
	if !found {
		return ""
	}
	return v
}
