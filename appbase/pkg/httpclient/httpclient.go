/*
httpclient パッケージは、REST APIの呼び出し等のためのHTTPクライアントの機能を提供するパッケージです。
*/
package httpclient

import (
	"bytes"
	"context"
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
	"example.com/appbase/pkg/message"
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
	// GetWithContext は、goroutine向けに、渡されたContextを利用して、GETメソッドでリクエストを送信します。
	GetWithContext(ctx context.Context, url string, header http.Header, params map[string]string) (*ResponseData, error)
	// Post は、POSTメソッドでリクエストを送信します。
	Post(url string, header http.Header, bbody []byte) (*ResponseData, error)
	// PostWithContext は、goroutine向けに、渡されたContextを利用して、POSTメソッドでリクエストを送信します。
	PostWithContext(ctx context.Context, url string, header http.Header, bbody []byte) (*ResponseData, error)
	// TODO: 必要に応じて追加
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
	logger  logging.Logger
	retryer retry.Retryer[*http.Response]
}

// NewHttpClient は、HttpClientを生成します。
func NewHttpClient(config config.Config, logger logging.Logger) HttpClient {
	retryer := retry.NewRetryer[*http.Response](logger)
	return &defaultHttpClient{
		config:  config,
		logger:  logger,
		retryer: retryer,
	}
}

// Get implements HttpClient.
func (c *defaultHttpClient) Get(url string, header http.Header, params map[string]string) (*ResponseData, error) {
	// リトライ対応のGet処理の実行
	response, err := c.retryer.Do(
		c.doGet(apcontext.Context, url, header, params),
		c.checkRetryable(),
		c.retryOptions()...,
	)
	if err != nil {
		return nil, err
	}
	return createResponseData(response)
}

// GetWithContext implements HttpClient.
func (c *defaultHttpClient) GetWithContext(ctx context.Context, url string, header http.Header, params map[string]string) (*ResponseData, error) {
	// リトライ対応のGet処理の実行
	response, err := c.retryer.DoWithContext(ctx,
		c.doGet(ctx, url, header, params),
		c.checkRetryable(),
		c.retryOptions()...,
	)
	if err != nil {
		return nil, err
	}
	return createResponseData(response)
}

// Post implements HttpClient.
func (c *defaultHttpClient) Post(url string, header http.Header, bbody []byte) (*ResponseData, error) {
	// リトライ対応のPost処理の実行
	response, err := c.retryer.Do(
		c.doPost(apcontext.Context, url, header, bbody),
		c.checkRetryable(),
		c.retryOptions()...,
	)
	if err != nil {
		return nil, err
	}
	return createResponseData(response)
}

// PostWithContext implements HttpClient.
func (c *defaultHttpClient) PostWithContext(ctx context.Context, url string, header http.Header, bbody []byte) (*ResponseData, error) {
	// リトライ対応のPost処理の実行
	response, err := c.retryer.DoWithContext(ctx,
		c.doPost(ctx, url, header, bbody),
		c.checkRetryable(),
		c.retryOptions()...,
	)
	if err != nil {
		return nil, err
	}
	return createResponseData(response)
}

// doGet は、GETメソッドを実行します。
func (c *defaultHttpClient) doGet(ctx context.Context, url string, header http.Header, params map[string]string) retry.RetryableFunc[*http.Response] {
	return func() (*http.Response, error) {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		// URLにクエリパラメータの追加
		if params != nil {
			qparam := req.URL.Query()
			for key, val := range params {
				qparam.Add(key, val)
			}
			req.URL.RawQuery = qparam.Encode()
		}
		// ヘッダー情報の設定
		if header != nil {
			req.Header = header
		}
		c.logger.Info(message.I_FW_0005, "GET", url)
		// Getメソッドの実行（X-Ray対応）
		response, err := ctxhttp.Do(ctx, xray.Client(nil), req)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return response, nil
	}
}

// doPost は、POSTメソッドを実行します。
func (c *defaultHttpClient) doPost(ctx context.Context, url string, header http.Header, bbody []byte) retry.RetryableFunc[*http.Response] {
	return func() (*http.Response, error) {
		req, err := http.NewRequest("POST", url, bytes.NewReader(bbody))
		if err != nil {
			return nil, err
		}
		// ヘッダー情報の設定
		if header != nil {
			req.Header = header
		}
		req.Header.Set("Content-Type", "application/json")
		c.logger.Info(message.I_FW_0005, "POST", url)
		// POSTメソッドの実行（X-Ray対応）
		response, err := ctxhttp.Do(ctx, xray.Client(nil), req)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return response, nil
	}
}

// checkRetryable は、リトライ可能かどうかを判定する関数です。
func (c *defaultHttpClient) checkRetryable() retry.CheckRetryable[*http.Response] {
	return func(result *http.Response, err error) bool {
		// エラーの場合は、リトライを行う
		if err != nil {
			return true
		}
		// リトライ可能なステータスコードの場合は、リトライを行う
		return c.isRetryable(result.StatusCode)
	}
}

// isRetryable は、リトライ可能なステータスコードかどうかを判定する関数です。
func (c *defaultHttpClient) isRetryable(statusCode int) bool {
	switch statusCode {
	// TODO: プロパティから設定可能にする
	case 408, 429, 500, 502, 503, 504:
		return true
	}
	return false
}

// retryOptions は、リトライオプションを取得する関数です。
func (c *defaultHttpClient) retryOptions() []retry.Option {
	return []retry.Option{
		retry.MaxRetryTimes(uint(c.config.GetInt(HTTP_CLIENT_MAX_RETRY_TIMES_NAME, HTTP_CLIENT_DEFAULT_MAX_RETRY_TIMES))),
		retry.Interval(time.Duration(c.config.GetInt(HTTP_CLIENT_RETRY_INTERVAL_NAME, HTTP_CLIENT_DEFAULT_RETRY_INTERVAL)) * time.Millisecond),
	}
}

// createResponseData は、レスポンスデータを生成する関数です。
func createResponseData(response *http.Response) (*ResponseData, error) {
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
