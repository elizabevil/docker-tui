package actionbar

import (
	"sort"
	"strings"

	"github.com/elizabevil/docker-tui/internal/data/config"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"
)

// ActionItem describes a complex operation exposed through Action Bar.
// Key is normally empty: actions with direct shortcuts belong in the footer,
// while Action Bar is reserved for dialog, workflow, or multi-step actions.
//
// The item list is built from the loaded Operations registry
// (R06-03); this struct only carries the display-time view.
type ActionItem struct {
	Key         string
	Label       string
	Action      keys.KeyAction
	Description string
	Disabled    bool
}

// scopeForPanel maps the active panel to the resource scope the action
// bar should project. Adding a new resource scope (Volume / Network /
// Compose) is a one-line addition here plus a matching scope file in
// config/defaults/operations/scopes/.
func scopeForPanel(panel state.PanelType) (config.OperationScope, bool) {
	switch panel {
	case state.PanelContainers:
		return config.OperationScopeContainer, true
	case state.PanelImages:
		return config.OperationScopeImage, true
	case state.PanelVolumes:
		return config.OperationScopeVolume, true
	case state.PanelNetworks:
		return config.OperationScopeNetwork, true
	case state.PanelCompose:
		return config.OperationScopeCompose, true
	default:
		return "", false
	}
}

// VisibleItems returns the action bar items for the active panel,
// filtered by the fuzzy-search buffer. The list is sourced from the
// loaded Operations registry; per-panel selection + disabled
// evaluation are the only logic remaining here.
func VisibleItems(m *state.AppModel) []ActionItem {
	if m == nil {
		return nil
	}
	scope, ok := scopeForPanel(m.Navigation.ActivePanel)
	if !ok {
		return nil
	}
	ops, err := config.CachedLoadOperations()
	if err != nil {
		return nil
	}
	raw := buildItemsForScope(m, ops, scope)
	filter := strings.ToLower(strings.TrimSpace(m.Navigation.ActionBar.Filter))
	if filter == "" {
		return raw
	}
	out := make([]ActionItem, 0, len(raw))
	for _, item := range raw {
		searchable := item.Label + " " + item.Description
		if strings.Contains(strings.ToLower(searchable), filter) {
			out = append(out, item)
		}
	}
	return out
}

// buildItemsForScope projects an Operations slice into the action bar
// display shape, evaluating Requires and DisabledWhen against the
// live AppModel.
func buildItemsForScope(m *state.AppModel, ops *config.Operations, scope config.OperationScope) []ActionItem {
	specs := ops.ForScope(scope)
	items := make([]ActionItem, 0, len(specs))
	for _, spec := range specs {
		if !spec.InActionBar {
			continue
		}
		items = append(items, ActionItem{
			Label:       spec.Label,
			Action:      keys.KeyAction(spec.Action),
			Description: spec.Description,
			Disabled:    !evaluateEnabled(spec, m),
		})
	}
	return items
}

// evaluateEnabled returns true when the Operation should be available
// in the action bar. The decision combines two checks:
//
//	Requires    — AND-combined positives. ALL must be satisfied.
//	DisabledWhen — OR-combined negatives. ANY satisfied means disabled.
//
// Token semantics (closed universe, enforced at load):
//
//	"engine"             — m.Connection.Engine != nil
//	"container"          — active panel is containers and a row is selected
//	"image"              — active panel is images and a row is selected
//	"volume"             — active panel is volumes and a row is selected
//	"network"            — active panel is networks and a row is selected
//	"running"            — selected container is in ContainerStateRunning
//	"manifest"           — selected image summary's IsManifest flag is true
//	"compose_project"    — compose panel + a project is selected
//	"compose_service"    — compose panel focus on services pane + a service is selected
//	"compose_running"    — selected project has ≥1 running container
//	"compose_paused"     — selected project has ≥1 paused container
//	"compose_stopped"    — selected project has ≥1 stopped container and no running
//	"compose_tagged"     — selected project has ≥1 service with explicit image tag
//	"compose_has_ports"  — selected project has ≥1 container exposing a host port
//
// Both lists empty → always enabled.
func evaluateEnabled(spec config.OperationSpec, m *state.AppModel) bool {
	for _, req := range spec.Requires {
		if !positiveSatisfied(req, m) {
			return false
		}
	}
	for _, neg := range spec.DisabledWhen {
		if positiveSatisfied(neg, m) {
			return false
		}
	}
	return true
}

