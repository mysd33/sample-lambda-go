package main

import (
	"app/internal/pkg/entity"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"example.com/appbase/pkg/apcontext"
	"example.com/appbase/pkg/component"
	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestPostTodo(t *testing.T) {
	//  テスト用にX-Rayのセグメント開始
	ctx, seg := xray.BeginSegment(context.Background(), "main_test")
	apcontext.Context = ctx
	defer seg.Close(nil)

	// TODO: 暫定対処　テスト用のApplicationContextの作成
	ac := component.NewApplicationContextForTest()
	// 業務の初期化処理実行
	r := gin.Default()
	initBiz(ac, r)

	t.Run("Postのテスト", func(t *testing.T) {
		w := httptest.NewRecorder()
		input := "{ \"todo_title\" : \"Buy Milk\"}"
		req, _ := http.NewRequest("POST", "/todo-api/v1/todo", strings.NewReader(input))
		// サーバ処理の実行
		r.ServeHTTP(w, req)
		// ステータスコード200で返却されること
		assert.Equal(t, 200, w.Code)
		var actual entity.Todo
		json.Unmarshal(w.Body.Bytes(), &actual)
		assert.NotEqual(t, "", actual.ID)
		assert.Equal(t, "Buy Milk", actual.Title)
	})
}
