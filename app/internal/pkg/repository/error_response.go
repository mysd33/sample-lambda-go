// repositoryのパッケージ
package repository

// ErrorResponse は、エラーレスポンスを表す構造体です。
type ErrorResponse struct {
	Code   string `json:"code"`
	Detail any    `json:"detail"`
}
