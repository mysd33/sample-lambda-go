/*
message パッケージはメッセージを管理する機能を提供します。
*/
package message

import (
	_ "embed"
	"fmt"
	"maps"

	"gopkg.in/yaml.v3"
)

var (
	//go:embed messages_fw.yaml
	fwMessagesBytes []byte
)

// MessageSource は、メッセージを取得するインタフェースです。
type MessageSource interface {
	// GetMessage は、メッセージID（id）に対応し、置換文字列argsを設定したするメッセージを取得します。
	GetMessage(id string, args ...any) string
	// Add は業務APのメッセージのyaml定義(appMessagesBytes)を、MessageSourceに追加します。
	Add(appMessagesBytes []byte) error
}

// defaultMessageSource は MessageSourceを実装する構造体です。
type defaultMessageSource struct {
	fwMessages  map[string]string
	appMessages map[string]string
}

// NewMessageSource は、MessageSourceを作成します。
func NewMessageSource() (MessageSource, error) {
	//フレームワークのメッセージ定義（messages_fw.yaml）作成する
	var fwMessages map[string]string
	err := yaml.Unmarshal(fwMessagesBytes, &fwMessages)
	if err != nil {
		return nil, err
	}
	return &defaultMessageSource{fwMessages: fwMessages, appMessages: map[string]string{}}, nil
}

// Add implements MessageSource.
func (ms *defaultMessageSource) Add(appMessagesBytes []byte) error {
	var appNewMessages map[string]string
	if err := yaml.Unmarshal(appMessagesBytes, &appNewMessages); err != nil {
		return err
	}
	maps.Copy(ms.appMessages, appNewMessages)
	return nil
}

// GetMessage implements MessageSource.
func (ms *defaultMessageSource) GetMessage(id string, args ...any) string {
	// メッセージIDに対するメッセージ（のテンプレート）を取得
	template := ""
	if val, ok := ms.appMessages[id]; ok {
		template = val
	} else if val, ok := ms.fwMessages[id]; ok {
		template = val
	}
	// idに対応するメッセージが取得できない場合は空列で返す
	if template == "" {
		return ""
	}
	// 置き換え文字列がない場合
	if len(args) == 0 {
		return template
	}
	return fmt.Sprintf(template, args...)
}
