package state

// TableFilter 统一表格过滤接口，供搜索框逻辑复用。
type TableFilter interface {
	SetFilter(query string)
	FilterText() string
}
