/*
domain パッケージは、ドメイン層の機能を提供するパッケージです。
*/
package domain

import "context"

// ServiceFunc は、Serviceで実行する関数です。
type ServiceFunc func() (any, error)

// ServiceFuncWithContext は、Context指定ありでServiceで実行する関数です。
type ServiceFuncWithContext func(ctx context.Context) (any, error)
