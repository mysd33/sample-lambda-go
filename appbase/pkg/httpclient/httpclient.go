package httpclient

import (
	"io"
	"net/http"
	"strings"

	"example.com/appbase/pkg/code"
	"example.com/appbase/pkg/errors"
	"example.com/appbase/pkg/logging"
)

type HttpClient interface {
	Get(url string, header http.Header, params map[string]string) (*ResponseData, error)
	//TODO: Post
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

func (c *httpClientImpl) Get(url string, header http.Header, params map[string]string) (*ResponseData, error) {
	// TODO: リトライの実装
	// TODO: X-Ray対応
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
	contentType := ""
	for key, val := range response.Header {
		if strings.ToLower(key) == "content-type" {
			contentType = val[0]
		}
	}

	return &ResponseData{
		StatusCode:     response.StatusCode,
		Status:         response.Status,
		ContentType:    contentType,
		Body:           data,
		ResponseHeader: response.Header}, nil

}
