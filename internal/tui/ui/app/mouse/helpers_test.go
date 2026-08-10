package mouse

import (
	"github.com/elizabevil/docker-tui/internal/tui/tables"
)

func columnX(bodyLeft int, cols []tables.ColumnDef, widths []int, gap int, key string) (int, bool) {
	cursor := 0
	for i, def := range cols {
		if def.Key == key {
			return bodyLeft + cursor, true
		}
		cursor += widths[i]
		if i < len(widths)-1 {
			cursor += gap
		}
	}
	return 0, false
}
