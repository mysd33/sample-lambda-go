// entityのパッケージ
package entity

// AsyncMessage は、非同期メッセージを表す構造体です。
type AsyncMessage struct {
	// TODO: メッセージをTempテーブルのID情報のみに変更
	// TempId は、TempテーブルのID情報です。
	TempId string `json:"tempId"`

	TodoTitles []string `json:"todoTitles"`
}
