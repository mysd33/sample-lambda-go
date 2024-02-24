// validator は、入力チェックに関する機能を提供するパッケージです。
package validator

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/ja"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	ja_translations "github.com/go-playground/validator/v10/translations/ja"
)

// Translator は、Validatorの日本語化を行います。
var Translator ut.Translator

// LogDebug は、デバッグログを出力する関数です。
type LogDebug func(template string, args ...any)

// ValidatorM
type ValidationManager interface {
	// AddCustomValidatorはカスタムバリデータを追加します。
	AddCustomValidator(tag string, customValidatorFunc validator.Func)
}

type defaultValidationManager struct {
	logDebug LogDebug
}

func NewValidationManager(logDebug LogDebug) ValidationManager {
	// 参考
	// https://github.com/go-playground/validator/blob/master/_examples/translations/main.go

	// 日本語化の設定
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		logDebug("Validator日本語化")
		var uni *ut.UniversalTranslator
		ja := ja.New()
		uni = ut.New(ja, ja)
		trans, found := uni.GetTranslator("ja")
		if found {
			logDebug("Translatorが見つかりました")
		} else {
			logDebug("Translatorが見つかりません")
		}
		ja_translations.RegisterDefaultTranslations(v, trans)
		Translator = trans
	}
	return &defaultValidationManager{
		logDebug: logDebug,
	}
}

// AddCustomValidator implements ValidationManager.
func (m *defaultValidationManager) AddCustomValidator(tag string, customValidatorFunc validator.Func) {
	// https://gin-gonic.com/ja/docs/examples/custom-validators/

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		m.logDebug("カスタムバリデータ追加:%s", tag)
		v.RegisterValidation(tag, customValidatorFunc)
	}
}
