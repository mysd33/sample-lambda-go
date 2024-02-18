package sdkhttpclient

import (
	"net"
	"time"

	myConfig "example.com/appbase/pkg/config"
	"github.com/aws/aws-sdk-go-v2/aws/transport/http"
)

const (
	// KeepAlive(秒)
	HTTP_KEEP_ALIVE_NAME    = "HTTP_KEEP_ALIVE"
	DEFAULT_HTTP_KEEP_ALIVE = 30
	// コネクション作成のタイムアウト（ミリ秒）
	HTTP_DIALER_TIMEOUT_NAME    = "HTTP_DIALER_TIMEOUT"
	DEFAULT_HTTP_DIALER_TIMEOUT = 500
)

// NewHTTPClient は、AWS SDKのHTTPクライアントを作成します。
func NewHTTPClient(myCfg myConfig.Config) *http.BuildableClient {
	// https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/custom-http/

	return http.NewBuildableClient().WithDialerOptions(
		func(d *net.Dialer) {
			ka := myCfg.GetInt(HTTP_KEEP_ALIVE_NAME, DEFAULT_HTTP_KEEP_ALIVE)
			dt := myCfg.GetInt(HTTP_DIALER_TIMEOUT_NAME, DEFAULT_HTTP_DIALER_TIMEOUT)
			d.KeepAlive = time.Duration(ka) * time.Second
			d.Timeout = time.Duration(dt) * time.Millisecond
		})
}
