package networks

import (
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/tables"
)

// ColumnDescriptors returns the active column set for the networks panel.
func ColumnDescriptors(m *state.AppModel, width int) []tables.ColumnDef {
	_ = m
	profile := profileSelector.Select(max(width, 42))
	return tc.Columns.Get(profile)
}
