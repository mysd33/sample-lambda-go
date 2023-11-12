/*
config パッケージは、設定ファイルを管理するパッケージです。
*/
package config

import (
	"os"

	"example.com/appbase/pkg/constant"
)

type Config interface {
	Get(key string) string
	Reload() error
}

func NewConfig() (Config, error) {
	var cfgs []Config
	if os.Getenv(constant.ENV_NAME) != constant.ENV_LOCAL {
		//クラウド上での実行（Env=Local以外）では、AppConfigから優先的に設定値を取得する
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
	for _, v := range c.cfgs {
		// 最初に見つかった設定値を返却する
		value := v.Get(key)
		if value != "" {
			return value
		}
	}
	return ""
}
