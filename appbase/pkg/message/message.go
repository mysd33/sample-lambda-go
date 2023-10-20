// message パッケージはメッセージを管理する機能を提供します。
package message

import "fmt"

type any interface{}

// MessageSource は、メッセージを取得するインタフェースです。
type MessageSource interface {
	// GetMessage は、codeに対応し、置換文字列argsを設定したするメッセージを取得します。
	GetMessage(code string, args ...any) string
}

// DefaultMessageSource は MessageSourceを実装する構造体です。
type DefaultMessageSource struct {
	//TODO:
}

func NewMessageSource() MessageSource {
	return &DefaultMessageSource{}
}

// GetMessage implements MessageSource.
func (*DefaultMessageSource) GetMessage(code string, args ...any) string {
	//TODO: configs/messages.yamlの定義メッセージIDを取得する実装
	//TODO: configs/messages_fw.yamlの定義メッセージIDを取得する実装
	template := ""
	if len(args) == 0 {
		return template
	}
	return fmt.Sprintf(template, args)
}
