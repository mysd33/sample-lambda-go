/*
config パッケージは、設定ファイルを管理するパッケージです。
*/
package config

type testConfig struct {
	cfg map[string]string
}

// NewTestConfig は、テスト用Configを作成します。
func NewTestConfig(cfg map[string]string) (Config, error) {
	return &testConfig{cfg: cfg}, nil
}

func (c *testConfig) Get(key string) string {
	v, found := c.getWithContains(key)
	if !found {
		return ""
	}
	return v
}

// getWithContains implements Config.
func (c *testConfig) getWithContains(key string) (string, bool) {
	v, found := c.cfg[key]
	return v, found
}

// Reload implements Config.
func (c *testConfig) Reload() error {
	// 何もしない
	return nil
}
