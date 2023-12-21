/*
criteria パッケージは、検索条件の入力データを扱うパッケージです。
*/
package criteria

import "example.com/appbase/pkg/dynamodb/gsi"

// PkOnlyQueryInput は、パーティションキーの完全一致の条件指定による検索時のインプット構造体
type PkOnlyQueryInput struct {
	// プライマリキー
	PrimarKey KeyPair
	// 取得項目
	SelectItems []string
	// 強い整合性読み込みの使用有無
	ConsitentRead bool
}

// PkQueryInput は、パーティションキーの完全一致とソートキーの条件指定時のインプット構造体
type PkQueryInput struct {
	// プライマリキーの条件
	PrimaryKey KeyPair
	// 取得項目
	SelectItems []string
	// フィルタ条件（プライマリキーの条件以外で絞込を行いたい場合）
	WhereKeys []*WhereKeys
	// 強い整合性読み込みの使用有無
	ConsitentRead bool
}

// GsiQueryInput は、GSIによる検索条件時のインプット構造体
type GsiQueryInput struct {
	// GSI名
	GSIName gsi.DynamoDBGSIName
	// インデックスキーの条件
	IndexKey KeyPair
	// 取得項目
	SelectItems []string
	// フィルタ条件（プライマリキーの条件以外で絞込を行いたい場合）
	WhereKeys []*WhereKeys
	// 取得件数の上限値
	TotalLimit *int32
	// 1回のクエリでの取得件数上限値
	LimitPerQuery *int32
}

// UpdateInput は、更新時のインプット構造体
type UpdateInput struct {
	// プライマリキーの条件
	PrimarKey KeyPair
	// フィルタ条件（プライマリキーの条件以外で絞込を行いたい場合）
	WhereKeys []*WhereKeys
	// 更新項目
	UpdateItems []*KeyValue
}

// DeleteInput は、削除時のインプット構造体
type DeleteInput struct {
	// プライマリキーの条件
	PrimarKey KeyPair
	// フィルタ条件（プライマリキーの条件以外で絞込を行いたい場合）
	WhereKeys []*WhereKeys
}

// KeyValue は、項目名と値のペア構造体です。
type KeyValue struct {
	Key   string
	Value any
}

// パーティションキーとソートキーの条件
type KeyPair struct {
	// パーティションキーの指定
	PartitionKey KeyValue
	// ソートキーの条件の値指定
	SortKey *KeyValue
	// ソートキーの検索条件句
	SortKeyCond SortKeyCond
	// ソートキーのソート条件句
	SortkeyOrderBy OrderBy
}

// SortKeyCond は、ソートキーの検索条件句です。
type SortKeyCond string

const (
	SORTKEY_COND_EQUAL           = SortKeyCond("Equal")
	SORTKEY_COND_BEGINS_WITH     = SortKeyCond("BeginWith")
	SORTKEY_COND_BETWEEN         = SortKeyCond("Between")
	SORTKEY_COND_GREATER_THAN    = SortKeyCond("GreaterThan")
	SORTKEY_COND_GREATER_THAN_EQ = SortKeyCond("GreaterThanEqual")
	SORTKEY_COND_LESS_THAN       = SortKeyCond("LessThan")
	SORTKEY_COND_LESS_THAN_EQL   = SortKeyCond("LessThanEqual")
)

// ソートキーのソート順指定
type OrderBy string

const (
	ORDER_BY_DESC = OrderBy("Desc")
	ORDER_BY_ASC  = OrderBy("Asc")
)

// WhereKeys は、GSIによる検索時のフィルタ条件句です。
type WhereKeys struct {
	WhereKey     KeyValue
	Method       WhereMethod
	AppendMethod AppendMethod
}

// WhereMethod は、GSIによる検索時のフィルタ条件句です。
type WhereMethod string

const (
	WHERE_EQUAL           = WhereMethod("Equal")
	WHERE_NOT_EQUAL       = WhereMethod("NotEqual")
	WHERE_BEGINS_WITH     = WhereMethod("BeginWith")
	WHERE_GREATER_THAN    = WhereMethod("GreaterThan")
	WHERE_GREATER_THAN_EQ = WhereMethod("GreaterThanEqual")
	WHERE_LESS_THAN       = WhereMethod("LessThan")
	WHERE_LESS_THAN_EQ    = WhereMethod("LessThanEqual")
)

// AppendMethod は、フィルタの結合条件
type AppendMethod string

const (
	APPEND_AND = AppendMethod("And")
	APPEND_OR  = AppendMethod("Or")
)
