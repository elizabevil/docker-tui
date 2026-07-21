package keyboard

import (
	"strings"

	"github.com/elizabevil/docker-tui/internal/data/docker"
	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

// handleImagePanelKeys handles image-panel-specific key bindings.
// Returns nil, nil if key not handled.
func handleImagePanelKeys(key string, m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.ActivePanel != state.PanelImages {
		return nil, nil
	}

	// ── Container sub-view (ContainersViewID is set) ─────────
	if m.Images.ContainersViewID != "" {
		action, known := resolveAction(key, m)
		if !known {
			if key == keys.KeyLeft {
				return doImageCollapse(m)
			}
			return m, nil
		}
		switch action {
		case keys.ActionDown:
			m.Images.ContainerCursor++
			return m, nil
		case keys.ActionUp:
			if m.Images.ContainerCursor > 0 {
				m.Images.ContainerCursor--
			}
			return m, nil
		case keys.ActionContainerExec:
			syncContainerCursorForSubView(m)
			m.Mode = state.ModeExec
			m.DialogState.OpenInput("/bin/sh", 0)
			return m, RecordKeyStroke(m, key, "Exec")
		case keys.ActionContainerStart:
			mm, cmd := doImageSubContainerCmd(m, containerStartCmd)
			return mm, tea.Batch(cmd, RecordKeyStroke(m, key, "Start"))
		case keys.ActionContainerStop:
			mm, cmd := doImageSubContainerCmd(m, func(c *docker.Client, id string) tea.Cmd { return containerStopCmd(c, id) })
			return mm, tea.Batch(cmd, RecordKeyStroke(m, key, "Stop"))
		case keys.ActionContainerRestart:
			mm, cmd := doImageSubContainerCmd(m, func(c *docker.Client, id string) tea.Cmd { return containerRestartCmd(c, id) })
			return mm, tea.Batch(cmd, RecordKeyStroke(m, key, "Restart"))
		case keys.ActionEnter, keys.ActionContainerLogs:
			mm, cmd := doImageContainerLog(m)
			return mm, tea.Batch(cmd, RecordKeyStroke(m, key, "Logs"))
		case keys.ActionBack:
			return doImageCollapse(m)
		}
		return m, nil
	}

	// ── Image list view ─────────────────────────────────────
	if key == keys.KeyCtrlB {
		return doImageDebug(m)
	}
	if key == keys.KeyCtrlE {
		return doImageExport(m)
	}
	if key == keys.KeyY {
		return doImageCopyRef(m)
	}
	if key == keys.KeyRight {
		return doImageExpand(m)
	}
	if key == keys.KeyLeft {
		return doImageCollapse(m)
	}
	return nil, nil
}

func imageSubContainerID(m *state.AppModel) string {
	if m.Docker == nil || m.Images.ContainersViewID == "" {
		return ""
	}
	imgShort := m.Images.ContainersViewID[:12]
	var imgNames []string
	for _, item := range m.Images.Items {
		if item.ID == m.Images.ContainersViewID || item.ID[:12] == imgShort {
			for _, tag := range item.RepoTags {
				imgNames = append(imgNames, tag)
			}
			break
		}
	}
	cursor := 0
	for _, c := range m.Containers.Items {
		match := strings.Contains(c.Image, imgShort)
		if !match {
			for _, n := range imgNames {
				if strings.Contains(c.Image, n) {
					match = true
					break
				}
			}
		}
		if !match {
			continue
		}
		if cursor == m.Images.ContainerCursor {
			return c.ID
		}
		cursor++
	}
	return ""
}

func syncContainerCursorForSubView(m *state.AppModel) {
	id := imageSubContainerID(m)
	if id == "" {
		return
	}
	for i, c := range m.Containers.FilteredItems() {
		if c.ID == id {
			m.Containers.Cursor = i
			return
		}
	}
}

func doImageSubContainerCmd(m *state.AppModel, cmdFn func(*docker.Client, string) tea.Cmd) (*state.AppModel, tea.Cmd) {
	if m.Docker == nil {
		return m, nil
	}
	id := imageSubContainerID(m)
	if id == "" {
		return m, nil
	}
	return m, cmdFn(m.Docker, id)
}

// doImageContainerLog opens the log view for the selected container in the image sub-view.
func doImageContainerLog(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Docker == nil {
		return m, nil
	}
	// Find the container at ContainerCursor matching ContainersViewID
	imgShort := m.Images.ContainersViewID[:12]
	var matchedID string
	var imgNames []string
	for _, item := range m.Images.Items {
		if item.ID == m.Images.ContainersViewID || item.ID[:12] == imgShort {
			for _, tag := range item.RepoTags {
				imgNames = append(imgNames, tag)
			}
			break
		}
	}
	cursor := 0
	for _, c := range m.Containers.Items {
		match := strings.Contains(c.Image, imgShort)
		if !match {
			for _, n := range imgNames {
				if strings.Contains(c.Image, n) {
					match = true
					break
				}
			}
		}
		if !match {
			continue
		}
		if cursor == m.Images.ContainerCursor {
			matchedID = c.ID
			break
		}
		cursor++
	}
	if matchedID == "" {
		return m, nil
	}
	m.LogContainerID = matchedID
	m.LogContent = m.LogContent[:0]
	m.Mode = state.ModeLogView
	cfg := m.Config.Logs
	return m, FetchLogBatch(m.Docker, matchedID, cfg.Since, cfg.Tail, cfg.Timestamps)
}
