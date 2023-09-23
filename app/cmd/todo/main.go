package main

import (
	"app/internal/app/todo/service"
	"app/internal/pkg/entity"
	"app/internal/pkg/repository"
	"context"
	"encoding/json"
	"net/http"

	"example.com/appbase/pkg/apcontext"
	"example.com/appbase/pkg/api"
	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/logging"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/pkg/errors"
)

var (
	// Service
	todoService service.TodoService
	// Logger
	log logging.Logger
	// Config
	cfg *config.Config
)

// リクエストデータ
type Request struct {
	TodoTitle string `json:"todo_title"`
}

// コードルドスタート時の初期化処理
func init() {
	log = logging.NewLogger()
	cfg, err := config.LoadConfig()
	if err != nil {
		//TODO: エラーハンドリング
		log.Fatal("初期化処理エラー:%s", err.Error())
		panic(err.Error())
	}
	todoRepository, err := repository.NewTodoRepository()
	if err != nil {
		//TODO: エラーハンドリング
		log.Fatal("初期化処理エラー:%s", err.Error())
		panic(err.Error())
	}
	todoService = service.NewTodoService(log, cfg, &todoRepository)
}

// ハンドラメソッド
func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	//ctxの格納
	apcontext.Context = ctx
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
	todoId, err := parseGetRequest(request)
	if err != nil {
		return api.ErrorResponse(err)
	}
	//サービスの実行
	result, err := todoService.Find(todoId)
	if err != nil {
		log.Error("service execution error: %s", err)
		return api.ErrorResponse(err)
	}
	//レスポンスデータの返却
	resultString, err := formatResponse(result)
	if err != nil {
		return api.ErrorResponse(err)
	}
	return api.OkResponse(resultString)
}

// Postリクエストの処理
func postHandler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	//リクエストデータの解析
	p, err := parsePostRequest(request)
	if err != nil {
		return api.ErrorResponse(err)
	}
	//サービスの実行
	result, err := todoService.Regist(p.TodoTitle)
	if err != nil {
		log.Error("service execution error: %s", err)
		return api.ErrorResponse(err)
	}
	//レスポンスデータの返却
	resultString, err := formatResponse(result)
	if err != nil {
		return api.ErrorResponse(err)
	}
	return api.OkResponse(resultString)
}

// Getリクエストデータの解析
func parseGetRequest(req events.APIGatewayProxyRequest) (string, error) {
	if req.HTTPMethod != http.MethodGet {
		return "", errors.Errorf("use GET request")
	}
	todoId := req.PathParameters["todo_id"]
	return todoId, nil
}

// Postリクエストデータの解析
func parsePostRequest(req events.APIGatewayProxyRequest) (*Request, error) {
	var r Request
	err := api.ParsePostRequest(req, &r)
	return &r, err
}

// レスポンスデータの生成
func formatResponse(todo *entity.Todo) (string, error) {
	resp, err := json.Marshal(todo)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse request")
	}
	return string(resp), nil
}

// Main関数
func main() {
	lambda.Start(handler)
}
