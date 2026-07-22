package keyboard

import (
	"fmt"
	"time"

	"github.com/elizabevil/docker-tui/internal/data/audit"
	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

func HandleKeyPress(msg tea.KeyPressMsg, m *state.AppModel) (*state.AppModel, tea.Cmd) {
	rawKey := msg.String()
	key := keys.Normalize(rawKey)
	var cmds []tea.Cmd

	if key != keys.KeyEsc {
		m.Navigation.EscPending = false
		m.Feedback.InfoMessage = ""
	}

	// Esc 也清除持久化错误
	if key == keys.KeyEsc && m.Feedback.ErrorMessage != "" {
		m.Feedback.ClearError()
	}

	if m.Navigation.Mode == state.ModeHelp {
		if action, known := resolveAction(key, m); known && (action == keys.ActionHelp || action == keys.ActionBack) {
			BackFromHelp(m)
		}
		return m, nil
	}

	if m.Navigation.Mode == state.ModeExecPassthrough {
		if key == keys.KeyEsc {
			FinishAudit(m, m.Exec.ExecAudit, audit.ResultCancelled, "Exec session closed by user", audit.Details{Shell: m.Exec.ExecShell})
			if m.Exec.ExecConn != nil {
				m.Exec.ExecConn.Close()
			}
			m.Exec.Reset()
			m.Navigation.Mode = state.ModeNormal
			return m, nil
		}
		if m.Exec.ExecConn != nil {
			m.Exec.ExecConn.Write(mapKeyToTerm(key))
		}
		return m, nil
	}

	if m.Navigation.Mode == state.ModeFilter {
		return handleFilterInput(normalizeInputKey(rawKey), m)
	}

	if m.Navigation.Mode == state.ModeSearch {
		return handleSearchInput(normalizeInputKey(rawKey), m), nil
	}
	if m.Navigation.Mode == state.ModeImagePull {
		return handleImagePullInput(normalizeInputKey(rawKey), m), nil
	}

	if m.Navigation.Mode == state.ModeCommand {
		return handleCommandInput(normalizeInputKey(rawKey), m)
	}
	if m.Navigation.Mode == state.ModeRuntimeSelect {
		return handleRuntimeSelectorKey(key, m)
	}
	if m.Navigation.Mode == state.ModeRename {
		return handleRenameDialogKey(normalizeInputKey(rawKey), m)
	}
	if m.Navigation.Mode == state.ModeResourceCreate {
		return handleResourceCreateKey(normalizeInputKey(rawKey), m)
	}
	if m.Navigation.Mode == state.ModeTop {
		return handleTopKey(key, m)
	}

	if handleDetailKeys(key, m) {
		return m, nil
	}

	if m.Dialog.Kind.IsSelection() {
		switch key {
		case keys.KeyTab:
			m.Dialog.MoveFocus(1, 2)
		case keys.KeyEnter:
			if m.Dialog.Focus == 0 {
				ShowToastNow(m, "✓ "+m.Dialog.Action)
			}
			clearDialogState(m)
		case keys.KeyEsc, keys.KeyN:
			clearDialogState(m)
		}
		return m, nil
	}

	if m.Navigation.Mode == state.ModeExec {
		return handleExecDialogKeys(key, m)
	}

	if handleLogKeys(key, m) {
		return m, nil
	}

	if m.Navigation.Mode == state.ModeConfirm {
		return handleConfirmKeys(key, m)
	}

	if m.Navigation.Mode == state.ModeMark {
		return handleMarkMode(key, m)
	}

	if keys.IsSpace(key) {
		if m.Navigation.Mode == state.ModeMark {
			cmds = append(cmds, RecordKeyStroke(m, key, keys.ActionLabelToggle))
			return doToggleMark(m)
		}
		cmds = append(cmds, RecordKeyStroke(m, key, keys.ActionLabelMark))
		return enterMarkMode(m)
	}

	// Nested views own additional cursor semantics and are resolved first.
	if m.Navigation.ActivePanel == state.PanelImages && m.Resources.Images.ContainersViewID != "" {
		if mm, cmd := handleImagePanelKeys(key, m); mm != nil || cmd != nil {
			return mm, cmd
		}
	}
	if m.Navigation.ActivePanel == state.PanelCompose && m.Compose.ComposeContainerViewID != "" {
		if mm, cmd := handleComposePanelKeys(key, m); mm != nil || cmd != nil {
			return mm, cmd
		}
	}
	if action, known := resolveAction(key, m); known {
		cmds = append(cmds, RecordKeyStroke(m, key, KeyStrokeActionLabel(key)))
		return handleAction(action, m, cmds)
	}

	if mm, cmd := handleImagePanelKeys(key, m); mm != nil || cmd != nil {
		return mm, cmd
	}
	if mm, cmd := handleComposePanelKeys(key, m); mm != nil || cmd != nil {
		return mm, cmd
	}

	if key == keys.KeyH {
		cmds = append(cmds, RecordKeyStroke(m, key, keys.ActionLabelHeader))
		m.Viewport.ToggleHeader()
		return m, tea.Batch(cmds...)
	}

	if key == keys.KeyC {
		cmds = append(cmds, RecordKeyStroke(m, key, keys.ActionLabelConn))
		return showConnectionInfo(m)
	}

	if key == keys.KeyO {
		if m.Navigation.Mode == state.ModeMark {
			return m, nil
		}
		switch m.Navigation.ActivePanel {
		case state.PanelContainers:
			cmds = append(cmds, RecordKeyStroke(m, key, keys.ActionLabelSort))
			return doContainerSort(m)
		case state.PanelImages:
			cmds = append(cmds, RecordKeyStroke(m, key, keys.ActionLabelSort))
			return doImageSort(m)
		case state.PanelNetworks:
			cmds = append(cmds, RecordKeyStroke(m, key, keys.ActionLabelSort))
			return doNetworkSort(m)
		}
	}

	if key == keys.KeyCtrlO {
		if m.Navigation.Mode == state.ModeMark {
			return m, nil
		}
		switch m.Navigation.ActivePanel {
		case state.PanelContainers:
			m.Resources.Containers.SortAsc = !m.Resources.Containers.SortAsc
			m.Resources.Containers.Cursor = 0
			m.Resources.Containers.ViewOffset = 0
			cmds = append(cmds, RecordKeyStroke(m, key, keys.ActionLabelSort))
			return m, tea.Batch(cmds...)
		case state.PanelImages:
			m.Resources.Images.SortAsc = !m.Resources.Images.SortAsc
			m.Resources.Images.Cursor = 0
			m.Resources.Images.ViewOffset = 0
			cmds = append(cmds, RecordKeyStroke(m, key, keys.ActionLabelSort))
			return m, tea.Batch(cmds...)
		case state.PanelNetworks:
			m.Resources.Networks.SortAsc = !m.Resources.Networks.SortAsc
			m.Resources.Networks.Cursor = 0
			m.Resources.Networks.ViewOffset = 0
			cmds = append(cmds, RecordKeyStroke(m, key, keys.ActionLabelSort))
			return m, tea.Batch(cmds...)
		}
	}

	return m, nil
}

func normalizeInputKey(key string) string {
	if len([]rune(key)) == 1 {
		return key
	}
	return keys.Normalize(key)
}

func resolveAction(key string, m *state.AppModel) (keys.KeyAction, bool) {
	if m == nil || m.Dependencies.Config == nil {
		return "", false
	}
	resolver := keys.NewResolver(keys.CompileBindings(m.Dependencies.Config.Keymap))
	return resolver.Resolve(key, keyContext(m))
}

func keyContext(m *state.AppModel) keys.Context {
	view := "containers"
	switch m.Navigation.ActivePanel {
	case state.PanelImages:
		if m.Resources.Images.ContainersViewID != "" {
			view = "image-containers"
		} else {
			view = "images"
		}
	case state.PanelVolumes:
		view = "volumes"
	case state.PanelNetworks:
		view = "networks"
	case state.PanelCompose:
		if m.Compose.ComposeContainerViewID != "" {
			view = "compose-containers"
		} else {
			view = "compose"
		}
	case state.PanelHelp:
		view = "help"
	}
	return keys.Context{App: "app", Surface: keySurface(m.Navigation.Mode), View: view, Mode: keyMode(m.Navigation.Mode)}
}

func keySurface(mode state.AppMode) string {
	switch mode {
	case state.ModeFilter, state.ModeSearch, state.ModeImagePull, state.ModeCommand:
		return "input"
	case state.ModeConfirm, state.ModeExport, state.ModeDebug, state.ModeExec, state.ModeExecShell, state.ModeRename, state.ModeResourceCreate:
		return "dialog"
	default:
		return "main"
	}
}

func keyMode(mode state.AppMode) string {
	switch mode {
	case state.ModeFilter:
		return "filter"
	case state.ModeSearch:
		return "search"
	case state.ModeImagePull:
		return "image-pull"
	case state.ModeCommand:
		return "command"
	case state.ModeLogView:
		return "logs"
	case state.ModeDetail:
		return "detail"
	case state.ModeMark:
		return "mark"
	case state.ModeTop:
		return "top"
	case state.ModeRename:
		return "rename"
	case state.ModeResourceCreate:
		return "resource-create"
	default:
		return "normal"
	}
}

func handleFilterInput(key string, m *state.AppModel) (*state.AppModel, tea.Cmd) {
	m.Navigation.FilterInput.Clamp()

	switch key {
	case keys.KeyEnter:
		ApplyFilter(m)
		focusFirstFilteredItem(m)
		m.Navigation.Mode = state.ModeNormal
		m.Navigation.ClearFilterExit()
	case keys.KeyEsc:
		if m.Navigation.FilterExitPending {
			BackFromFilter(m)
			return m, nil
		}
		token := m.Navigation.BeginFilterExit()
		ShowToastWarn(m, "Press Esc again within 5s to clear filter and exit")
		return m, tea.Tick(5*time.Second, func(time.Time) tea.Msg {
			return state.FilterExitTimeout{Token: token}
		})
	default:
		if handled, changed := editQueryInput(key, &m.Navigation.FilterInput); handled && changed {
			ApplyFilter(m)
		}
	}
	return m, nil
}

func handleSearchInput(key string, m *state.AppModel) *state.AppModel {
	m.Navigation.SearchInput.Clamp()
	switch key {
	case keys.KeyEnter:
		matches := m.Log.ApplySearch(m.Navigation.SearchInput.Text)
		m.Navigation.Mode = state.ModeLogView
		if matches == 0 && m.Log.LogSearchText != "" {
			ShowToastWarn(m, "No log matches for: "+m.Log.LogSearchText)
		} else if matches > 0 {
			ShowToastNow(m, fmt.Sprintf("Log match 1/%d", matches))
		}
	case keys.KeyEsc:
		m.Navigation.Mode = state.ModeLogView
		m.Navigation.SearchInput.Set(m.Log.LogSearchText)
	default:
		editQueryInput(key, &m.Navigation.SearchInput)
	}
	return m
}

func focusFirstFilteredItem(m *state.AppModel) {
	if m == nil {
		return
	}
	switch m.Navigation.ActivePanel {
	case state.PanelContainers:
		if m.Resources.Containers != nil {
			m.Resources.Containers.Cursor = 0
			m.Resources.Containers.ViewOffset = 0
		}
	case state.PanelImages:
		if m.Resources.Images != nil {
			m.Resources.Images.Cursor = 0
			m.Resources.Images.ViewOffset = 0
		}
	case state.PanelVolumes:
		if m.Resources.Volumes != nil {
			m.Resources.Volumes.Cursor = 0
			m.Resources.Volumes.ViewOffset = 0
		}
	case state.PanelNetworks:
		if m.Resources.Networks != nil {
			m.Resources.Networks.Cursor = 0
			m.Resources.Networks.ViewOffset = 0
		}
	case state.PanelCompose:
		if m.Compose.ComposeFocus == 1 {
			m.Compose.ComposeServiceCursor = 0
		} else {
			m.Compose.ComposeCursor = 0
			m.Compose.ComposeServiceCursor = 0
		}
	}
}
