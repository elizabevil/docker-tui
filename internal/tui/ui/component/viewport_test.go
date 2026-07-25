package component

import (
	"testing"
)

func TestEnsureVisible(t *testing.T) {
	tests := []struct {
		name      string
		offset    int
		cursor    int
		rowHeight int
		total     int
		want      int
	}{
		{"cursor in view", 0, 3, 10, 20, 0},
		{"cursor above viewport", 10, 2, 5, 20, 2},
		{"cursor below viewport", 0, 12, 5, 20, 8},
		{"cursor at top", 5, 0, 5, 20, 0},
		{"cursor at bottom", 0, 19, 5, 20, 15},
		{"small list", 0, 2, 10, 3, 0},
		{"offset would go negative", 2, 0, 5, 10, 0},
		{"offset would overflow", 15, 9, 5, 10, 5},
	}
	for _, tc := range tests {
		offset := tc.offset
		EnsureVisible(&offset, tc.cursor, tc.rowHeight, tc.total)
		if offset != tc.want {
			t.Errorf("%s: EnsureVisible(offset=%d, cursor=%d, rowHeight=%d, total=%d) = %d, want %d",
				tc.name, tc.offset, tc.cursor, tc.rowHeight, tc.total, offset, tc.want)
		}
	}
}
