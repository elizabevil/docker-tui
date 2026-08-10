package component

import (
	"strings"
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/config"
	"github.com/elizabevil/docker-tui/internal/tui/tables"
)

func TestTableColumnsDoNotOwnBackgrounds(t *testing.T) {
	previous := tableCfg
	previousStyles := rawStyles
	previousSafe := safeFallbackRef
	defer func() {
		tableCfg = previous
		rawStyles = previousStyles
		safeFallbackRef = previousSafe
	}()

	ApplyThemeStyles(config.DefaultTheme())
	styles := GetColumnStyles([]tables.ColumnDef{{Key: "id"}, {Key: "name"}, {Key: "state"}})
	for i, column := range styles {
		if column.Style.Background != "" {
			t.Fatalf("column %d owns background %q", i, column.Style.Background)
		}
	}
	if GetRowStyle(RowStyleSelected).Background == "" || GetRowStyle(RowStyleMarked).Background == "" {
		t.Fatal("selected and marked rows must retain row-level backgrounds")
	}
}

func TestRenderTableProjectsPageAndSummaryAtTopRight(t *testing.T) {
	rendered := RenderTable(TableData{
		Cols:       []tables.ColumnDef{{Key: "name", Header: "Name", Fixed: 20}},
		Rows:       [][]string{{"one"}, {"two"}},
		Selected:   0,
		Total:      10,
		Offset:     0,
		Limit:      2,
		BannerW:    40,
		FooterHint: "2 running",
	})
	lines := strings.Split(StripANSI(rendered), "\n")
	if !strings.Contains(lines[0], "1-2/10 │ 2 running") {
		t.Fatalf("top-right summary missing: %q", lines[0])
	}
	if strings.Contains(lines[len(lines)-1], "1-2/10") {
		t.Fatalf("page summary is still duplicated at table bottom: %q", lines[len(lines)-1])
	}
}

func TestTableHitLayoutComputesDataStartRow(t *testing.T) {
	// Given
	layout := NewTableHitLayout(TableHitLayoutInput{
		SelectionPreview:    true,
		SelectionPreviewPad: 1,
		Total:               3,
		BannerLines:         1,
	})

	// When
	start := layout.DataStartRow()

	// Then
	if start != 5 {
		t.Fatalf("data start row = %d, want 5", start)
	}
}

func TestTableHitLayoutRejectsNonDataRows(t *testing.T) {
	// Given
	layout := NewTableHitLayout(TableHitLayoutInput{SelectionPreview: true, Total: 3})

	// When
	_, ok := layout.DataRowAt(layout.DataStartRow()-1, 3)

	// Then
	if ok {
		t.Fatal("row before data area was accepted as a data row")
	}
}

func TestTableHitLayoutMapsRowsWithConfiguredSpacing(t *testing.T) {
	// Given
	previous := GetTableLayout()
	defer ApplyTableLayout(previous)
	updated := previous
	updated.RowSpacing = 1
	ApplyTableLayout(updated)
	layout := NewTableHitLayout(TableHitLayoutInput{Total: 3})
	start := layout.DataStartRow()

	// When
	first, firstOK := layout.DataRowAt(start, 3)
	_, gapOK := layout.DataRowAt(start+1, 3)
	second, secondOK := layout.DataRowAt(start+2, 3)

	// Then
	if !firstOK || first != 0 {
		t.Fatalf("first row = (%d, %t), want (0, true)", first, firstOK)
	}
	if gapOK {
		t.Fatal("row spacing was accepted as a data row")
	}
	if !secondOK || second != 1 {
		t.Fatalf("second row = (%d, %t), want (1, true)", second, secondOK)
	}
}

func TestTableConfigOwnsSharedLayoutAndSemanticColumnStyles(t *testing.T) {
	layout := GetTableLayout()
	if layout.ColumnSpacing != tables.DefaultColumnGap {
		t.Fatalf("table gap = %d, want %d", layout.ColumnSpacing, tables.DefaultColumnGap)
	}
	if layout.RowPrefix == "" || layout.RowPrefixSelected == "" {
		t.Fatalf("table prefixes must come from shared table config: %+v", layout)
	}

	styles := GetColumnStyles([]tables.ColumnDef{{Key: "name"}, {Key: "id"}, {Key: "tag"}, {Key: "unknown"}})
	if len(styles) != 4 {
		t.Fatalf("resolved %d styles, want 4", len(styles))
	}
	if !styles[0].Style.Bold {
		t.Fatal("semantic name style is not bold")
	}
	if styles[0].Style.Color != string(config.FallbackColorTableNameForeground) {
		t.Fatalf("semantic name color = %q, want %q", styles[0].Style.Color, config.FallbackColorTableNameForeground)
	}
	if ansi := cellStyleANSI("name", styles[0].Style); !strings.Contains(ansi, "\x1b[38;2;250;240;230m") {
		t.Fatalf("hex column color was not converted to truecolor ANSI: %q", ansi)
	}
	if !styles[1].Style.Faint {
		t.Fatal("semantic id style is not faint")
	}
	if styles[2].Style.Color != "primary" {
		t.Fatalf("semantic tag color = %q, want primary", styles[2].Style.Color)
	}
	if styles[3].Style != (styleRef{}) {
		t.Fatalf("unknown column should use empty style: %+v", styles[3].Style)
	}
}
