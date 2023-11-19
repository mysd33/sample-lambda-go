/*
config パッケージは、設定ファイルを管理するパッケージです。
*/
package config

import (
	"os"
	"strings"

	"example.com/appbase/pkg/constant"
	"github.com/cockroachdb/errors"
	"github.com/spf13/viper"
)

// viperConfigは、spf13/viperによるConfig実装です。
type viperConfig struct {
	cfg map[string]string
}

// NewViperConfig は、設定ファイルをロードし、viperConfigを作成します。
func newViperConfig() (Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	if os.Getenv(constant.ENV_NAME) == constant.ENV_LOCAL_TEST {
		// 処理テストコード実行の場合のみパスを相対パスに変更
		viper.AddConfigPath("../../../configs/")
	} else {
		viper.AddConfigPath("configs/")
	}
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
	v, found := c.getWithContains(key)
	if !found {
		return ""
	}
	return v
}

// getWithContains implements Config.
func (c *viperConfig) getWithContains(key string) (string, bool) {
	// Viperは大文字小文字を区別しないのでkeyを一旦小文字にして検索している
	v, found := c.cfg[strings.ToLower(key)]
	return v, found
}

// Reload implements Config.
func (c *viperConfig) Reload() error {
	// 何もしない
	return nil
}
