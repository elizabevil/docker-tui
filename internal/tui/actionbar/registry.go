package actionbar

import (
	"strings"

	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"
)

// ActionItem describes a complex operation exposed through Action Bar.
// Key is normally empty: actions with direct shortcuts belong in the footer,
// while Action Bar is reserved for dialog, workflow, or multi-step actions.
type ActionItem struct {
	Key         string
	Label       string
	Action      keys.KeyAction
	Description string
	Disabled    bool
}

func VisibleItems(m *state.AppModel) []ActionItem {
	if m == nil {
		return nil
	}
	raw := actionsForPanel(m)
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

func actionsForPanel(m *state.AppModel) []ActionItem {
	switch m.Navigation.ActivePanel {
	case state.PanelImages:
		return imageActions(m)
	case state.PanelContainers:
		return containerActions(m)
	default:
		return nil
	}
}

func containerActions(m *state.AppModel) []ActionItem {
	container := m.Resources.Containers.Selected()
	missingContainer := container == nil
	missingEngine := m.Connection.Engine == nil
	topDisabled := missingEngine || missingContainer || container.State != state.ContainerStateRunning
	return []ActionItem{
		{
			Label: "Rename", Action: keys.ActionContainerRename,
			Description: "Open the container rename dialog", Disabled: missingEngine || missingContainer,
		},
		{
			Label: "Top", Action: keys.ActionContainerTop,
			Description: "Open the running container process view", Disabled: topDisabled,
		},
		{
			Label: "Port", Action: keys.ActionContainerPort,
			Description: "Open structured container port mappings", Disabled: missingContainer,
		},
		{
			Label: "Filesystem Diff", Action: keys.ActionContainerDiff,
			Description: "Inspect filesystem changes from the base image", Disabled: missingEngine || missingContainer,
		},
		{
			Label: "Wait", Action: keys.ActionContainerWait,
			Description: "Wait until the container is no longer running", Disabled: missingEngine || missingContainer,
		},
		{
			Label: "Copy", Action: keys.ActionContainerCopy,
			Description: "Copy a file or directory out of the container as a tar", Disabled: missingEngine || missingContainer,
		},
		{
			Label: "Update", Action: keys.ActionContainerUpdate,
			Description: "Change memory, CPU and restart-policy resources", Disabled: missingEngine || missingContainer,
		},
		{
			Label: "Export", Action: keys.ActionContainerExport,
			Description: "Export the container filesystem as a tar", Disabled: missingEngine || missingContainer,
		},
		{
			Label: "Commit", Action: keys.ActionContainerCommit,
			Description: "Snapshot the container as a new image", Disabled: missingEngine || missingContainer,
		},
	}
}

func imageActions(m *state.AppModel) []ActionItem {
	summary := m.Resources.Images.Selected()
	missingImage := summary == nil
	missingEngine := m.Connection.Engine == nil
	return []ActionItem{
		{
			Label: "Image History", Action: keys.ActionImageHistory,
			Description: "View per-layer build history", Disabled: missingEngine || missingImage || summary.IsManifest,
		},
	}
}
