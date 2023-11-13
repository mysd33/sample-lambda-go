/*
config パッケージは、設定ファイルを管理するパッケージです。
*/
package config

import (
	"io"
	"net/http"
	"os"

	"example.com/appbase/pkg/constant"
	"github.com/cockroachdb/errors"
	"gopkg.in/yaml.v3"
)

// appConfigConfigは、AppConfigによるConfig実装です。
type appConfigConfig struct {
	cfg map[string]string
}

// NewAppConfigConfig は、設定ファイルをロードし、viperConfigを作成します。
func newAppConfigConfig() (*appConfigConfig, error) {
	cfg, err := loadAppConfig()
	if err != nil {
		return nil, err
	}
	return &appConfigConfig{cfg: cfg}, nil
}

// Get implements Config.
func (c *appConfigConfig) Get(key string) string {
	v, found := c.getWithContains(key)
	if !found {
		return ""
	}
	return v
}

// getWithContains implements Config.
func (c *appConfigConfig) getWithContains(key string) (string, bool) {
	v, found := c.cfg[key]
	return v, found
}

// Reload implements Config.
func (c *appConfigConfig) Reload() error {
	//ウォームスタート時もリアルタイムに最新の設定を再取得するよう
	//Handlerメソッドの最初で取得するようにする実装しているが
	//init関数のみで各コンポーネント作成時にConfigの値を利用するケースも考えると
	//設定のバージョン不整合が発生してしまう可能性があるため注意が必要
	cfg, err := loadAppConfig()
	if err != nil {
		return err
	}
	c.cfg = cfg
	return nil
}

func loadAppConfig() (map[string]string, error) {
	url := os.Getenv(constant.APPCONFIG_EXTENSION_URL_NAME)
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
	return cfg, nil
}
