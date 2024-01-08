/*
config パッケージは、設定ファイルを管理するパッケージです。
*/
package config

import (
	"strconv"

	"example.com/appbase/pkg/env"
)

// Config は、設定ファイルを管理するインターフェースです。
type Config interface {
	GetWithContains(key string) (string, bool)
	Get(key string, defaultValue string) string
	GetIntWithContains(key string) (int, bool)
	GetInt(key string, defaultValue int) int
	GetBoolWithContains(key string) (bool, bool)
	GetBool(key string, defaultValue bool) bool
	Reload() error
}

// NewConfig は、設定ファイルをロードし、Configを作成します。

func NewConfig() (Config, error) {
	var cfgs []Config
	if !env.IsLocalOrLocalTest() {
		// ローカル実行（Env=Local,LocalTest）以外では、AppConfigから優先的に設定値を取得する
		ac, err := newAppConfigConfig()
		if err != nil {
			return nil, err
		}
		cfgs = append(cfgs, ac)
	}
	// Viperを使って設定ファイルの設定値を取得する
	vc, err := newViperConfig()
	if err != nil {
		return nil, err
	}
	cfgs = append(cfgs, vc)

	return &compositeConfig{cfgs: cfgs}, nil
}

// compositeConfigは、複数のConfigをまとめたConfig実装です。
type compositeConfig struct {
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
	// int変換に失敗した場合は、値が見つからなかったとしてfalseを返す
	return returnIntValue(found, value)
}

// GetInt implements Config.
func (c *compositeConfig) GetInt(key string, defaultValue int) int {
	value, found := c.GetIntWithContains(key)
	return returnIntValueIfFound(found, value, defaultValue)
}

// GetBoolWithContains implements Config.
func (c *compositeConfig) GetBoolWithContains(key string) (bool, bool) {
	value, found := c.GetWithContains(key)
	// bool変換に失敗した場合は、値が見つからなかったとしてfalseを返す
	return returnBoolValue(found, value)
}

// GetBool implements Config.
func (c *compositeConfig) GetBool(key string, defaultValue bool) bool {
	value, found := c.GetBoolWithContains(key)
	return returnBoolValueIfFound(found, value, defaultValue)
}

func returnStringValueIfFound(found bool, value string, defaultValue string) string {
	if found {
		return value
	}
	return defaultValue
}

func returnIntValue(found bool, value string) (int, bool) {
	if found {
		intValue, err := strconv.Atoi(value)
		if err != nil {

			return 0, false
		}
		return intValue, true
	}
	return 0, false
}

func returnIntValueIfFound(found bool, value int, defaultValue int) int {
	if found {
		return value
	}
	return defaultValue
}

func returnBoolValue(found bool, value string) (bool, bool) {
	if found {
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return false, false
		}
		return boolValue, true
	}
	return false, false
}

func returnBoolValueIfFound(found bool, value bool, defaultValue bool) bool {
	if found {
		return value
	}
	return defaultValue
}
