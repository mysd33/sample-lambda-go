/*
config パッケージは、設定ファイルを管理するパッケージです。
*/
package config

import (
	"example.com/appbase/pkg/env"
)

type Config interface {
	Get(key string) string
	getWithContains(key string) (string, bool)
	Reload() error
}

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

// Get implements Config.
func (c *compositeConfig) Get(key string) string {
	value, found := c.getWithContains(key)
	if found {
		return value
	}
	return ""
}

// getWithContains implements Config.
func (c *compositeConfig) getWithContains(key string) (string, bool) {
	for _, v := range c.cfgs {
		value, found := v.getWithContains(key)
		if found {
			return value, found
		}
	}
	return "", false
}
