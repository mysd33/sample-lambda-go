// awssdk パッケージは、AWS SDKを利用する際のユーティリティを提供します。
package awssdk

import (
	"strings"

	"example.com/appbase/pkg/config"
	"github.com/aws/aws-sdk-go-v2/aws"
)

const (
	AWSSDK_CLIENT_LOG_MODE_NAME = "AWSSDK_CLIENT_LOG_MODE"
	AWSSDK_CLIENT_LOG_MODE_SEP  = ","
)

// GetClientLogMode は、設定からAWS SDKのClientLogModeを取得します。
func GetClientLogMode(config config.Config) (aws.ClientLogMode, bool) {
	var logModes aws.ClientLogMode

	logModesCfg, found := config.GetWithContains(AWSSDK_CLIENT_LOG_MODE_NAME)
	if !found {
		return logModes, false
	}
	// カンマで区切られた文字列を配列に変換
	logModesStr := strings.Split(logModesCfg, AWSSDK_CLIENT_LOG_MODE_SEP)

	for _, logModeStr := range logModesStr {
		// 設定されたログモードをAWS SDKのClientLogModeに変換してOR演算で結合
		logMode := convertToClientLogMode(logModeStr)
		logModes = logModes | logMode
	}
	return logModes, true
}

func convertToClientLogMode(logMode string) aws.ClientLogMode {
	switch logMode {
	case "LogSigning":
		return aws.LogSigning
	case "LogRetries":
		return aws.LogRetries
	case "LogRequest":
		return aws.LogRequest
	case "LogRequestWithBody":
		return aws.LogRequestWithBody
	case "LogResponse":
		return aws.LogResponse
	case "LogResponseWithBody":
		return aws.LogResponseWithBody
	case "LogDeprecatedUsage":
		return aws.LogDeprecatedUsage
	case "LogRequestEventMessage":
		return aws.LogRequestEventMessage
	case "LogResponseEventMessage":
		return aws.LogResponseEventMessage
	default:
		return aws.ClientLogMode(0)
	}
}
