/*
config パッケージは、設定ファイルを管理するパッケージです。
*/
package config

import (
	"io"
	"net/http"
	"os"

	"github.com/cockroachdb/errors"
	"gopkg.in/yaml.v3"
)

// appConfigConfigは、AppConfigによるConfig実装です。
type appConfigConfig struct {
	cfg map[string]string
}

// NewAppConfigConfig は、設定ファイルをロードし、viperConfigを作成します。
func newAppConfigConfig() (*appConfigConfig, error) {
	//コールドスタート（init）時のみに、AppConfigから最新取得する実装としている
	//ウォームスタート時もリアルタイムに最新取得したい場合はHandlerメソッドの最初で取得するようにする必要がある
	url := os.Getenv("AppConfigExtensionsURL")
	// AppConfig Lambda Extensionsのエンドポイントへアクセスして設定データを取得
	response, err := http.Get(url)
	if err != nil {
		return nil, errors.Errorf("AppConfig読み込みエラー:%w", err)
	}
	var cfg map[string]string
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Errorf("AppConfig読み込みエラー:%w", err)
	}
	// YAMLの設定データを読み込み
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, errors.Errorf("AppConfig読み込みエラー:%w", err)
	}
	return &appConfigConfig{cfg: cfg}, nil
}

// Get implements Config.
func (c *appConfigConfig) Get(key string) string {
	v, found := c.cfg[key]
	if !found {
		return ""
	}
	return v
}
