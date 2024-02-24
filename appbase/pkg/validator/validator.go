// validator は、入力チェックに関する機能を提供するパッケージです。
package validator

import (
	"reflect"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/ja"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	ja_translations "github.com/go-playground/validator/v10/translations/ja"
)

const (
	LABEL_TAG = "label"
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

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// 日本語化の設定
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

		// エラーメッセージの項目名をlabelタグがあれば表示するよう設定
		v.RegisterTagNameFunc(func(field reflect.StructField) string {
			name := field.Tag.Get(LABEL_TAG)
			if name == "" {
				name = field.Name
			}
			return name
		})
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
