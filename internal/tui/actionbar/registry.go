package actionbar

import (
	"strings"

	"github.com/elizabevil/docker-tui/internal/data/config"
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
//	"engine"     — m.Connection.Engine != nil
//	"container"  — active panel is containers and a row is selected
//	"image"      — active panel is images and a row is selected
//	"volume"     — active panel is volumes and a row is selected
//	"network"    — active panel is networks and a row is selected
//	"running"    — selected container is in ContainerStateRunning
//	"manifest"   — selected image summary's IsManifest flag is true
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
	default:
		return false
	}
}