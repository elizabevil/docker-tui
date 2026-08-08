package component

// EnsureVisible adjusts the viewport offset so that the cursor is visible within the visible window.
// offset: pointer to current viewport offset (will be mutated)
// cursor: the currently selected item index
// rowHeight: number of visible rows
// total: total number of items
func EnsureVisible(offset *int, cursor, rowHeight, total int) {
	if cursor < *offset {
		*offset = cursor
	}
	if cursor >= *offset+rowHeight {
		*offset = cursor - rowHeight + 1
	}
	if *offset < 0 {
		*offset = 0
	}
	if *offset+rowHeight > total {
		*offset = max(total-rowHeight, 0)
	}
}
