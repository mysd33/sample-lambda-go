// message パッケージはメッセージを管理する機能を提供します。
package message

import "fmt"

// MessageSource は、メッセージを取得するインタフェースです。
type MessageSource interface {
	// GetMessage は、メッセージID（id）に対応し、置換文字列argsを設定したするメッセージを取得します。
	GetMessage(id string, args ...interface{}) string
}

// defaultMessageSource は MessageSourceを実装する構造体です。
type defaultMessageSource struct {
	//TODO:
}

func NewMessageSource() MessageSource {
	return &defaultMessageSource{}
}

// GetMessage implements MessageSource.
func (*defaultMessageSource) GetMessage(id string, args ...interface{}) string {
	//TODO: configs/messages.yamlの定義メッセージIDを取得する実装
	//TODO: configs/messages_fw.yamlの定義メッセージIDを取得する実装
	template := ""
	// idに対応するメッセージが取得できない場合はそのまま出力
	if template == "" {
		return fmt.Sprint(id, args)
	}
	// 置き換え文字列がない場合
	if len(args) == 0 {
		return template
	}
	return fmt.Sprintf(template, args...)
}
