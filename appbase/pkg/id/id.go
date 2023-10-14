/*
id パッケージはID生成に関する機能を提供するパッケージです。
*/
package id

import (
	"github.com/google/uuid"
	"github.com/teris-io/shortid"
)

// GenerateId は、UUIDでIDを採番します。
func GenerateId() string {
	uuidObj, _ := uuid.NewUUID()
	return uuidObj.String()
}

// GenerateShortId は、ShortIDでIDを採番します。
func GenerateShortId() string {
	return shortid.MustGenerate()
}
