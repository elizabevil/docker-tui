package component

import (
	"github.com/elizabevil/docker-tui/internal/tui/tables"
)

// CellGetter returns the display value for a given column key.
// The idx parameter is the column index in the profile.
type CellGetter[T any] func(item T, col tables.ColumnDef, idx int) string

// BuildRows converts a slice of items into [][]string rows for RenderTable.
// Each row corresponds to one item, with cells populated by the getter.
func BuildRows[T any](items []T, cols []tables.ColumnDef, offset, limit int, getter CellGetter[T]) [][]string {
	rows := make([][]string, 0, limit)
	for i := offset; i < len(items) && len(rows) < limit; i++ {
		cells := make([]string, len(cols))
		for j, cd := range cols {
			cells[j] = getter(items[i], cd, j)
		}
		rows = append(rows, cells)
	}
	return rows
}
