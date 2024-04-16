/*
env パッケージは、動作環境の情報を扱うパッケージです。
*/
package env

import (
	"os"
	"testing"
)

const (
	// 動作環境を表す環境変数名
	ENV_NAME = "ENV"
	// sam localによるローカル実行での実行環境名
	ENV_LOCAL = "Local"
	// go testでのテスト実行での実行環境名
	ENV_LOCAL_TEST = "LocalTest"
	// クラウド上の開発・テスト環境で動作する実行環境名
	ENV_DEVELOPMENT = "Dev"
	// クラウド上のステージング環境で動作する実行環境名
	ENV_STAGING = "Staging"
	// クラウド上の商用環境で動作する実行環境名
	ENV_PRODUCTION = "Prod"
)

// GetEnv は、環境変数から動作環境名を取得します。
func GetEnv() string {
	return os.Getenv(ENV_NAME)
}

// IsLocalOrLocalTestは、動作環境がローカル実行環境(Local、LocalTest)かを判定します。
func IsLocalOrLocalTest() bool {
	env := GetEnv()
	return env == ENV_LOCAL || env == ENV_LOCAL_TEST
}

// IsLocalは、動作環境がsam localによるローカル実行環境（Local）かを判定します。
func IsLocal() bool {
	env := GetEnv()
	return env == ENV_LOCAL
}

// IsLocalTestは、動作環境が、go testでのテスト実行環境（LocalTest）かを判定します。
func IsLocalTest() bool {
	env := GetEnv()
	return env == ENV_LOCAL_TEST
}

// IsDevは、動作環境が、クラウド上の開発・テスト環境（Dev）かを判定します。
func IsDev() bool {
	env := GetEnv()
	return env == ENV_DEVELOPMENT
}

// IsStragingOrProdは、 動作環境が本番相当の環境（Staging、Prod）かを判定します。
func IsStragingOrProd() bool {
	env := GetEnv()
	return env == ENV_STAGING || env == ENV_PRODUCTION
}

// IsStagingは、動作環境が、ステージング環境で動作する実行環境（Staging）かを判定します。
func IsStaging() bool {
	env := GetEnv()
	return env == ENV_STAGING
}

// IsProdは、動作環境が、ステージング環境で動作する実行環境（Prod）かを判定します。
func IsProd() bool {
	env := GetEnv()
	return env == ENV_PRODUCTION
}

// SetTestEnvにテスト動作実行時の動作環境名をを設定します。
func SetTestEnv(t *testing.T) {
	t.Setenv(ENV_NAME, ENV_LOCAL_TEST)
}

// SetTestEnvForBechMarkは、ベンチマークテスト実行時の動作環境名をを設定します。
func SetTestEnvForBechMark(t *testing.B) {
	t.Setenv(ENV_NAME, ENV_LOCAL_TEST)
}
