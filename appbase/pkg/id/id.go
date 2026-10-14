/*
id パッケージはID生成に関する機能を提供するパッケージです。
*/
package id

import (
	"uuid"

	"github.com/teris-io/shortid"
)

// IDGenerator は、ID生成機能を提供するインターフェースです。
type IDGenerator interface {
	// GenerateUUID は、UUIDでIDを採番します。
	GenerateUUID() string
	// GenerateShortID は、ShortIDでIDを採番します。
	GenerateShortID() string
}

// defaultIDGenerator は、IDGeneratorのデフォルト実装です。
type defaultIDGenerator struct{}

// NewIDGenerator は、IDGeneratorを作成します。
func NewIDGenerator() IDGenerator {
	return &defaultIDGenerator{}
}

// GenerateUUID implements IDGenerator.
func (*defaultIDGenerator) GenerateUUID() string {
	// Go1.27から標準機能でuuidパッケージが利用可能
	// 標準ではUUIDのバージョン7で生成する実装とする
	uuidObj := uuid.NewV7()
	return uuidObj.String()
}

// GenerateShortID implements IDGenerator.
func (*defaultIDGenerator) GenerateShortID() string {
	return shortid.MustGenerate()
}
