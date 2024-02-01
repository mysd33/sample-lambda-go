/*
handler パッケージは、Lambdaのハンドラメソッドに関する機能を提供するパッケージです。
*/
package handler

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/gin-gonic/gin"
)

// ControllerFunc は、同期処理のControllerで実行する関数です。
type ControllerFunc func(ctx *gin.Context) (any, error)

// AsyncControllerFunc は、非同期処理のControllerで実行する関数です。
type AsyncControllerFunc func(sqsMessage events.SQSMessage) error

// SimpleControllerFunc は、その他のトリガのControllerで実行する関数です。
type SimpleControllerFunc func(ctx context.Context, event any) (any, error)
