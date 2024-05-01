/*
id パッケージはID生成に関する機能を提供するパッケージです。
*/
package id

import (
	"github.com/google/uuid"
	"github.com/pkg/errors"
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
	uuidObj, err := uuid.NewUUID()
	if err != nil {
		errors.WithStack(err)
	}
	return uuidObj.String(), nil
}

// GenerateShortID implements IDGenerator.
func (*defaultIDGenerator) GenerateShortID() string {
	return shortid.MustGenerate()
}
