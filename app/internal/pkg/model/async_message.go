// entityのパッケージ
package model

// AsyncMessage は、非同期メッセージを表す構造体です。
type AsyncMessage struct {
	// TempId は、TempテーブルのID情報です。
	TempId string `json:"tempId"`
}
