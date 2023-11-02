package logging

import (
	_ "embed"
	"fmt"
	"testing"

	"example.com/appbase/pkg/errors"
	"example.com/appbase/pkg/message"
)

//go:embed logging_testmsg.yaml
var logging_test_yaml []byte

// テスト対象の構造体
func sut() *zapLogger {
	messageSource, _ := message.NewMessageSource()
	// テスト用のメッセージ定義を読み込み
	messageSource.Add(logging_test_yaml)
	logger, _ := NewLogger(messageSource)
	return logger.(*zapLogger)
}

func Test_zapLogger_Debug(t *testing.T) {

	type args struct {
		template string
		args     []interface{}
	}
	tests := []struct {
		name string
		z    *zapLogger
		args args
	}{
		// テストケース
		{name: "置換文字列なしのテスト",
			z:    sut(),
			args: args{template: "デバッグログ"}},
		{name: "置換文字列ありのテスト",
			z:    sut(),
			args: args{template: "デバッグログ:%s,%s", args: []interface{}{"hoge", "fuga"}}},
		{name: "置き換え文字列誤りのテスト",
			z:    sut(),
			args: args{template: "デバッグログ", args: []interface{}{"hoge", "fuga"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.z.Debug(tt.args.template, tt.args.args...)
		})
	}
}

func Test_zapLogger_Info(t *testing.T) {
	type args struct {
		code string
		args []interface{}
	}
	tests := []struct {
		name string
		z    *zapLogger
		args args
	}{
		// テストケース
		{name: "メッセージID取得できた場合",
			z:    sut(),
			args: args{code: "logtest001", args: []interface{}{"aaaa"}},
		},
		{name: "メッセージID取得できた場合(置換文字列が多い)",
			z:    sut(),
			args: args{code: "logtest001", args: []interface{}{"aaaa", "bbbb"}},
		},
		{name: "メッセージID取得できない場合",
			z:    sut(),
			args: args{code: "xxxxx", args: []interface{}{"aaaa", "bbbb"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.z.Info(tt.args.code, tt.args.args...)
		})
	}
}

func Test_zapLogger_Warn(t *testing.T) {
	type args struct {
		code string
		args []interface{}
	}
	tests := []struct {
		name string
		z    *zapLogger
		args args
	}{
		// テストケース
		{name: "メッセージID取得できた場合",
			z:    sut(),
			args: args{code: "logtest001", args: []interface{}{"aaaa"}},
		},
		{name: "メッセージID取得できた場合(置換文字列が多い)",
			z:    sut(),
			args: args{code: "logtest001", args: []interface{}{"aaaa", "bbbb"}},
		},
		{name: "メッセージID取得できない場合",
			z:    sut(),
			args: args{code: "xxxxx", args: []interface{}{"aaaa", "bbbb"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.z.Warn(tt.args.code, tt.args.args...)
		})
	}
}

func Test_zapLogger_WarnWithCodableError(t *testing.T) {
	type args struct {
		err errors.CodableError
	}
	tests := []struct {
		name string
		z    *zapLogger
		args args
	}{
		// テストケース
		{name: "業務エラーを受け取りメッセージID取得できた場合",
			z:    sut(),
			args: args{err: errors.NewBusinessError("logtest001", "aaaa")},
		},
		{name: "業務エラーを受け取りメッセージID取得できない場合",
			z:    sut(),
			args: args{err: errors.NewBusinessError("xxxxx", "aaaa")},
		},
		{name: "ラップされた業務エラーを受け取りメッセージID取得できた場合",
			z:    sut(),
			args: args{err: errors.NewBusinessErrorWithCause(fmt.Errorf("原因のエラー"), "logtest001", "aaaa")},
		},
		{name: "ラップされた業務エラーを受け取りメッセージID取得できない場合",
			z:    sut(),
			args: args{err: errors.NewBusinessErrorWithCause(fmt.Errorf("原因のエラー"), "xxxx", "aaaa")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.z.WarnWithCodableError(tt.args.err)
		})
	}
}

func Test_zapLogger_Error(t *testing.T) {
	type args struct {
		code string
		args []interface{}
	}
	tests := []struct {
		name string
		z    *zapLogger
		args args
	}{
		// テストケース
		{name: "メッセージID取得できた場合",
			z:    sut(),
			args: args{code: "logtest001", args: []interface{}{"aaaa"}},
		},
		{name: "メッセージID取得できた場合(置換文字列が多い)",
			z:    sut(),
			args: args{code: "logtest001", args: []interface{}{"aaaa", "bbbb"}},
		},
		{name: "メッセージID取得できない場合",
			z:    sut(),
			args: args{code: "xxxxx", args: []interface{}{"aaaa", "bbbb"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.z.Error(tt.args.code, tt.args.args...)
		})
	}
}

func Test_zapLogger_ErrorWithCodableError(t *testing.T) {
	type args struct {
		err errors.CodableError
	}
	tests := []struct {
		name string
		z    *zapLogger
		args args
	}{
		// テストケース
		{name: "ラップされたシステムエラーを受け取りメッセージID取得できた場合",
			z:    sut(),
			args: args{err: errors.NewSystemError(fmt.Errorf("原因のエラー"), "logtest001", "aaaa")},
		},

		{name: "ラップされたシステムエラーを受け取りメッセージID取得できない場合",
			z:    sut(),
			args: args{err: errors.NewSystemError(fmt.Errorf("原因のエラー"), "xxxxx", "aaaa")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.z.ErrorWithCodableError(tt.args.err)
		})
	}
}

func Test_zapLogger_Fatal(t *testing.T) {
	type args struct {
		code string
		args []interface{}
	}
	tests := []struct {
		name string
		z    *zapLogger
		args args
	}{
		// テストケース
		{name: "メッセージID取得できた場合",
			z:    sut(),
			args: args{code: "logtest001", args: []interface{}{"aaaa"}},
		},
		{name: "メッセージID取得できた場合(置換文字列が多い)",
			z:    sut(),
			args: args{code: "logtest001", args: []interface{}{"aaaa", "bbbb"}},
		},
		{name: "メッセージID取得できない場合",
			z:    sut(),
			args: args{code: "xxxxx", args: []interface{}{"aaaa", "bbbb"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.z.Fatal(tt.args.code, tt.args.args...)
		})
	}
}

func Test_zapLogger_FatalWithCodableError(t *testing.T) {
	type args struct {
		err errors.CodableError
	}
	tests := []struct {
		name string
		z    *zapLogger
		args args
	}{
		// テストケース
		{name: "ラップされたシステムエラーを受け取りメッセージID取得できた場合",
			z:    sut(),
			args: args{err: errors.NewSystemError(fmt.Errorf("原因のエラー"), "logtest001", "aaaa")},
		},

		{name: "ラップされたシステムエラーを受け取りメッセージID取得できない場合",
			z:    sut(),
			args: args{err: errors.NewSystemError(fmt.Errorf("原因のエラー"), "xxxxx", "aaaa")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.z.FatalWithCodableError(tt.args.err)
		})
	}
}
