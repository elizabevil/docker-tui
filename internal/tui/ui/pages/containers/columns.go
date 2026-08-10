package containers

import (
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/tables"
)

// ColumnDescriptors returns the active column set for the containers
// panel based on the available width and the table profile selector.
func ColumnDescriptors(m *state.AppModel, width int) []tables.ColumnDef {
	profile := profileSelector.Select(max(width, 52))
	return tc.Columns.Get(profile)
}
