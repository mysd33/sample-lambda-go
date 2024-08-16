/*
config パッケージは、設定ファイルを管理するパッケージです。
*/
package config

type testConfig struct {
	cfg map[string]string
}

// NewTestConfig は、テスト用Configを作成します。
func NewTestConfig(cfg map[string]string) Config {
	return &testConfig{cfg: cfg}
}

// GetWithContains implements Config.
func (c *testConfig) GetWithContains(key string) (string, bool) {
	v, found := c.cfg[key]
	return v, found
}

// Get implements Config.
func (c *testConfig) Get(key string, defaultValue string) string {
	value, found := c.GetWithContains(key)
	return returnStringValueIfFound(found, value, defaultValue)
}

// GetIntWithContains implements Config.
func (c *testConfig) GetIntWithContains(key string) (int, bool) {
	value, found := c.GetWithContains(key)
	result, err := convertIntValueIfFound(found, value)
	if err != nil {
		// int変換に失敗した場合は、値が見つからなかったとしてfalseを返す
		return 0, false
	}
	return result, found
}

// GetInt implements Config.
func (c *testConfig) GetInt(key string, defaultValue int) int {
	value, found := c.GetIntWithContains(key)
	return returnIntValueIfFound(found, value, defaultValue)
}

// GetBoolWithContains implements Config.
func (c *testConfig) GetBoolWithContains(key string) (bool, bool) {
	value, found := c.GetWithContains(key)
	result, err := convertBoolValueIfFound(found, value)
	if err != nil {
		// bool変換に失敗した場合は、値が見つからなかったとしてfalseを返す
		return false, false
	}
	return result, found
}

// GetBool implements Config.
func (c *testConfig) GetBool(key string, defaultValue bool) bool {
	value, found := c.GetBoolWithContains(key)
	return returnBoolValueIfFound(found, value, defaultValue)
}

// Reload implements Config.
func (c *testConfig) Reload() error {
	// 何もしない
	return nil
}
