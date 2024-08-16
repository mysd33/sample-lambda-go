/*
config パッケージは、設定ファイルを管理するパッケージです。
*/
package config

import (
	"encoding/json"
	"io"
	"maps"
	"net/http"
	"os"

	"example.com/appbase/pkg/logging"
	"example.com/appbase/pkg/message"
	"github.com/cockroachdb/errors"
	"gopkg.in/yaml.v3"
)

const (
	APPCONFIG_HOSTED_EXTENSION_URL_NAME = "APPCONFIG_HOSTED_EXTENSION_URL"
	APPCONFIG_SM_EXTENSION_URL_NAME     = "APPCONFIG_SM_EXTENSION_URL"
)

// appConfigConfigは、AWS AppConfigによるConfig実装です。
type appConfigConfig struct {
	log logging.Logger
	cfg map[string]string
}

// NewAppConfigConfig は、AWS AppConfigから設定をロードする、Configを作成します。
func newAppConfigConfig(log logging.Logger) (Config, error) {
	cfg, err := loadAppConfigConfig(log)
	if err != nil {
		return nil, err
	}
	return &appConfigConfig{log: log, cfg: cfg}, nil
}

// GetWithContains implements Config.
func (c *appConfigConfig) GetWithContains(key string) (string, bool) {
	v, found := c.cfg[key]
	return v, found
}

// Get implements Config.
func (c *appConfigConfig) Get(key string, defaultValue string) string {
	value, found := c.GetWithContains(key)
	return returnStringValueIfFound(found, value, defaultValue)
}

// GetIntWithContains implements Config.
func (c *appConfigConfig) GetIntWithContains(key string) (int, bool) {
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
func (c *appConfigConfig) GetInt(key string, defaultValue int) int {
	value, found := c.GetIntWithContains(key)
	return returnIntValueIfFound(found, value, defaultValue)
}

// GetBoolWithContains implements Config.
func (c *appConfigConfig) GetBoolWithContains(key string) (bool, bool) {
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
func (c *appConfigConfig) GetBool(key string, defaultValue bool) bool {
	value, found := c.GetBoolWithContains(key)
	return returnBoolValueIfFound(found, value, defaultValue)
}

// Reload implements Config.
func (c *appConfigConfig) Reload() error {
	//ウォームスタート時もリアルタイムに最新の設定を再取得するよう
	//Handlerメソッドの最初で取得するようにする実装しているが
	//init関数のみで各コンポーネント作成時にConfigの値を利用するケースも考えると
	//設定のバージョン不整合が発生してしまう可能性があるため注意が必要
	cfg, err := loadAppConfigConfig(c.log)
	if err != nil {
		return err
	}
	c.cfg = cfg
	return nil
}

// loadAppConfigConfig は、AWS AppConfigから設定をロードします。
func loadAppConfigConfig(log logging.Logger) (map[string]string, error) {
	// Hosted ConfigurationのProfileからの設定読み込み
	cfg, err := loadHostedAppConfig(log)
	if err != nil {
		return nil, err
	}
	log.Debug("AppConfig設定(Hosted):%v\n", cfg)
	// SecretManagerのProfileからの設定読み込み
	smCfg, err := loadSecretManagerConfig(log)
	if err != nil {
		return nil, err
	}
	log.Debug("AppConfig設定(SM):%v\n", smCfg)
	// 設定をマージ
	maps.Copy(cfg, smCfg)
	log.Debug("AppConfig設定(マージ):%v\n", cfg)
	return cfg, nil
}

// loadHostedAppConfig は、AWS AppConfigからHosted Configurationの設定をロードします。
func loadHostedAppConfig(log logging.Logger) (map[string]string, error) {
	// Hosted Configurationのエンドポイントを環境変数から取得
	hostedCfgUrl, ok := os.LookupEnv(APPCONFIG_HOSTED_EXTENSION_URL_NAME)
	if !ok {
		// 環境変数が設定されていない場合は、空で返す
		return nil, nil
	}
	// AppConfig Lambda ExtensionsのエンドポイントへアクセスしてHosted Configurationの設定データを取得
	response, err := http.Get(hostedCfgUrl)
	if err != nil {
		return nil, errors.Errorf("Hosted ConfigurtionのAppConfig読み込みエラー:%w", err)
	}
	var cfg map[string]string
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Errorf("Hosted ConfigurtionのAppConfig読み込みエラー:%w", err)
	}
	log.Debug("Hosted ConfigurationのAppConfig読み込み(yaml)):%s\n", string(data))
	// YAMLの設定データを読み込み
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, errors.Errorf("Hosted ConfigurtionのAppConfig読み込みエラー:%w", err)
	}
	return cfg, nil
}

// loadSecretManagerConfig は、AWS AppConfigからSecretManagerの設定をロードします。
func loadSecretManagerConfig(log logging.Logger) (map[string]string, error) {
	// SecretManagerのエンドポイントを環境変数から取得
	smCfgUrl, ok := os.LookupEnv(APPCONFIG_SM_EXTENSION_URL_NAME)
	if !ok {
		// 環境変数が設定されていない場合は、空で返す
		return nil, nil
	}
	// AppConfig Lambda ExtensionsのエンドポイントへアクセスしてSecretManagerの設定データを取得
	response, err := http.Get(smCfgUrl)
	if err != nil {
		return nil, errors.Errorf("SecretManagerのAppConfig読み込みエラー:%w", err)
	}
	var cfg map[string]string
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Errorf("SecretManagerのAppConfig読み込みエラー:%w", err)
	}
	log.Debug("SecretManagerのAppConfig読み込み(json):%s\n", string(data))
	// JSONの設定データを読み込み
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, errors.Errorf("SecretManagerのAppConfig読み込みエラー:%w", err)
	}
	return cfg, nil
}
