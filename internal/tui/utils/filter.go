// Package utils provides tui-level helper functions shared across packages.
//
// The helpers in this package must be pure (no state mutations, no
// side-effects beyond argument inspection) so they can be safely used
// from any tui subpackage without introducing circular dependencies.
package utils

import "github.com/elizabevil/docker-tui/internal/tui/state"

// ActiveTableFilter returns the TableFilter for the currently active
// panel, or nil if no panel supports filtering (or m is nil).
//
// Panels supported: Containers, Images, Volumes, Networks, Audit.
// Compose is handled separately by the filter.Controller because it
// uses a different field layout (m.Compose.ComposeServiceFilter /
// ComposeProjectFilter).
func ActiveTableFilter(m *state.AppModel) state.TableFilter {
	if m == nil {
		return nil
	}
	switch m.Navigation.ActivePanel {
	case state.PanelContainers:
		return m.Resources.Containers
	case state.PanelImages:
		return m.Resources.Images
	case state.PanelVolumes:
		return m.Resources.Volumes
	case state.PanelNetworks:
		return m.Resources.Networks
	case state.PanelAudit:
		return &m.Audit
	default:
		return nil
	}
}
