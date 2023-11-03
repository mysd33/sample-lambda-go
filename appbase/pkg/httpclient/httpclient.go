/*
httpclient パッケージは、REST APIの呼び出し等のためのHTTPクライアントの機能を提供するパッケージです。
*/
package httpclient

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/cockroachdb/errors"
	"golang.org/x/net/context/ctxhttp"

	"example.com/appbase/pkg/apcontext"
	"example.com/appbase/pkg/logging"
)

// HttpClient は、HTTPクライアントのインタフェースです。
type HttpClient interface {
	Get(url string, header http.Header, params map[string]string) (*ResponseData, error)
	Post(url string, header http.Header, bbody []byte) (*ResponseData, error)
	// TODO
}

// ResponseData は、httpのレスポンスを表す構造体です。
type ResponseData struct {
	// ステータスコード
	StatusCode int
	// ステータス文字列
	Status string
	// コンテントタイプ
	ContentType string
	// レスポンスボディ
	Body []byte
	// レスポンスヘッダー
	ResponseHeader http.Header
}

// defaultHttpClientは、HttpClientインタフェースを実装する構造体です。
type defaultHttpClient struct {
	log logging.Logger
	// TODO:
}

func NewHttpClient(log logging.Logger) HttpClient {
	return &defaultHttpClient{log: log}
}

// Get implements HttpClient.
func (c *defaultHttpClient) Get(url string, header http.Header, params map[string]string) (*ResponseData, error) {
	// TODO: headerの設定
	// TODO: リトライの実装

	// Getメソッドの実行（X-Ray対応）
	response, err := ctxhttp.Get(apcontext.Context, xray.Client(nil), url)

	if err != nil {
		return nil, errors.WithStack(err)
	}
	// TODO: 200以外のレスポンスエラー時の対応

	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)

	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &ResponseData{
		StatusCode:     response.StatusCode,
		Status:         response.Status,
		ContentType:    getResponseContentType(response.Header),
		Body:           data,
		ResponseHeader: response.Header}, nil

}

// Post implements HttpClient.
func (c *defaultHttpClient) Post(url string, header http.Header, bbody []byte) (*ResponseData, error) {
	// TODO: headerの設定
	// TODO: リトライの実装

	// Postメソッドの実行（X-Ray対応）
	// TODO: ContentType固定でよいか？
	response, err := ctxhttp.Post(apcontext.Context, xray.Client(nil), url, "application/json", bytes.NewReader(bbody))

	if err != nil {
		return nil, errors.WithStack(err)
	}
	// TODO: 200以外のレスポンスエラー時の対応

	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &ResponseData{
		StatusCode:     response.StatusCode,
		Status:         response.Status,
		ContentType:    getResponseContentType(response.Header),
		Body:           data,
		ResponseHeader: response.Header}, nil

}

func getResponseContentType(header http.Header) string {
	contentType := ""
	for key, val := range header {
		if strings.ToLower(key) == "content-type" {
			contentType = val[0]
		}
	}
	return contentType
}
