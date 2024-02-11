/*
httpclient パッケージは、REST APIの呼び出し等のためのHTTPクライアントの機能を提供するパッケージです。
*/
package httpclient

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/cockroachdb/errors"
	"golang.org/x/net/context/ctxhttp"

	"example.com/appbase/pkg/apcontext"
	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/logging"
	"example.com/appbase/pkg/retry"
)

const (
	// デフォルト最大リトライ回数3回
	HTTP_CLIENT_MAX_RETRY_TIMES_NAME    = "HTTP_CLIENT_MAX_RETRY_TIMES"
	HTTP_CLIENT_DEFAULT_MAX_RETRY_TIMES = 3
	// デフォルトリトライ間隔500ms
	HTTP_CLIENT_RETRY_INTERVAL_NAME    = "HTTP_CLIENT_RETRY_INTERVAL"
	HTTP_CLIENT_DEFAULT_RETRY_INTERVAL = 500
)

// HttpClient は、HTTPクライアントのインタフェースです。
type HttpClient interface {
	// Get は、GETメソッドでリクエストを送信します。
	Get(url string, header http.Header, params map[string]string) (*ResponseData, error)
	// Post は、POSTメソッドでリクエストを送信します。
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
	config  config.Config
	log     logging.Logger
	retryer retry.Retryer[*http.Response]
}

// NewHttpClient は、HttpClientを生成します。
func NewHttpClient(config config.Config, log logging.Logger) HttpClient {
	retryer := retry.NewRetryer[*http.Response](log)
	return &defaultHttpClient{
		config:  config,
		log:     log,
		retryer: retryer,
	}
}

// Get implements HttpClient.
func (c *defaultHttpClient) Get(url string, header http.Header, params map[string]string) (*ResponseData, error) {
	// リトライ処理の実行
	response, err := c.retryer.Do(func() (*http.Response, error) {
		// TODO: headerの設定

		// Getメソッドの実行（X-Ray対応）
		response, err := ctxhttp.Get(apcontext.Context, xray.Client(nil), url)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return response, nil
	}, func(result *http.Response, err error) bool {
		// レスポンスがエラーの場合は、リトライを行う
		if err != nil {
			return true
		}
		// リトライ可能なステータスコードの場合は、リトライを行う
		return c.isRetryable(result.StatusCode)
	},
		// リトライオプション設定
		retry.MaxRetryTimes(uint(c.config.GetInt(HTTP_CLIENT_MAX_RETRY_TIMES_NAME, HTTP_CLIENT_DEFAULT_MAX_RETRY_TIMES))),
		retry.Interval(time.Duration(c.config.GetInt(HTTP_CLIENT_RETRY_INTERVAL_NAME, HTTP_CLIENT_DEFAULT_RETRY_INTERVAL))*time.Millisecond),
	)
	if err != nil {
		return nil, err
	}
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
	// リトライ処理の実行
	response, err := c.retryer.Do(func() (*http.Response, error) {
		// TODO: headerの設定

		// Postメソッドの実行（X-Ray対応）
		response, err := ctxhttp.Post(apcontext.Context, xray.Client(nil), url, "application/json", bytes.NewReader(bbody))
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return response, nil
	}, func(result *http.Response, err error) bool {
		// レスポンスがエラーの場合は、リトライを行う
		if err != nil {
			return true
		}
		// リトライ可能なステータスコードの場合は、リトライを行う
		return c.isRetryable(result.StatusCode)
	},
		// リトライオプション設定
		retry.MaxRetryTimes(uint(c.config.GetInt(HTTP_CLIENT_MAX_RETRY_TIMES_NAME, HTTP_CLIENT_DEFAULT_MAX_RETRY_TIMES))),
		retry.Interval(time.Duration(c.config.GetInt(HTTP_CLIENT_RETRY_INTERVAL_NAME, HTTP_CLIENT_DEFAULT_RETRY_INTERVAL))*time.Millisecond),
	)
	if err != nil {
		return nil, err
	}
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

// isRetryable は、リトライ可能なステータスコードかどうかを判定する関数です。
func (c *defaultHttpClient) isRetryable(statusCode int) bool {
	switch statusCode {
	// TODO: 外から設定可能にする
	case 408, 429, 500, 502, 503, 504:
		return true
	}
	// TODO: それ以外のステータスコード時の対応
	return false
}

// getResponseContentType は、レスポンスのContent-Typeを取得する関数です。
func getResponseContentType(header http.Header) string {
	contentType := ""
	for key, val := range header {
		if strings.ToLower(key) == "content-type" {
			contentType = val[0]
		}
	}
	return contentType
}
