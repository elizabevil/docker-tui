package images

import (
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/tables"
)

// ColumnDescriptors returns the active column set for the image panel
// based on the available width.
func ColumnDescriptors(m *state.AppModel, width int) []tables.ColumnDef {
	_ = m
	profile := profileSelector.Select(max(width, 0))
	return tc.Columns.Get(profile)
}
