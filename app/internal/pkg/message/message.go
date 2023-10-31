package message

import (
	_ "embed"
)

//go:embed messages.yaml
var Messages_yaml []byte
