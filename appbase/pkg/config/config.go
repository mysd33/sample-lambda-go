package config

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// 設定ファイルの構造体(Viper)
type Config struct {
	Hoge Hoge `yaml:"hoge"`
}

// TODO: とりあえずのサンプル
type Hoge struct {
	Name string `yaml:"name"`
}

// 設定ファイルのロード
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
		return nil, errors.Errorf("設定ファイル読み込みエラー")
	}
	var cfg Config
	err = viper.Unmarshal(&cfg)
	if err != nil {
		return nil, errors.Errorf("設定ファイルアンマーシャルエラー")
	}
	return &cfg, nil
}
