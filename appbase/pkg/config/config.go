/*
config パッケージは、設定ファイルを管理するパッケージです。
*/
package config

import (
	"strconv"

	"example.com/appbase/pkg/env"
	"example.com/appbase/pkg/logging"
	"example.com/appbase/pkg/message"
)

// Config は、設定ファイルを管理するインターフェースです。
type Config interface {
	// GetWithContains は、指定されたキーの設定値をstring型で取得します。OKカンマイディオムにより、値が見つからなかった場合にfalseを返します。
	GetWithContains(key string) (string, bool)
	// Get は、指定されたキーの設定値をstring型で取得します。値が見つからなかった場合には、デフォルト値を返します。
	Get(key string, defaultValue string) string
	// GetIntWithContains は、指定されたキーの設定値をint型で取得します。OKカンマイディオムにより、値が見つからなかった場合にfalseを返します。
	// int変換に失敗した場合もfalseを返します。
	GetIntWithContains(key string) (int, bool)
	// GetInt は、指定されたキーの設定値をint型で取得します。値が見つからなかった場合には、デフォルト値を返します。
	// int変換に失敗した場合デフォルト値を返します。
	GetInt(key string, defaultValue int) int
	// GetBoolWithContains は、指定されたキーの設定値をbool型で取得します。OKカンマイディオムにより、値が見つからなかった場合にfalseを返します。
	// bool変換に失敗した場合もfalseを返します。
	GetBoolWithContains(key string) (bool, bool)
	// GetBool は、指定されたキーの設定値をbool型で取得します。値が見つからなかった場合には、デフォルト値を返します。
	// bool変換に失敗した場合デフォルト値を返します。
	GetBool(key string, defaultValue bool) bool
	// Reload は、設定の取得元よりを最新の設定を再読み込みします。
	Reload() error
}

// NewConfig は、設定ファイルをロードし、Configを作成します。

func NewConfig(log logging.Logger) (Config, error) {
	var cfgs []Config
	if !env.IsLocalOrLocalTest() {
		// ローカル実行（Env=Local,LocalTest）以外では、AppConfigから優先的に設定値を取得する
		ac, err := newAppConfigConfig(log)
		if err != nil {
			return nil, err
		}
		cfgs = append(cfgs, ac)
	}
	// Viperを使って設定ファイルの設定値を取得する
	vc, err := newViperConfig(log)
	if err != nil {
		return nil, err
	}
	cfgs = append(cfgs, vc)

	return &compositeConfig{
		log:  log,
		cfgs: cfgs,
	}, nil
}

// compositeConfigは、複数のConfigをまとめたConfig実装です。
type compositeConfig struct {
	log  logging.Logger
	cfgs []Config
}

// Reload implements Config.
func (c *compositeConfig) Reload() error {
	for _, v := range c.cfgs {
		if err := v.Reload(); err != nil {
			return err
		}
	}
	return nil
}

// GetWithContains implements Config.
func (c *compositeConfig) GetWithContains(key string) (string, bool) {
	for _, v := range c.cfgs {
		value, found := v.GetWithContains(key)
		if found {
			return value, found
		}
	}
	return "", false
}

// Get implements Config.
func (c *compositeConfig) Get(key string, defaultValue string) string {
	value, found := c.GetWithContains(key)
	return returnStringValueIfFound(found, value, defaultValue)
}

// GetIntWithContains implements Config.
func (c *compositeConfig) GetIntWithContains(key string) (int, bool) {
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
func (c *compositeConfig) GetInt(key string, defaultValue int) int {
	value, found := c.GetIntWithContains(key)
	return returnIntValueIfFound(found, value, defaultValue)
}

// GetBoolWithContains implements Config.
func (c *compositeConfig) GetBoolWithContains(key string) (bool, bool) {
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
func (c *compositeConfig) GetBool(key string, defaultValue bool) bool {
	value, found := c.GetBoolWithContains(key)
	return returnBoolValueIfFound(found, value, defaultValue)
}

// returnStringValueIfFound は、値が見つかった場合にその値を返し、見つからなかった場合にデフォルト値を返します。
func returnStringValueIfFound(found bool, value string, defaultValue string) string {
	if found {
		return value
	}
	return defaultValue
}

// convertIntValueIfFound は、値が見つかった場合にその値をint型に変換して返します。見つらない場合はゼロ値(0)を返します。
// int変換に失敗した場合はエラーを返します。
func convertIntValueIfFound(found bool, value string) (int, error) {
	if found {
		intValue, err := strconv.Atoi(value)
		if err != nil {
			return 0, err
		}
		return intValue, nil
	}
	return 0, nil
}

// returnIntValueIfFound は、値が見つかった場合にその値を返し、見つからなかった場合にデフォルト値を返します。
func returnIntValueIfFound(found bool, value int, defaultValue int) int {
	if found {
		return value
	}
	return defaultValue
}

// convertBoolValueIfFound は、値が見つかった場合にその値をbool型に変換して返します。見つらない場合はゼロ値(false)を返します。
// bool変換に失敗した場合はエラーを返します。
func convertBoolValueIfFound(found bool, value string) (bool, error) {
	if found {
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return false, err
		}
		return boolValue, nil
	}
	return false, nil
}

// returnBoolValueIfFound は、値が見つかった場合にその値を返し、見つからなかった場合にデフォルト値を返します。
func returnBoolValueIfFound(found bool, value bool, defaultValue bool) bool {
	if found {
		return value
	}
	return defaultValue
}
