// validator は、入力チェックに関する機能を提供するパッケージです。
package validator

import (
	"log"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/ja"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	ja_translations "github.com/go-playground/validator/v10/translations/ja"
)

// Translator はValidatorの日本語化を行います。
var Translator ut.Translator

// Setupは、Validatorの初期化処理を実施します。
func Setup() {
	// 参考
	// https://github.com/go-playground/validator/blob/master/_examples/translations/main.go

	// 日本語化の設定
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		//TODO: ログ仮置き
		log.Print("Validator日本語化")
		var uni *ut.UniversalTranslator
		ja := ja.New()
		uni = ut.New(ja, ja)
		trans, found := uni.GetTranslator("ja")
		if found {
			//TODO: ログ仮置き
			log.Print("Translatorが見つかりました")
		} else {
			//TODO: エラーハンドリングの検討
			log.Print("Translatorが見つかりません")
		}
		ja_translations.RegisterDefaultTranslations(v, trans)
		Translator = trans
	}
}

// AddCustomValidatorはカスタムバリデータを追加します。
func AddCustomValidator(tag string, customValidatorFunc validator.Func) {
	// https://gin-gonic.com/ja/docs/examples/custom-validators/

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		//TODO: ログ仮置き
		log.Printf("カスタムバリデータ追加:%s", tag)
		v.RegisterValidation(tag, customValidatorFunc)
	}
}
