package compose

import (
	"sort"
	"strings"

	"github.com/elizabevil/docker-tui/internal/tui/state"
)

// ServiceNamesFor returns the sorted, filter-aware service names for
// the active project. Mirrors renderServicePanel so the mouse click
// handler can route to the same items the table renders.
func ServiceNamesFor(m *state.AppModel) []string {
	if m == nil {
		return nil
	}
	ordered, _ := gatherComposeProjects(m)
	ordered = filterProjects(ordered, m.Compose.ComposeProjectFilter)
	if len(ordered) == 0 {
		return nil
	}
	idx := m.Compose.ProjectCursor(len(ordered))
	proj := ordered[idx]
	names := make([]string, 0, len(proj.svcs))
	for name := range proj.svcs {
		if m.Compose.ComposeServiceFilter != "" && !strings.Contains(name, m.Compose.ComposeServiceFilter) {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
