package component

import (
	"testing"

	"github.com/elizabevil/docker-tui/internal/tui/tables"
)

func TestResolveColumnHitPicksColumn(t *testing.T) {
	cols := []tables.ColumnDef{{Key: "name"}, {Key: "state"}, {Key: "created"}}
	widths := []int{20, 10, 15}
	key, ok := ResolveColumnHit(2, 2+21, cols, widths, 2)
	if !ok {
		t.Fatal("column hit not found")
	}
	if key != "state" {
		t.Fatalf("hit key = %q, want state", key)
	}
}

func TestResolveColumnHitGapSnapsToNearest(t *testing.T) {
	cols := []tables.ColumnDef{{Key: "name"}, {Key: "state"}, {Key: "created"}}
	widths := []int{20, 10, 15}
	key, ok := ResolveColumnHit(2, 2+21, cols, widths, 2)
	if !ok || key != "state" {
		t.Fatalf("right of name = (%q, %t), want (state, true)", key, ok)
	}
	key, ok = ResolveColumnHit(2, 2+24, cols, widths, 2)
	if !ok || key != "state" {
		t.Fatalf("inside state = (%q, %t), want (state, true)", key, ok)
	}
}

func TestResolveColumnHitRejectsXBeforeBody(t *testing.T) {
	cols := []tables.ColumnDef{{Key: "name"}}
	widths := []int{20}
	if _, ok := ResolveColumnHit(10, 5, cols, widths, 2); ok {
		t.Fatal("X before body left was accepted as a column hit")
	}
}
