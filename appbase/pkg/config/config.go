/*
config パッケージは、設定ファイルを管理するパッケージです。
*/
package config

import (
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/spf13/viper"
)

// Config は、設定ファイルの構造体(Viper)です。
type Config struct {
	Hoge Hoge `yaml:"hoge"`
}

// TODO: とりあえずのサンプル設定
type Hoge struct {
	Name string `yaml:"name"`
}

// LoadConfig は、設定ファイルをロードします。
func LoadConfig() (*Config, error) {
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
	var cfg Config
	err = viper.Unmarshal(&cfg)
	if err != nil {
		return nil, errors.Errorf("設定ファイルアンマーシャルエラー:%w", err)
	}
	return &cfg, nil
}
