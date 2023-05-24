package main

import (
	"context"
	"encoding/json"
	"net/http"

	"example.com/appbase/pkg/apcontext"
	"example.com/appbase/pkg/api"
	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/logging"
	"example.com/appbase/pkg/rdb"

	"app/internal/app/user/service"
	"app/internal/pkg/entity"
	"app/internal/pkg/repository"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/pkg/errors"
)

var (
	// Service
	userService service.UserService
	// Logger
	log logging.Logger
	// Config
	cfg *config.Config
)

// リクエストデータ
type Request struct {
	Name string `json:"user_name"`
}

// コードルドスタート時の初期化処理
func init() {
	var err error
	log = logging.NewLogger()
	cfg, err = config.LoadConfig()
	if err != nil {
		//TODO: エラーハンドリング
		log.Fatal("初期化処理エラー:%s", err.Error())
		panic(err.Error())
	}
	// リポジトリの作成
	// DynamoDBの場合
	// userRepository := repository.NewUserRepositoryForDynamoDB()
	// RDBの場合
	userRepository := repository.NewUserRepositoryForRDB()

	// サービスの作成
	userService = service.NewUserService(log, cfg, &userRepository)
}

// ハンドラメソッド
func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// RDBコネクションの確立
	db, err := rdb.RDSConnect()
	if err != nil {
		return api.ErrorResponse(err)
	}
	// 終了時にRDBコネクションの切断
	defer db.Close()

	//ctxの格納
	apcontext.Context = ctx
	apcontext.DB = db

	//Getリクエストの処理
	if request.HTTPMethod == http.MethodGet {
		return getHandler(ctx, request)
	}
	//Postリクエストの処理
	return postHandler(ctx, request)
}

// Getリクエストの処理
func getHandler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	//リクエストデータの解析
	userId, err := parseGetRequest(request)
	if err != nil {
		log.Error("parse request error: %s", err)
		return api.ErrorResponse(err)
	}
	// トランザクション開始
	tx, err := apcontext.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Error("transaction begin error: %s", err)
		return api.ErrorResponse(err)
	}
	apcontext.Tx = tx

	// サービスの実行
	result, err := userService.Find(userId)
	if err != nil {
		// トランザクションロールバック
		err2 := tx.Rollback()
		if err2 != nil {
			log.Error("transaction rollback error: %s", err2)
			return api.ErrorResponse(err2)
		}
		log.Error("service execution error: %s", err)
		return api.ErrorResponse(err)
	}
	// レスポンスデータの返却
	resultString, err := formatResponse(result)
	if err != nil {
		// トランザクションロールバック
		err2 := tx.Rollback()
		if err2 != nil {
			log.Error("transaction rollback error: %s", err2)
			return api.ErrorResponse(err2)
		}
		log.Error("format response error: %s", err)
		return api.ErrorResponse(err)
	}
	// トランザクションコミット
	err = tx.Commit()
	if err != nil {
		log.Error("transaction commit error: %s", err)
		return api.ErrorResponse(err)
	}

	return api.OkResponse(resultString)
}

// Postリクエストの処理
func postHandler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	//リクエストデータの解析
	p, err := parsePostRequest(request)
	if err != nil {
		log.Error("parse request error: %s", err)
		return api.ErrorResponse(err)
	}
	// トランザクション開始
	tx, err := apcontext.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Error("transaction begin error: %s", err)
		return api.ErrorResponse(err)
	}
	apcontext.Tx = tx

	//サービスの実行
	result, err := userService.Regist(p.Name)
	if err != nil {
		// トランザクションロールバック
		err2 := tx.Rollback()
		if err2 != nil {
			log.Error("transaction rollback error: %s", err2)
			return api.ErrorResponse(err2)
		}
		log.Error("service execution error: %s", err)
		return api.ErrorResponse(err)
	}

	//レスポンスデータの返却
	resultString, err := formatResponse(result)
	if err != nil {
		// トランザクションロールバック
		err2 := tx.Rollback()
		if err2 != nil {
			log.Error("transaction rollback error: %s", err2)
			return api.ErrorResponse(err2)
		}
		log.Error("format response error: %s", err)
		return api.ErrorResponse(err)
	}
	// トランザクションコミット
	err = tx.Commit()
	if err != nil {
		log.Error("transaction commit error: %s", err)
		return api.ErrorResponse(err)
	}

	return api.OkResponse(resultString)
}

// Getリクエストデータの解析
func parseGetRequest(req events.APIGatewayProxyRequest) (string, error) {
	if req.HTTPMethod != http.MethodGet {
		return "", errors.Errorf("use GET request")
	}
	userId := req.PathParameters["user_id"]
	return userId, nil
}

// Postリクエストデータの解析
func parsePostRequest(req events.APIGatewayProxyRequest) (*Request, error) {
	var r Request
	err := api.ParsePostRequest(req, &r)
	return &r, err
}

// レスポンスデータの生成
func formatResponse(user *entity.User) (string, error) {
	resp, err := json.Marshal(user)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse request")
	}
	return string(resp), nil
}

// Main関数
func main() {
	lambda.Start(handler)
}
