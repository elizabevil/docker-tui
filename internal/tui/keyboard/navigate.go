package keyboard

import (
	"github.com/elizabevil/docker-tui/internal/data/config"
	"github.com/elizabevil/docker-tui/internal/tui/filter"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

// ── Sub-view navigation (drill-down / back) ──────────────────

// ToImageContainers enters the container sub-view for the selected image.
// Sets ContainersViewID so the images panel renders containers using that image.
func ToImageContainers(m *state.AppModel, imgID, imageRef string) {
	m.Resources.Images.ContainersViewID = imgID
	m.Resources.Images.ContainersViewRef = imageRef
	m.Resources.Images.ContainerCursor = 0
}

// BackFromImageContainers leaves the container sub-view.
func BackFromImageContainers(m *state.AppModel) {
	m.Resources.Images.ContainersViewID = ""
	m.Resources.Images.ContainersViewRef = ""
	m.Resources.Images.ContainerCursor = 0
	m.Feedback.InfoMessage = ""
}

// ToVolumeDetail enters the container sub-view for the selected volume.
func ToVolumeDetail(m *state.AppModel, volName string) {
	m.Resources.Volumes.DetailName = volName
	m.Resources.Volumes.Cursor = 0
	m.Feedback.InfoMessage = "Volume: " + volName
}

// BackFromVolumeDetail leaves the volume container sub-view.
func BackFromVolumeDetail(m *state.AppModel) {
	m.Resources.Volumes.DetailName = ""
	m.Feedback.InfoMessage = ""
}

// ── Mode transitions ────────────────────────────────────────

// ToLogView opens the log viewer for a container.
func ToLogView(m *state.AppModel, containerID string) tea.Cmd {
	m.Navigation.Mode = state.ModeLogView
	m.Log.Open(containerID)
	cfg := m.Dependencies.Config.Logs
	return FetchLogBatch(m.Connection.Engine, containerID, cfg.Since, cfg.Tail, cfg.Timestamps)
}

// BackFromLogView closes the log viewer.
func BackFromLogView(m *state.AppModel) {
	m.Navigation.Mode = state.ModeNormal
	m.Log.Close()
}

// ToDetail opens the detail inspector for the current item.
func ToDetail(m *state.AppModel, title, content string) {
	m.Navigation.PrevPanel = m.Navigation.ActivePanel
	m.Navigation.Mode = state.ModeDetail
	m.Detail.Open(title, content)
}

// BackFromDetail closes the detail inspector and restores the previous panel.
func BackFromDetail(m *state.AppModel) {
	m.Navigation.Mode = state.ModeNormal
	if m.Navigation.PrevPanel != m.Navigation.ActivePanel {
		m.Navigation.ActivePanel = m.Navigation.PrevPanel
	}
	m.Detail.Close()
}

// ToHelp opens the help screen.
func ToHelp(m *state.AppModel) {
	m.Navigation.PrevPanel = m.Navigation.ActivePanel
	m.Navigation.Mode = state.ModeHelp
	m.Navigation.ActivePanel = state.PanelHelp
}

// BackFromHelp closes the help screen.
func BackFromHelp(m *state.AppModel) {
	m.Navigation.Mode = state.ModeNormal
	m.Navigation.ActivePanel = m.Navigation.PrevPanel
}

// ToFilter opens the search filter bar.
func ToFilter(m *state.AppModel) tea.Cmd {
	if m.Navigation.Mode == state.ModeLogView {
		ToSearch(m)
		return nil
	}
	filter.New(m).Open()
	return nil
}

// BackFromFilter clears the active filter and closes the filter bar.
func BackFromFilter(m *state.AppModel) {
	filter.New(m).Close()
}

// ToSearch opens log search without changing the underlying log data.
func ToSearch(m *state.AppModel) {
	m.Navigation.Mode = state.ModeSearch
	m.Navigation.SearchInput.Set(m.Log.LogSearchText)
}

// ToCommand opens the command palette.
func ToCommand(m *state.AppModel) {
	m.Navigation.Mode = state.ModeCommand
	m.Navigation.CommandInput.Reset()
}

// BackFromCommand closes the command palette.
func BackFromCommand(m *state.AppModel) {
	m.Navigation.Mode = state.ModeNormal
	m.Navigation.CommandInput.Reset()
}

// ToExec opens the exec shell dialog.
func ToExec(m *state.AppModel) {
	m.Dialog.Open(state.DialogSpec{Kind: state.DialogExec, Input: config.DefaultShell})
	m.Navigation.Mode = m.Dialog.Kind.Mode()
}

// BackFromExec closes the exec shell dialog.
func BackFromExec(m *state.AppModel) {
	m.Navigation.Mode = state.ModeNormal
	m.Dialog.Close()
}
