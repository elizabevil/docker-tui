package component

import (
	"strings"
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/config"
	"github.com/elizabevil/docker-tui/internal/tui/tables"
)

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
