/*
config パッケージは、設定ファイルを管理するパッケージです。
*/
package config

import (
	"fmt"
	"os"
	"strings"

	"example.com/appbase/pkg/constant"
	"github.com/cockroachdb/errors"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

// viperConfigは、spf13/viperによるConfig実装です。
type viperConfig struct {
}

// NewViperConfig は、設定ファイルをロードし、viperConfigを作成します。
func newViperConfig() (Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	env := os.Getenv(constant.ENV_NAME)
	if env == constant.ENV_LOCAL_TEST {
		// 処理テストコード実行の場合のみパスを相対パスに変更
		// 環境ごとのConfigを読み取る
		viper.AddConfigPath(fmt.Sprintf("../../../configs/%s/", strings.ToLower(env)))
	} else {
		// 環境ごとのConfigを読み取る
		viper.AddConfigPath(fmt.Sprintf("configs/%s/", strings.ToLower(env)))
	}
	// 環境変数がすでに指定されてる場合はそちらを優先させる
	viper.AutomaticEnv()
	// 環境変数の値が空列の場合も優先して扱う
	viper.AllowEmptyEnv(true)

	if err := viper.ReadInConfig(); err != nil {
		return nil, errors.Errorf("設定ファイル読み込みエラー:%w", err)
	}
	return &viperConfig{}, nil
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
	v := viper.Get(key)
	if v == nil {
		return "", false
	}
	return cast.ToString(v), true
}

// Reload implements Config.
func (c *viperConfig) Reload() error {
	// 何もしない
	return nil
}
