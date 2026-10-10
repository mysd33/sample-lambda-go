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
	GenerateUUID() (string, error)
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
func (*defaultIDGenerator) GenerateUUID() (string, error) {
	// Go1.27から標準機能でuuidパッケージが利用可能
	uuidObj := uuid.New()
	return uuidObj.String(), nil
}

// GenerateShortID implements IDGenerator.
func (*defaultIDGenerator) GenerateShortID() string {
	return shortid.MustGenerate()
}
