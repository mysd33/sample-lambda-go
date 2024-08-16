/*
config パッケージは、設定ファイルを管理するパッケージです。
*/
package config

import (
	"fmt"
	"os"
	"strings"

	"example.com/appbase/pkg/env"
	"example.com/appbase/pkg/logging"
	"example.com/appbase/pkg/message"
	"github.com/cockroachdb/errors"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

const (
	CONFIG_BASE_PATH_NAME = "CONFIG_BASE_PATH" // 設定ファイルのパスを明示的に指定する場合のOS環境変数名
)

const (
	configFileName  = "config-%s"         // 設定ファイル名
	extension       = "yaml"              // 設定ファイルの拡張子
	configsPath     = "configs/"          // 設定ファイルのパス
	testConfigsPath = "../../../configs/" // 処理テスト実行時の設定ファイルのパス
)

// viperConfigは、spf13/viperによるConfig実装です。
type viperConfig struct {
	log logging.Logger
}

// NewViperConfig は、設定ファイルをロードし、viperConfigを作成します。
func newViperConfig(log logging.Logger) (Config, error) {
	viper.SetConfigName(fmt.Sprintf(configFileName, strings.ToLower(env.GetEnv())))
	viper.SetConfigType(extension)
	if configBasePath, found := os.LookupEnv(CONFIG_BASE_PATH_NAME); found {
		// 環境変数の定義があればそれをベースパスとしてのConfigを読み取る
		viper.AddConfigPath(fmt.Sprintf("%s/", strings.TrimRight(configBasePath, "/")))
	} else if env.IsLocalTest() {
		// テストコード実行の場合、テストコードからの相対パスに変更
		// 環境ごとのConfigを読み取る
		viper.AddConfigPath(testConfigsPath)
	} else {
		// 環境ごとのConfigを読み取る
		viper.AddConfigPath(configsPath)
	}
	// 環境変数がすでに指定されてる場合はそちらを優先させる
	viper.AutomaticEnv()
	// 環境変数の値が空列の場合も優先して扱う
	viper.AllowEmptyEnv(true)

	if err := viper.ReadInConfig(); err != nil {
		return nil, errors.Errorf("設定ファイル読み込みエラー:%w", err)
	}
	return &viperConfig{log: log}, nil
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
	result, err := convertIntValueIfFound(found, value)
	if err != nil {
		c.log.WarnWithError(err, message.W_FW_8009, key, value)
		// int変換に失敗した場合は、値が見つからなかったとしてfalseを返す
		return 0, false
	}
	return result, found
}

// GetInt implements Config.
func (c *viperConfig) GetInt(key string, defaultValue int) int {
	value, found := c.GetIntWithContains(key)
	return returnIntValueIfFound(found, value, defaultValue)
}

// GetBoolWithContains implements Config.
func (c *viperConfig) GetBoolWithContains(key string) (bool, bool) {
	value, found := c.GetWithContains(key)
	result, err := convertBoolValueIfFound(found, value)
	if err != nil {
		c.log.WarnWithError(err, message.W_FW_8010, key, value)
		// bool変換に失敗した場合は、値が見つからなかったとしてfalseを返す
		return false, false
	}
	return result, found
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
