package component

import (
	"testing"

	"github.com/elizabevil/docker-tui/internal/tui/tables"
)

func TestTableLayoutCacheHitAndInvalidation(t *testing.T) {
	cache := NewTableLayoutCache(8)
	data := TableData{
		Cols:   []tables.ColumnDef{{Key: "name", Header: "Name", Flex: 1}},
		Widths: []int{20}, Rows: [][]string{{"nginx"}}, BannerW: 20,
	}
	headers := resolveHeaders(data)
	first := cache.resolve(data, headers, 20)
	second := cache.resolve(data, headers, 20)
	if cache.hits != 1 || cache.misses != 1 {
		t.Fatalf("cache stats hits=%d misses=%d", cache.hits, cache.misses)
	}
	if first.widths[0] != second.widths[0] {
		t.Fatalf("cached width changed")
	}

	data.Rows[0][0] = "a-much-longer-image-name"
	third := cache.resolve(data, resolveHeaders(data), 20)
	if cache.misses != 2 {
		t.Fatalf("row change did not invalidate cache")
	}
	if third.widths[0] == first.widths[0] {
		t.Fatalf("invalidated width was not recomputed")
	}
}

func TestTableLayoutCacheIsBounded(t *testing.T) {
	cache := NewTableLayoutCache(2)
	for _, value := range []string{"one", "two", "three"} {
		data := TableData{
			Cols:   []tables.ColumnDef{{Key: "name", Header: "Name"}},
			Widths: []int{20}, Rows: [][]string{{value}}, BannerW: 20,
		}
		cache.resolve(data, resolveHeaders(data), 20)
	}
	if len(cache.entries) > 2 {
		t.Fatalf("cache grew past limit: %d", len(cache.entries))
	}
}

func BenchmarkTableLayoutCache(b *testing.B) {
	data := benchmarkTableData()
	headers := resolveHeaders(data)
	cache := NewTableLayoutCache(8)
	cache.resolve(data, headers, data.BannerW)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.resolve(data, headers, data.BannerW)
	}
}

func BenchmarkTableLayoutUncached(b *testing.B) {
	data := benchmarkTableData()
	headers := resolveHeaders(data)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		widths := computeContentWidths(data, headers)
		_ = computeGap(widths, data.BannerW)
	}
}

func benchmarkTableData() TableData {
	rows := make([][]string, 50)
	for i := range rows {
		rows[i] = []string{"\x1b[32mnginx-production\x1b[0m", "running", "127.0.0.1:8080->80/tcp"}
	}
	return TableData{
		Cols: []tables.ColumnDef{
			{Key: "name", Header: "Name"}, {Key: "state", Header: "State"}, {Key: "ports", Header: "Ports"},
		},
		Widths: []int{30, 20, 50}, Rows: rows, BannerW: 102,
	}
}
