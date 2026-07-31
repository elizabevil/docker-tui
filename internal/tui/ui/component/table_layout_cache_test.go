package component

import (
	"testing"

	"github.com/elizabevil/docker-tui/internal/tui/tables"
)

func TestTableLayoutCacheTracksColumnContentWidths(t *testing.T) {
	cache := NewTableLayoutCache(8)
	data := TableData{
		Cols: []tables.ColumnDef{
			{Key: "name", Header: "Name", Basis: 8, Min: 6, Shrink: 1},
			{Key: "image", Header: "Image", Basis: 8, Min: 6, Shrink: 1},
		},
		Rows: [][]string{{"nginx", "alpine"}}, BannerW: 30,
	}
	headers := resolveHeaders(data)
	first := cache.resolve(data, headers, 20, 2)
	second := cache.resolve(data, headers, 20, 2)
	if cache.hits != 1 || cache.misses != 1 {
		t.Fatalf("cache stats hits=%d misses=%d", cache.hits, cache.misses)
	}
	if first.widths[0] != second.widths[0] {
		t.Fatalf("cached width changed")
	}

	data.Rows[0][0] = "a-much-longer-name"
	third := cache.resolve(data, resolveHeaders(data), 20, 2)
	if cache.hits != 1 || cache.misses != 2 {
		t.Fatalf("content widths did not invalidate layout cache: hits=%d misses=%d", cache.hits, cache.misses)
	}
	if third.widths[0] <= third.widths[1] || first.widths[0] != first.widths[1] {
		t.Fatalf("content demand was not prioritized: first=%v third=%v", first.widths, third.widths)
	}
}

func TestTableLayoutKeepsFixedGapAndFillsViewport(t *testing.T) {
	cache := NewTableLayoutCache(8)
	data := TableData{
		Cols: []tables.ColumnDef{
			{Key: "id", Header: "ID", Fixed: 12},
			{Key: "name", Header: "Name", Basis: 16, Min: 10, Max: 24, Shrink: 1},
			{Key: "stats", Header: "Stats", Basis: 30, Min: 18, Max: 36, Shrink: 1},
		},
		Rows: [][]string{{"container", "dtui-test-redis", "0% 0% RX:4.3KB TX:1.1KB"}},
	}
	layout := cache.resolve(data, resolveHeaders(data), 80, 2)
	total := layout.gap * (len(layout.widths) - 1)
	for _, width := range layout.widths {
		total += width
	}
	if total != layout.contentWidth || layout.contentWidth+layout.trailingWidth != 80 || layout.gap != 2 {
		t.Fatalf("layout width=%d gap=%d widths=%v", total, layout.gap, layout.widths)
	}
	if layout.widths[1] != 24 || layout.widths[2] != 36 || layout.trailingWidth != 4 {
		t.Fatalf("max widths or trailing space are wrong: widths=%v trailing=%d", layout.widths, layout.trailingWidth)
	}
}

func TestTableLayoutCacheIncludesColumnGap(t *testing.T) {
	cache := NewTableLayoutCache(8)
	data := TableData{Cols: []tables.ColumnDef{{Key: "name", Header: "Name", Basis: 10, Min: 6, Shrink: 1}}}
	headers := resolveHeaders(data)
	cache.resolve(data, headers, 40, 1)
	cache.resolve(data, headers, 40, 2)
	if cache.hits != 0 || cache.misses != 2 {
		t.Fatalf("gap did not invalidate layout cache: hits=%d misses=%d", cache.hits, cache.misses)
	}
}

func TestTableLayoutCacheIsBounded(t *testing.T) {
	cache := NewTableLayoutCache(2)
	for _, value := range []string{"one", "two", "three"} {
		data := TableData{
			Cols: []tables.ColumnDef{{Key: "name", Header: "Name"}},
			Rows: [][]string{{value}}, BannerW: 20,
		}
		cache.resolve(data, resolveHeaders(data), 20, 2)
	}
	if len(cache.entries) > 2 {
		t.Fatalf("cache grew past limit: %d", len(cache.entries))
	}
}

func BenchmarkTableLayoutCache(b *testing.B) {
	data := benchmarkTableData()
	headers := resolveHeaders(data)
	cache := NewTableLayoutCache(8)
	cache.resolve(data, headers, data.BannerW, 2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.resolve(data, headers, data.BannerW, 2)
	}
}

func BenchmarkTableLayoutUncached(b *testing.B) {
	data := benchmarkTableData()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		widths := tables.ResolveColumnWidths(data.Cols, data.BannerW, 2)
		_ = widths
	}
}

func benchmarkTableData() TableData {
	rows := make([][]string, 50)
	for i := range rows {
		rows[i] = []string{"\x1b[32mnginx-production\x1b[0m", "running", "127.0.0.1:8080->80/tcp"}
	}
	return TableData{
		Cols: []tables.ColumnDef{
			{Key: "name", Header: "Name", Basis: 24, Min: 12, Shrink: 1},
			{Key: "state", Header: "State", Basis: 12, Min: 8, Shrink: 1},
			{Key: "ports", Header: "Ports", Basis: 30, Min: 12, Shrink: 1},
		},
		Rows: rows, BannerW: 102,
	}
}
