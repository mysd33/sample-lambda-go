package id

import (
	"github.com/google/uuid"
	"github.com/teris-io/shortid"
)

func GenerateId() string {
	uuidObj, _ := uuid.NewUUID()
	return uuidObj.String()
}

func GenerateShortId() string {
	return shortid.MustGenerate()
}
