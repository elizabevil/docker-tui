package tables

import "testing"

func TestResolveLayoutFillsViewportAcrossFlexibleColumns(t *testing.T) {
	columns := []ColumnDef{
		{Key: "id", Fixed: 12},
		{Key: "name", Basis: 24, Min: 12, Max: 32, Shrink: 1, Fill: 1},
		{Key: "stats", Basis: 30, Min: 18, Shrink: 1, Fill: 1},
	}
	layout := ResolveLayout(columns, 100, 2)
	if layout.Widths[0] != 12 || layout.Widths[1] != 32 || layout.Widths[2] != 52 {
		t.Fatalf("widths = %v, want [12 32 52]", layout.Widths)
	}
	if layout.ContentWidth != 100 || layout.TrailingWidth != 0 {
		t.Fatalf("content=%d trailing=%d, want 100/0", layout.ContentWidth, layout.TrailingWidth)
	}
}

func TestResolveContentLayoutPrioritizesTruncatedCells(t *testing.T) {
	columns := []ColumnDef{
		{Key: "registry", Basis: 12, Min: 8, Shrink: 1, Fill: 1},
		{Key: "name", Basis: 24, Min: 12, Shrink: 1},
		{Key: "id", Basis: 12, Min: 10, Shrink: 1},
	}
	layout := ResolveContentLayout(columns, []int{24, 20, 32}, 86, 2)
	if layout.ContentWidth+layout.TrailingWidth != 86 || layout.TrailingWidth == 0 {
		t.Fatalf("layout does not preserve trailing viewport space: %+v", layout)
	}
	if layout.Widths[0] < 24 || layout.Widths[2] < 32 {
		t.Fatalf("truncated content was not expanded first: %v", layout.Widths)
	}
}

func TestColumnWidthsUsesSharedGridResolver(t *testing.T) {
	config := MustLoad("containers")
	columns := config.Columns.Get("more_stats")
	direct := ResolveColumnWidths(columns, 158, DefaultColumnGap)
	profile := config.ColumnWidths("more_stats", 158)
	if len(direct) != len(profile) {
		t.Fatalf("width count direct=%d profile=%d", len(direct), len(profile))
	}
	for i := range direct {
		if direct[i] != profile[i] {
			t.Fatalf("profile diverged at column %d: direct=%v profile=%v", i, direct, profile)
		}
	}
}

func TestEveryTableProfileRespectsGridConstraints(t *testing.T) {
	for _, table := range []string{"containers", "images", "volumes", "networks", "compose", "audit"} {
		config := MustLoad(table)
		for _, profile := range config.Columns.All() {
			columns := profile.Columns
			minimumWidth := DefaultColumnGap * max(0, len(columns)-1)
			preferredWidth := minimumWidth
			for _, column := range columns {
				spec := resolveColumnSizing(column)
				minimumWidth += spec.min
				preferredWidth += spec.basis
			}

			for _, viewportWidth := range []int{minimumWidth, preferredWidth, preferredWidth + 80} {
				layout := ResolveLayout(columns, viewportWidth, DefaultColumnGap)
				if layout.ContentWidth+layout.TrailingWidth != viewportWidth {
					t.Errorf("%s/%s viewport=%d content=%d trailing=%d", table, profile.Name, viewportWidth, layout.ContentWidth, layout.TrailingWidth)
				}
				unboundedFlexible := false
				for i, columnWidth := range layout.Widths {
					spec := resolveColumnSizing(columns[i])
					if columnWidth < spec.min {
						t.Errorf("%s/%s column %s width %d below min %d", table, profile.Name, columns[i].Key, columnWidth, spec.min)
					}
					if spec.max > 0 && columnWidth > spec.max {
						t.Errorf("%s/%s column %s width %d above max %d", table, profile.Name, columns[i].Key, columnWidth, spec.max)
					}
					unboundedFlexible = unboundedFlexible || spec.fill > 0 && spec.max == 0
				}
				if viewportWidth > preferredWidth && unboundedFlexible && layout.TrailingWidth != 0 {
					t.Errorf("%s/%s left %d cells unused in wide viewport", table, profile.Name, layout.TrailingWidth)
				}
			}
		}
	}
}

func TestResolveContentLayoutDoesNotStretchSparseColumns(t *testing.T) {
	columns := []ColumnDef{
		{Key: "name", Basis: 20, Min: 12, Shrink: 1, Fill: 1},
		{Key: "ports", Basis: 12, Min: 8, Max: 32, Shrink: 1},
		{Key: "ip", Basis: 15, Min: 12, Max: 39, Shrink: 1},
	}
	layout := ResolveContentLayout(columns, []int{15, 1, 1}, 80, 2)
	if layout.Widths[1] != 12 || layout.Widths[2] != 15 {
		t.Fatalf("sparse columns absorbed fill space: %v", layout.Widths)
	}
	if layout.ContentWidth != 51 || layout.TrailingWidth != 29 {
		t.Fatalf("sparse layout should leave trailing space: %+v", layout)
	}
}

func TestResolveLayoutShrinksFromPreferredToMin(t *testing.T) {
	columns := []ColumnDef{
		{Key: "name", Basis: 20, Min: 10, Shrink: 1},
		{Key: "image", Basis: 20, Min: 12, Shrink: 1},
		{Key: "created", Fixed: 19},
	}
	layout := ResolveLayout(columns, 50, 2)
	if layout.ContentWidth != 50 || layout.TrailingWidth != 0 {
		t.Fatalf("content=%d trailing=%d, want 50/0: %v", layout.ContentWidth, layout.TrailingWidth, layout.Widths)
	}
	if layout.Widths[0] < 10 || layout.Widths[1] < 12 {
		t.Fatalf("column shrank below min: %v", layout.Widths)
	}
}

// TestPodsSubHasPodColumn pins the §4.3 pod-column contract: the
// container sub-view's pods_sub profile carries a "pod" column whose
// header routes through the i18n layer. Visibility (whether the column
// is rendered at all) is decided by renderComposeContainers based on
// the runtime capability — this test only pins that the column
// definition exists in YAML.
func TestPodsSubHasPodColumn(t *testing.T) {
	tc := MustLoad("compose")
	cols := tc.Columns.Get("pods_sub")
	var found bool
	for _, c := range cols {
		if c.Key == "pod" {
			found = true
			if c.Header != "table.pod" {
				t.Fatalf("pod column header = %q, want %q", c.Header, "table.pod")
			}
		}
	}
	if !found {
		t.Fatalf("pods_sub missing \"pod\" column; got %d cols", len(cols))
	}
}
