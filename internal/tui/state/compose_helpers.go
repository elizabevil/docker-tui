package state

import (
	"sort"
	"strings"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

// composeAggregations groups every helper that derives a project /
// service / container slice from m.Resources. Both the keyboard layer
// and the action-bar registry need the same projections, so the
// functions live here next to the AppModel they read.
//
// All functions take *AppModel by pointer because several of them
// re-clamp m.Compose.Compose{Cursor,ServiceCursor} when the cursor
// has drifted past the current list. The receiver is named m per
// the keyboard layer convention.

type composeProjectSet map[string]struct{}
type composeServiceSet map[string]struct{}

// ComposeProjectNames returns the sorted, filter-aware list of compose
// projects visible to the panel. The result is built from the
// container label aggregator; an empty project list means no
// compose-managed container is present (R08-03 F2: empty projects
// stay hidden, matching `docker compose ls` semantics).
func ComposeProjectNames(m *AppModel) []string {
	if m == nil {
		return nil
	}
	filter := m.Compose.ComposeProjectFilter
	set := make(composeProjectSet)
	for _, c := range m.Resources.Containers.Items {
		project := c.ComposeProject
		if project == "" {
			continue
		}
		if filter != "" && !strings.Contains(project, filter) {
			continue
		}
		set[project] = struct{}{}
	}
	out := sortedComposeKeys(set)
	return out
}

// ComposeServiceNames returns the sorted, filter-aware list of service
// names within the given project. Containers without a service label
// are bucketed under "unknown" so they remain visible in the panel.
func ComposeServiceNames(m *AppModel, project string) []string {
	if m == nil {
		return nil
	}
	filter := m.Compose.ComposeServiceFilter
	set := make(composeServiceSet)
	for _, c := range m.Resources.Containers.Items {
		if c.ComposeProject != project {
			continue
		}
		name := c.ComposeService
		if name == "" {
			name = composeUnknownService
		}
		if filter != "" && !strings.Contains(name, filter) {
			continue
		}
		set[name] = struct{}{}
	}
	return sortedComposeKeys(set)
}

// SelectedComposeProject returns the project name at the current
// cursor (re-clamped when out of range). Empty means "no project
// selectable".
func SelectedComposeProject(m *AppModel) string {
	names := ComposeProjectNames(m)
	if len(names) == 0 {
		return ""
	}
	idx := m.Compose.ComposeCursor
	if idx >= len(names) {
		idx = len(names) - 1
	}
	if idx < 0 {
		idx = 0
	}
	return names[idx]
}

// SelectedComposeService returns the service name at the current
// cursor when the right pane is focused (ComposeFocus == 1). Empty
// otherwise — callers rely on this to gate service-scoped actions.
func SelectedComposeService(m *AppModel, project string) string {
	if m == nil || m.Compose.ComposeFocus != 1 {
		return ""
	}
	services := ComposeServiceNames(m, project)
	if len(services) == 0 {
		return ""
	}
	idx := m.Compose.ComposeServiceCursor
	if idx >= len(services) {
		idx = len(services) - 1
	}
	if idx < 0 {
		idx = 0
	}
	return services[idx]
}

// ComposeProjectContainers returns every ContainerSummary whose
// compose project label matches the given project. The returned slice
// is a fresh copy so callers may range without aliasing the model.
func ComposeProjectContainers(m *AppModel, project string) []runtimeapi.ContainerSummary {
	if m == nil {
		return nil
	}
	out := make([]runtimeapi.ContainerSummary, 0)
	for _, c := range m.Resources.Containers.Items {
		if c.ComposeProject == project {
			out = append(out, c)
		}
	}
	return out
}

// ComposeProjectVolumes returns the names of project-labeled volumes
// (Volumes whose `com.docker.compose.project` label matches). The
// returned slice is a fresh copy.
func ComposeProjectVolumes(m *AppModel, project string) []string {
	if m == nil {
		return nil
	}
	want := runtimeapi.ComposeProjectLabelValue(project)
	out := make([]string, 0)
	for _, v := range m.Resources.Volumes.Items {
		if v.Labels == nil {
			continue
		}
		if v.Labels[runtimeapi.ComposeLabelProject] == project || v.Labels[runtimeapi.ComposeLabelProject] == strings.TrimPrefix(want, runtimeapi.ComposeLabelProject+"=") {
			out = append(out, v.Name)
		}
	}
	return out
}

// ComposeProjectNetworks returns the IDs of project-labeled networks
// matching the given project. The returned slice is a fresh copy.
func ComposeProjectNetworks(m *AppModel, project string) []string {
	if m == nil {
		return nil
	}
	out := make([]string, 0)
	for _, n := range m.Resources.Networks.Items {
		if n.Labels == nil {
			continue
		}
		if n.Labels[runtimeapi.ComposeLabelProject] == project {
			out = append(out, n.ID)
		}
	}
	return out
}

// ComposeGroupContainers filters the project's containers by an
// optional CoLocated group ID (R08-14). Empty groupID = "all
// containers" so the function doubles as the project-wide iterator
// for the keyboard layer's group-scoped verbs.
func ComposeGroupContainers(m *AppModel, project, groupID string) []runtimeapi.ContainerSummary {
	all := ComposeProjectContainers(m, project)
	if groupID == "" {
		return all
	}
	out := make([]runtimeapi.ContainerSummary, 0, len(all))
	for _, c := range all {
		if c.CoLocatedGroupID == groupID {
			out = append(out, c)
		}
	}
	return out
}

// sortedComposeKeys turns any string-keyed set into a sorted slice.
// The set type is a tiny named alias so the helper accepts both
// composeProjectSet and composeServiceSet callers without re-typing
// the sort boilerplate at every site.
func sortedComposeKeys[T ~map[string]struct{}](set T) []string {
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

const composeUnknownService = "unknown"
