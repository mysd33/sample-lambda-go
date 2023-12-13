/*
handler パッケージは、Lambdaのハンドラメソッドに関する機能を提供するパッケージです。
*/
package handler

import (
	"context"
)

// SimpleLambdaHandlerFunc は、その他のトリガのLambdaのハンドラメソッドを表す関数です。
type SimpleLambdaHandlerFunc func(ctx context.Context) error

// TODO:他のトリガのLambdaHandlerFuncの実装
func ScheduledBatchLambdaHandler() SimpleLambdaHandlerFunc {
	return func(ctx context.Context) error {
		//　TODO: 実装
		panic("not implement")
	}
}

func FlowBatchLambdaHandler() SimpleLambdaHandlerFunc {
	return func(ctx context.Context) error {
		//　TODO: 実装
		panic("not implement")
	}
}
