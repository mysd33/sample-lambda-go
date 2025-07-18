/*
input パッケージは、検索条件の入力データを扱うパッケージです。
*/
package input

import "example.com/appbase/pkg/dynamodb/gsi"

// PKOnlyQueryInput は、プライマリキーの完全一致の条件指定による検索時のインプット構造体
type PKOnlyQueryInput struct {
	// プライマリキー
	PrimaryKey PrimaryKey
	// 取得項目
	SelectAttributes []string
	// 強い整合性読み込みの使用有無
	ConsitentRead bool
}

// PKQueryInput は、パーティションキーの完全一致とソートキーの条件指定による複数検索用のインプット構造体
type PKQueryInput struct {
	// プライマリキーの条件
	PrimaryKey PrimaryKey
	// 取得項目
	SelectAttributes []string
	// フィルタ条件（プライマリキーの条件以外で絞込を行いたい場合）
	WhereClauses []*WhereClause
	// 強い整合性読み込みの使用有無
	ConsitentRead bool
}

// GsiQueryInput は、GSIによる検索条件指定による複数検索用のインプット構造体
type GsiQueryInput struct {
	// GSI名
	GSIName gsi.DynamoDBGSIName
	// インデックスキーの条件
	IndexKey PrimaryKey
	// 取得項目
	SelectAttributes []string
	// フィルタ条件（プライマリキーの条件以外で絞込を行いたい場合）
	WhereClauses []*WhereClause
	// 取得件数の上限値
	TotalLimit *int32
	// 1回のクエリでの取得件数上限値
	LimitPerQuery *int32
}

// UpdateInput は、更新時のインプット構造体
type UpdateInput struct {
	// プライマリキーの条件
	PrimaryKey PrimaryKey
	// 条件付き更新の条件
	WhereClauses []*WhereClause
	// 更新項目
	UpdateAttributes []*Attribute
	// 削除項目名（更新時に、属性の値事態を削除したい）
	RemoveAttributeNames []string
}

// DeleteInput は、削除時のインプット構造体
type DeleteInput struct {
	// プライマリキーの条件
	PrimaryKey PrimaryKey
	// 条件付き削除の条件
	WhereClauses []*WhereClause
}

// Attribute は、属性の名称と値のペア構造体です。
type Attribute struct {
	Name  string
	Value any
}

// PrimaryKey は、プライマリキー（パーティションキーとソートキー）の条件句です。
type PrimaryKey struct {
	// パーティションキーの指定
	PartitionKey Attribute
	// ソートキーの条件の値指定
	SortKey *Attribute
	// ソートキーの検索条件演算子
	SortKeyOp SortKeyOperator
	// ソートキーのソート条件句
	SortkeyOrderBy OrderBy
}

// SortKeyOperator は、ソートキーの検索条件句です。
type SortKeyOperator string

const (
	SORTKEY_EQUAL           = SortKeyOperator("Equal")
	SORTKEY_BEGINS_WITH     = SortKeyOperator("BeginWith")
	SORTKEY_BETWEEN         = SortKeyOperator("Between")
	SORTKEY_GREATER_THAN    = SortKeyOperator("GreaterThan")
	SORTKEY_GREATER_THAN_EQ = SortKeyOperator("GreaterThanEqual")
	SORTKEY_LESS_THAN       = SortKeyOperator("LessThan")
	SORTKEY_LESS_THAN_EQ    = SortKeyOperator("LessThanEqual")
)

// ソートキーのソート順指定
type OrderBy string

const (
	ORDER_BY_DESC = OrderBy("Desc")
	ORDER_BY_ASC  = OrderBy("Asc")
)

// WhereClause は、検索時のフィルタ条件または、更新・登録時の条件を表します。
type WhereClause struct {
	// Where句で指定する属性
	Attribute Attribute
	// Where句の演算子
	WhereOp WhereOperator
	// Where句を連結する演算子
	AppendOp AppendOperator
}

// WhereOperator は、フィルタの条件指定する際の演算子です。
type WhereOperator string

const (
	WHERE_EQUAL           = WhereOperator("Equal")
	WHERE_NOT_EQUAL       = WhereOperator("NotEqual")
	WHERE_BEGINS_WITH     = WhereOperator("BeginWith")
	WHERE_GREATER_THAN    = WhereOperator("GreaterThan")
	WHERE_GREATER_THAN_EQ = WhereOperator("GreaterThanEqual")
	WHERE_LESS_THAN       = WhereOperator("LessThan")
	WHERE_LESS_THAN_EQ    = WhereOperator("LessThanEqual")
)

// AppendOperator は、フィルタの条件を連結する際の演算子です。
type AppendOperator string

const (
	APPEND_AND = AppendOperator("And")
	APPEND_OR  = AppendOperator("Or")
)