// positiveSatisfied returns true when the named Requirement is met.
// The token set is closed at load time so this is a flat switch —
// any unknown token fails closed (returns false) so a JSONC typo
// disables the action instead of leaking into the UI.
func positiveSatisfied(req config.Requirement, m *state.AppModel) bool {
	switch req {
	case config.RequirementEngine:
		return m.Connection.Engine != nil
	case config.RequirementContainer:
		return m.Navigation.ActivePanel == state.PanelContainers && m.Resources.Containers.Selected() != nil
	case config.RequirementImage:
		return m.Navigation.ActivePanel == state.PanelImages && m.Resources.Images.Selected() != nil
	case config.RequirementVolume:
		return m.Navigation.ActivePanel == state.PanelVolumes && m.Resources.Volumes.Selected() != nil
	case config.RequirementNetwork:
		return m.Navigation.ActivePanel == state.PanelNetworks && m.Resources.Networks.Selected() != nil
	case config.RequirementRunning:
		c := m.Resources.Containers.Selected()
		return c != nil && c.State == state.ContainerStateRunning
	case config.RequirementManifest:
		img := m.Resources.Images.Selected()
		return img != nil && img.IsManifest
	case config.RequirementComposeProject:
		return m != nil && m.Navigation.ActivePanel == state.PanelCompose && selectedComposeProject(m) != ""
	case config.RequirementComposeService:
		return m != nil && m.Navigation.ActivePanel == state.PanelCompose && m.Compose.ComposeFocus == 1 && selectedComposeService(m) != ""
	case config.RequirementComposeRunning:
		return composeProjectHasState(m, state.ContainerStateRunning)
	case config.RequirementComposePaused:
		return composeProjectHasState(m, state.ContainerStatePaused)
	case config.RequirementComposeStopped:
		project := selectedComposeProject(m)
		if project == "" {
			return false
		}
		containers := composeProjectContainers(m, project)
		hasStopped, hasRunning := false, false
		for _, c := range containers {
			switch c.State {
			case state.ContainerStateExited, state.ContainerStateStopped, state.ContainerStateDead:
				hasStopped = true
			case state.ContainerStateRunning:
				hasRunning = true
			}
		}
		return hasStopped && !hasRunning
	case config.RequirementComposeTagged:
		return composeProjectHasTagged(m)
	case config.RequirementComposeHasPorts:
		return composeProjectHasPorts(m)
	default:
		return false
	}
}

// selectedComposeProject returns the project name at the current cursor,
// or "" if no project is selectable. Mirrors keyboard.currentComposeProject
// without crossing the package boundary (actionbar must not import keyboard
// to avoid an import cycle).
func selectedComposeProject(m *state.AppModel) string {
	if m == nil || m.Resources.Containers.Items == nil {
		return ""
	}
	filter := m.Compose.ComposeProjectFilter
	names := composeProjectNames(m, filter)
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

// selectedComposeService returns the service name at the current cursor
// when the right pane is focused, "" otherwise.
func selectedComposeService(m *state.AppModel) string {
	if m == nil || m.Compose.ComposeFocus != 1 {
		return ""
	}
	project := selectedComposeProject(m)
	if project == "" {
		return ""
	}
	filter := m.Compose.ComposeServiceFilter
	services := composeServiceNames(m, project, filter)
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

// composeProjectNames derives the sorted, filter-aware project name list
// from container summaries. The keyboard package owns the same logic
// (composeProjectNames there); duplicating here keeps the actionbar
// package free of an import cycle.
func composeProjectNames(m *state.AppModel, filter string) []string {
	set := make(map[string]struct{})
	for _, c := range m.Resources.Containers.Items {
		if c.ComposeProject == "" {
			continue
		}
		if filter != "" && !strings.Contains(c.ComposeProject, filter) {
			continue
		}
		set[c.ComposeProject] = struct{}{}
	}
	out := make([]string, 0, len(set))
	for name := range set {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

// composeServiceNames derives the sorted, filter-aware service name list
// for the given project.
func composeServiceNames(m *state.AppModel, project, filter string) []string {
	set := make(map[string]struct{})
	for _, c := range m.Resources.Containers.Items {
		if c.ComposeProject != project {
			continue
		}
		name := c.ComposeService
		if name == "" {
			name = "unknown"
		}
		if filter != "" && !strings.Contains(name, filter) {
			continue
		}
		set[name] = struct{}{}
	}
	out := make([]string, 0, len(set))
	for name := range set {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

// composeProjectContainers returns every ContainerSummary belonging to
// the selected project. The returned slice is a fresh copy so callers
// can range without aliasing m.Resources.
func composeProjectContainers(m *state.AppModel, project string) []runtimeapi.ContainerSummary {
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

// composeProjectHasState reports whether the selected project contains
// at least one container in the given state.
func composeProjectHasState(m *state.AppModel, want string) bool {
	project := selectedComposeProject(m)
	if project == "" {
		return false
	}
	for _, c := range composeProjectContainers(m, project) {
		if c.State == want {
			return true
		}
	}
	return false
}

// composeProjectHasTagged reports whether the selected project has at
// least one service whose image is pinned with an explicit tag other
// than ":latest". The heuristic uses the com.docker.compose.image
// label, which docker / podman write when the compose file declares
// `image:` with a non-empty tag.
func composeProjectHasTagged(m *state.AppModel) bool {
	project := selectedComposeProject(m)
	if project == "" {
		return false
	}
	for _, c := range composeProjectContainers(m, project) {
		img := imageFromComposeLabel(c)
		if !strings.Contains(img, ":") {
			continue
		}
		tag := img[strings.LastIndex(img, ":")+1:]
		if tag == "" || tag == "latest" {
			continue
		}
		if tag == img {
			// registry:port form (no tag component); treat as tagged.
			return true
		}
		return true
	}
	return false
}

// composeProjectHasPorts reports whether the selected project has any
// container exposing at least one host port binding.
func composeProjectHasPorts(m *state.AppModel) bool {
	project := selectedComposeProject(m)
	if project == "" {
		return false
	}
	for _, c := range composeProjectContainers(m, project) {
		if len(c.PortBindings) > 0 {
			return true
		}
	}
	return false
}

// imageFromComposeLabel returns the image reference declared by the
// compose file (com.docker.compose.image label) when present, falling
// back to the container's effective Image so untagged builds still
// surface a value to inspect.
func imageFromComposeLabel(c runtimeapi.ContainerSummary) string {
	if c.Labels != nil {
		if img := c.Labels[runtimeapi.ComposeLabelImage]; img != "" {
			return img
		}
	}
	return c.Image
}