/*
config パッケージは、設定ファイルを管理するパッケージです。
*/
package config

import (
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/spf13/viper"
)

type Config interface {
	Get(key string) string
}

func NewConfig() (Config, error) {
	//TODO: Env=Localの場合はviperでconfigファイルからとるようにする
	//それ以外の場合は、Envに合わせたAppConfigから値をとるように作る
	return loadViperConfig()
}

type viperConfig struct {
	cfg map[string]string
}

func (c *viperConfig) Get(key string) string {
	v, found := c.cfg[key]
	if !found {
		return ""
	}
	return v
}

// LoadConfig は、設定ファイルをロードします。
func loadViperConfig() (*viperConfig, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("configs/")
	// 環境変数がすでに指定されてる場合はそちらを優先させる
	viper.AutomaticEnv()
	// データ構造をキャメルケースに切り替える用の設定
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	err := viper.ReadInConfig()
	if err != nil {
		return nil, errors.Errorf("設定ファイル読み込みエラー:%w", err)
	}
	var cfg map[string]string
	err = viper.Unmarshal(&cfg)
	if err != nil {
		return nil, errors.Errorf("設定ファイルアンマーシャルエラー:%w", err)
	}
	return &viperConfig{cfg: cfg}, nil
}
