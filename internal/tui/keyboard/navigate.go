package keyboard

import (
	"fmt"

	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

// ── Sub-view navigation (drill-down / back) ──────────────────

// ToImageContainers enters the container sub-view for the selected image.
// Sets ContainersViewID so the images panel renders containers using that image.
func ToImageContainers(m *state.AppModel, imgID string) {
	m.Images.ContainersViewID = imgID
	m.Images.ContainerCursor = 0
	m.InfoMessage = fmt.Sprintf("Containers using %s", shortID(imgID))
}

// BackFromImageContainers leaves the container sub-view.
func BackFromImageContainers(m *state.AppModel) {
	m.Images.ContainersViewID = ""
	m.Images.ContainerCursor = 0
	m.InfoMessage = ""
}

// ToVolumeDetail enters the container sub-view for the selected volume.
func ToVolumeDetail(m *state.AppModel, volName string) {
	m.Volumes.DetailName = volName
	m.Volumes.Cursor = 0
	m.InfoMessage = "Volume: " + volName
}

// BackFromVolumeDetail leaves the volume container sub-view.
func BackFromVolumeDetail(m *state.AppModel) {
	m.Volumes.DetailName = ""
	m.InfoMessage = ""
}

// ── Mode transitions ────────────────────────────────────────

// ToLogView opens the log viewer for a container.
func ToLogView(m *state.AppModel, containerID string) tea.Cmd {
	m.Mode = state.ModeLogView
	m.LogContainerID = containerID
	m.LogContent = m.LogContent[:0]
	m.LogViewOffset = 0
	m.LogSearchText = ""
	m.LogSearchMatch = 0
	cfg := m.Config.Logs
	return FetchLogBatch(m.Docker, containerID, cfg.Since, cfg.Tail, cfg.Timestamps)
}

// BackFromLogView closes the log viewer.
func BackFromLogView(m *state.AppModel) {
	m.Mode = state.ModeNormal
	m.LogContainerID = ""
	m.LogContent = nil
	m.LogViewOffset = 0
	m.LogSearchText = ""
}

// ToDetail opens the detail inspector for the current item.
func ToDetail(m *state.AppModel, title, content string) {
	m.PrevPanel = m.ActivePanel
	m.Mode = state.ModeDetail
	m.DetailTitle = title
	m.ImageDetailContent = content
	m.ImageDetailData = nil
	m.DetailOffset = 0
	m.DetailHint = ""
}

// BackFromDetail closes the detail inspector and restores the previous panel.
func BackFromDetail(m *state.AppModel) {
	m.Mode = state.ModeNormal
	if m.PrevPanel != m.ActivePanel {
		m.ActivePanel = m.PrevPanel
	}
	m.ImageDetailID = ""
	m.ImageDetailContent = ""
	m.ImageDetailData = nil
	m.DetailTitle = ""
	m.DetailHint = ""
	m.DetailOffset = 0
}

// ToHelp opens the help screen.
func ToHelp(m *state.AppModel) {
	m.PrevPanel = m.ActivePanel
	m.Mode = state.ModeHelp
	m.ActivePanel = state.PanelHelp
}

// BackFromHelp closes the help screen.
func BackFromHelp(m *state.AppModel) {
	m.Mode = state.ModeNormal
	m.ActivePanel = m.PrevPanel
}

// ToFilter opens the search filter bar.
func ToFilter(m *state.AppModel) tea.Cmd {
	m.Mode = state.ModeFilter
	if m.ActivePanel == state.PanelCompose {
		if m.ComposeFocus == 1 {
			m.FilterText = m.ComposeServiceFilter
		} else {
			m.FilterText = m.ComposeProjectFilter
		}
	} else {
		m.FilterText = ""
	}
	m.FilterCursor = 0
	m.FilterCursor = len([]rune(m.FilterText))
	m.SearchTimer = 0
	return nil
}

// BackFromFilter closes the search filter bar.
func BackFromFilter(m *state.AppModel) {
	m.Mode = state.ModeNormal
	m.FilterText = ""
	m.FilterCursor = 0
}

// ToCommand opens the command palette.
func ToCommand(m *state.AppModel) {
	m.Mode = state.ModeCommand
	m.FilterText = ""
	m.FilterCursor = 0
}

// BackFromCommand closes the command palette.
func BackFromCommand(m *state.AppModel) {
	m.Mode = state.ModeNormal
	m.FilterText = ""
	m.FilterCursor = 0
}

// ToExec opens the exec shell dialog.
func ToExec(m *state.AppModel) {
	m.Mode = state.ModeExec
	m.DialogFocus = 0
	m.DialogCursor = 0
	m.FilterText = "/bin/sh"
}

// BackFromExec closes the exec shell dialog.
func BackFromExec(m *state.AppModel) {
	m.Mode = state.ModeNormal
	m.DialogTitle = ""
	m.DialogBody = ""
	m.DialogAction = ""
	m.DialogFocus = 0
	m.DialogCursor = 0
}
