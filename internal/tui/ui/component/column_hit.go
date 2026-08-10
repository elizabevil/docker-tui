package component

import "github.com/elizabevil/docker-tui/internal/tui/tables"

// ResolveColumnHit returns the column whose header occupies the given X
// position. bodyLeft is the absolute panel-body column where rendering
// starts (matches rep.Panel.bodyLeft). widths must be the resolved column
// widths in render order; gap is the column gutter width. When x falls
// inside the gutter between two columns the closer column wins so the
// whole header row reacts to clicks.
func ResolveColumnHit(bodyLeft, x int, cols []tables.ColumnDef, widths []int, gap int) (string, bool) {
	if x < bodyLeft {
		return "", false
	}
	col := x - bodyLeft
	if col < 0 {
		return "", false
	}
	cursor := 0
	prev := -1
	for i, width := range widths {
		if i >= len(cols) {
			break
		}
		end := cursor + width
		if col <= end {
			return cols[i].Key, true
		}
		prev = i
		cursor = end
		if i < len(widths)-1 {
			cursor += gap
		}
	}
	if prev >= 0 && prev < len(cols) {
		return cols[prev].Key, true
	}
	return "", false
}
