/*
domain パッケージは、ドメイン層の機能を提供するパッケージです。
*/
package domain

// ServiceFunc は、Serviceで実行する関数です。
type ServiceFunc func() (interface{}, error)
