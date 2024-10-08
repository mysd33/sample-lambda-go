package main

import (
	"app/cmd/common"
	"app/internal/pkg/model"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"example.com/appbase/pkg/apcontext"
	"example.com/appbase/pkg/component"
	"example.com/appbase/pkg/env"
	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/stretchr/testify/assert"
)

func TestPostTodo(t *testing.T) {
	// テスト実行用に動作環境名を環境変数に設定
	env.SetTestEnv(t)

	//  テスト用にX-Rayのセグメント開始
	ctx, seg := xray.BeginSegment(context.Background(), "main_test")
	apcontext.Context = ctx
	defer seg.Close(nil)

	ac := component.NewApplicationContext()
	apiLambdaHandler := ac.GetAPILambdaHandler()
	r := apiLambdaHandler.GetDefaultGinEngine(common.NewCommonErrorResponse(ac.GetMessageSource()), nil)
	initBiz(ac, r)

	// TODO: データ駆動テスト化
	t.Run("Postのテスト", func(t *testing.T) {
		w := httptest.NewRecorder()
		input := "{ \"todo_title\" : \"Buy Milk\"}"
		req, _ := http.NewRequest("POST", "/todo-api/v1/todo", strings.NewReader(input))
		// サーバ処理の実行
		r.ServeHTTP(w, req)
		// ステータスコード200で返却されること
		assert.Equal(t, 200, w.Code)
		var actual model.Todo
		json.Unmarshal(w.Body.Bytes(), &actual)
		assert.NotEqual(t, "", actual.ID)
		assert.Equal(t, "Buy Milk", actual.Title)

		//TODO: DBの状態確認
	})
}
