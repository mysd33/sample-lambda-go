package httpclient

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"example.com/appbase/pkg/code"
	"example.com/appbase/pkg/errors"
	"example.com/appbase/pkg/logging"
)

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

type httpClientImpl struct {
	log logging.Logger
	// TODO:
}

func NewHttpClient(log logging.Logger) HttpClient {
	return &httpClientImpl{log: log}
}

// Get implements HttpClient.
func (c *httpClientImpl) Get(url string, header http.Header, params map[string]string) (*ResponseData, error) {
	// TODO: headerの設定
	// TODO: リトライの実装
	// TODO: X-Ray対応

	// Getメソッドの実行
	response, err := http.Get(url)
	// TODO: エラーコード
	if err != nil {
		return nil, errors.NewSystemError(err, code.E_FW_9002)
	}
	// TODO: 200以外のレスポンスエラー時の対応

	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	// TODO: エラーコード
	if err != nil {
		return nil, errors.NewSystemError(err, code.E_FW_9002)
	}
	return &ResponseData{
		StatusCode:     response.StatusCode,
		Status:         response.Status,
		ContentType:    getResponseContentType(response.Header),
		Body:           data,
		ResponseHeader: response.Header}, nil

}

// Post implements HttpClient.
func (*httpClientImpl) Post(url string, header http.Header, bbody []byte) (*ResponseData, error) {
	// TODO: headerの設定
	// TODO: リトライの実装
	// TODO: X-Ray対応

	// Postメソッドの実行
	// TODO: ContentType固定でよいか？
	response, err := http.Post(url, "application/json", bytes.NewReader(bbody))
	// TODO: エラーコード
	if err != nil {
		return nil, errors.NewSystemError(err, code.E_FW_9002)
	}
	// TODO: 200以外のレスポンスエラー時の対応

	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	// TODO: エラーコード
	if err != nil {
		return nil, errors.NewSystemError(err, code.E_FW_9002)
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
