/*
httpclient パッケージは、REST APIの呼び出し等のためのHTTPクライアントの機能を提供するパッケージです。
*/
package httpclient

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/cockroachdb/errors"
	"golang.org/x/net/context/ctxhttp"

	"example.com/appbase/pkg/apcontext"
	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/env"
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
	// デフォルトリトライ対象ステータスコード
	HTTP_CLIENT_RETRY_STATUS_CODE_NAME     = "HTTP_CLIENT_RETRY_STATUS_CODE"
	HTTP_CLIENT_DEFAULT_RETRY_STATUS_CODES = "408,429,500,502,503,504"
)

// HTTPClient は、HTTPクライアントのインタフェースです。
type HTTPClient interface {
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

// defaultHTTPClientは、HTTPClientインタフェースを実装する構造体です。
type defaultHTTPClient struct {
	config  config.Config
	logger  logging.Logger
	retryer retry.Retryer[*http.Response]
}

// NewHTTPClient は、HTTPClientを生成します。
func NewHTTPClient(config config.Config, logger logging.Logger) HTTPClient {
	retryer := retry.NewRetryer[*http.Response](logger)
	return &defaultHTTPClient{
		config:  config,
		logger:  logger,
		retryer: retryer,
	}
}

// Get implements HTTPClient.
func (c *defaultHTTPClient) Get(url string, header http.Header, params map[string]string) (*ResponseData, error) {
	// リトライ対応のGet処理の実行
	return c.GetWithContext(apcontext.Context, url, header, params)
}

// GetWithContext implements HTTPClient.
func (c *defaultHTTPClient) GetWithContext(ctx context.Context, url string, header http.Header, params map[string]string) (*ResponseData, error) {
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

// Post implements HTTPClient.
func (c *defaultHTTPClient) Post(url string, header http.Header, bbody []byte) (*ResponseData, error) {
	// リトライ対応のPost処理の実行
	return c.PostWithContext(apcontext.Context, url, header, bbody)
}

// PostWithContext implements HTTPClient.
func (c *defaultHTTPClient) PostWithContext(ctx context.Context, url string, header http.Header, bbody []byte) (*ResponseData, error) {
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
func (c *defaultHTTPClient) doGet(ctx context.Context, url string, header http.Header, params map[string]string) retry.RetryableFunc[*http.Response] {
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
func (c *defaultHTTPClient) doPost(ctx context.Context, url string, header http.Header, bbody []byte) retry.RetryableFunc[*http.Response] {
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
func (c *defaultHTTPClient) checkRetryable() retry.CheckRetryable[*http.Response] {
	statusCodes := c.getRetryableStatusCodes()
	return func(result *http.Response, err error) bool {
		// エラーの場合は、リトライを行う
		if err != nil {
			return true
		}
		// リトライ可能なステータスコードの場合は、リトライを行う
		_, ok := statusCodes[result.StatusCode]
		return ok
	}
}

// getRetryableStatusCodes は、プロパティよりリトライ対象のステータスコードを取得します。
func (c *defaultHTTPClient) getRetryableStatusCodes() map[int]struct{} {
	statusCodesStr := strings.Split(
		c.config.Get(HTTP_CLIENT_RETRY_STATUS_CODE_NAME, HTTP_CLIENT_DEFAULT_RETRY_STATUS_CODES),
		",",
	)
	statusCodes := make(map[int]struct{}, len(statusCodesStr))
	for _, codeStr := range statusCodesStr {
		code, err := strconv.Atoi(codeStr)
		if err != nil {
			c.logger.Warn(message.W_FW_8012, codeStr)
			if !env.IsProd() {
				// 開発中は、設定誤りを検知するため、異常終了
				panic(err)
			}
			// 本番環境では、設定誤りをスキップして続行
			continue
		}
		statusCodes[code] = struct{}{}
	}
	return statusCodes
}

// retryOptions は、リトライオプションを取得する関数です。
func (c *defaultHTTPClient) retryOptions() []retry.Option {
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
