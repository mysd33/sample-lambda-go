/*
config パッケージは、設定ファイルを管理するパッケージです。
*/
package config

type Config interface {
	Get(key string) string
}

func NewConfig() (Config, error) {
	//TODO: Env=Localの場合はviperでconfigファイルからとるようにする
	//それ以外の場合は、Envに合わせたAppConfigから値をとるように作る
	return newViperConfig()
}
