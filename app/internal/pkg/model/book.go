// modelのパッケージ
package model

// Book は、書籍の情報を表す構造体です。
type Book struct {
	// Title は、書籍のタイトルです。
	Title string `json:"title"`
	// Author は、書籍の著者です。
	Author string `json:"author"`
	// Publisher は、書籍の出版社です。
	Publisher string `json:"publisher"`
	// PublishedDate は、書籍の発売日です。
	PublishedDate string `json:"published_date"`
	// ISBNは、書籍のISBNです。
	ISBN string `json:"isbn"`
}
