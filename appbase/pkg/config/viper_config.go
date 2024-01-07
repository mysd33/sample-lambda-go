/*
config パッケージは、設定ファイルを管理するパッケージです。
*/
package config

import (
	"fmt"
	"strings"

	"example.com/appbase/pkg/env"
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
	if env.IsLocalTest() {
		// テストコード実行の場合のみパスを相対パスに変更
		// 環境ごとのConfigを読み取る
		viper.AddConfigPath(fmt.Sprintf("../../../configs/%s/", strings.ToLower(env.GetEnv())))
	} else {
		// 環境ごとのConfigを読み取る
		viper.AddConfigPath(fmt.Sprintf("configs/%s/", strings.ToLower(env.GetEnv())))
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

// GetWithContains implements Config.
func (c *viperConfig) GetWithContains(key string) (string, bool) {
	v := viper.Get(key)
	if v == nil {
		return "", false
	}
	return cast.ToString(v), true
}

// Get implements Config.
func (c *viperConfig) Get(key string, defaultValue string) string {
	value, found := c.GetWithContains(key)
	return returnStringValueIfFound(found, value, defaultValue)
}

// GetIntWithContains implements Config.
func (c *viperConfig) GetIntWithContains(key string) (int, bool) {
	value, found := c.GetWithContains(key)
	// int変換に失敗した場合は、値が見つからなかったとしてfalseを返す
	return returnIntValue(found, value)
}

// GetInt implements Config.
func (c *viperConfig) GetInt(key string, defaultValue int) int {
	value, found := c.GetIntWithContains(key)
	return returnIntValueIfFound(found, value, defaultValue)
}

// GetBoolWithContains implements Config.
func (c *viperConfig) GetBoolWithContains(key string) (bool, bool) {
	value, found := c.GetWithContains(key)
	// bool変換に失敗した場合は、値が見つからなかったとしてfalseを返す
	return returnBoolValue(found, value)
}

// GetBool implements Config.
func (c *viperConfig) GetBool(key string, defaultValue bool) bool {
	value, found := c.GetBoolWithContains(key)
	return returnBoolValueIfFound(found, value, defaultValue)
}

// Reload implements Config.
func (c *viperConfig) Reload() error {
	// 何もしない
	return nil
}
