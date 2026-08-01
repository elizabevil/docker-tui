package component

import (
	"reflect"
	"testing"

	"github.com/elizabevil/docker-tui/internal/tui/tables"
)

func TestTableLayoutTracksColumnContentWidths(t *testing.T) {
	data := TableData{
		Cols: []tables.ColumnDef{
			{Key: "name", Header: "Name", Basis: 8, Min: 6, Shrink: 1, Fill: 1},
			{Key: "image", Header: "Image", Basis: 8, Min: 6, Shrink: 1, Fill: 1},
		},
		Rows: [][]string{{"nginx", "alpine"}}, BannerW: 30,
	}
	headers := resolveHeaders(data)
	first := resolveTableLayout(data, headers, 20, 2)
	data.Rows[0][0] = "a-much-longer-name"
	second := resolveTableLayout(data, headers, 20, 2)
	if second.widths[0] <= second.widths[1] || first.widths[0] != first.widths[1] {
		t.Fatalf("content demand was not prioritized: first=%v second=%v", first.widths, second.widths)
	}
}

func TestTableLayoutKeepsCompactTracksAndFillsWithGutters(t *testing.T) {
	data := TableData{
		Cols: []tables.ColumnDef{
			{Key: "id", Header: "ID", Fixed: 12},
			{Key: "name", Header: "Name", Basis: 16, Min: 10, Max: 24, Shrink: 1, Fill: 1},
			{Key: "stats", Header: "Stats", Basis: 30, Min: 18, Max: 36, Shrink: 1, Fill: 1},
		},
		Rows: [][]string{{"container", "dtui-test-redis", "0% 0% RX:4.3KB TX:1.1KB"}},
	}
	layout := resolveTableLayout(data, resolveHeaders(data), 80, 2)
	if !reflect.DeepEqual(layout.widths, []int{12, 16, 30}) || layout.trailingWidth != 18 {
		t.Fatalf("content tracks are wrong: widths=%v trailing=%d", layout.widths, layout.trailingWidth)
	}
	if got := renderedLayoutWidth(layout); got != 80 {
		t.Fatalf("rendered layout width=%d, want 80: %+v", got, layout)
	}
}

func TestTableLayoutDoesNotStretchWideSparseRows(t *testing.T) {
	data := TableData{
		Cols: []tables.ColumnDef{
			{Key: "id", Header: "ID", Fixed: 12},
			{Key: "name", Header: "NAME", Basis: 30, Min: 12, Shrink: 1, Fill: 2},
			{Key: "subnet", Header: "SUBNET", Basis: 30, Min: 14, Shrink: 1, Fill: 2},
			{Key: "created", Header: "CREATED", Fixed: 19},
		},
		Rows: [][]string{{"e664062f6e46", "integration_default", "10.89.0.0/24", "2026-07-30 05:22:32"}},
	}
	layout := resolveTableLayout(data, resolveHeaders(data), 180, 2)
	if layout.widths[1] != 30 || layout.widths[2] != 30 {
		t.Fatalf("wide sparse columns were stretched: %v", layout.widths)
	}
	if renderedLayoutWidth(layout) != 180 {
		t.Fatalf("adaptive gutters did not fill viewport: %+v", layout)
	}
}

func TestAdaptiveGapWidthsConsumeTrailingSpaceEvenly(t *testing.T) {
	got := adaptiveGapWidths(2, 4, 10)
	want := []int{5, 5, 4, 4}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("adaptive gaps = %v, want %v", got, want)
	}
}

func renderedLayoutWidth(layout tableLayout) int {
	total := 0
	for _, width := range layout.widths {
		total += width
	}
	for _, width := range layout.gapWidths {
		total += width
	}
	return total
}

func BenchmarkResolveTableLayout(b *testing.B) {
	data := benchmarkTableData()
	headers := resolveHeaders(data)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = resolveTableLayout(data, headers, data.BannerW, 2)
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
