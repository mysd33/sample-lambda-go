package main

import (
	"app/internal/pkg/model"
	"app/internal/pkg/repository"
	"context"
	"fmt"
	"testing"

	"example.com/appbase/pkg/apcontext"
	"example.com/appbase/pkg/component"
	"example.com/appbase/pkg/env"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-xray-sdk-go/xray"
)

func TestRegisterAllAsync(t *testing.T) {
	// テスト実行用に動作環境名を環境変数に設定
	env.SetTestEnv(t)

	//  テスト用にX-Rayのセグメント開始
	ctx, seg := xray.BeginSegment(context.Background(), "main_test")
	apcontext.Context = ctx
	defer seg.Close(nil)
	// ApplicationContextの作成
	ac := component.NewApplicationContext()
	// 業務の初期化処理実行
	asyncControllerFunc := initBiz(ac)

	// TODO: データ駆動テスト化
	t.Run("RegisterAllAsyncのテスト", func(t *testing.T) {
		// TODO: テーブル作成
		// TODO: tempテーブルのテストデータ登録（仮置きのコード）
		value := "todoFiles/fef300f7-b37d-11ee-9c20-0242ac110006.json"
		tempRepository := repository.NewTempRepository(ac.GetDynamoDBTemplate(), ac.GetDynamoDBAccessor(), ac.GetLogger(), ac.GetConfig(), ac.GetIDGenerator())
		testData, err := tempRepository.CreateOne(&model.Temp{Value: value})
		tempId := testData.ID
		if err != nil {
			t.Errorf("テストデータ登録エラー: %v", err)
		}
		// TODO: S3バケット作成
		// TODO: jsonファイル作成

		// 入力メッセージの作成
		sqsMessage := events.SQSMessage{
			Body: fmt.Sprintf("{\"tempId\":\"%s\"}", tempId),
		}
		// テスト対処処理実行
		asyncControllerFunc(sqsMessage)

		//TODO: DBの状態確認

		//TODO: DBのテーブル・テストデータ削除
	})

}
