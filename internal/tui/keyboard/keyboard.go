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
		m.EscPending = false
		m.InfoMessage = ""
	}

	// Esc 也清除持久化错误
	if key == keys.KeyEsc && m.ErrorMessage != "" {
		m.FeedbackState.ClearError()
	}

	if m.Mode == state.ModeHelp {
		if action, known := resolveAction(key, m); known && (action == keys.ActionHelp || action == keys.ActionBack) {
			BackFromHelp(m)
		}
		return m, nil
	}

	if m.Mode == state.ModeExecPassthrough {
		if key == keys.KeyEsc {
			FinishAudit(m, m.ExecAudit, audit.ResultCancelled, "Exec session closed by user", audit.Details{Shell: m.ExecShell})
			if m.ExecConn != nil {
				m.ExecConn.Close()
			}
			m.ExecState.Reset()
			m.Mode = state.ModeNormal
			return m, nil
		}
		if m.ExecConn != nil {
			m.ExecConn.Write(mapKeyToTerm(key))
		}
		return m, nil
	}

	if m.Mode == state.ModeFilter {
		return handleFilterInput(normalizeInputKey(rawKey), m)
	}

	if m.Mode == state.ModeSearch {
		return handleSearchInput(normalizeInputKey(rawKey), m), nil
	}
	if m.Mode == state.ModeImagePull {
		return handleImagePullInput(normalizeInputKey(rawKey), m), nil
	}

	if m.Mode == state.ModeCommand {
		return handleCommandInput(normalizeInputKey(rawKey), m)
	}
	if m.Mode == state.ModeRuntimeSelect {
		return handleRuntimeSelectorKey(key, m)
	}

	if handleDetailKeys(key, m) {
		return m, nil
	}

	if m.DialogState.Kind.IsSelection() {
		switch key {
		case keys.KeyTab:
			m.DialogState.MoveFocus(1, 2)
		case keys.KeyEnter:
			if m.DialogState.Focus == 0 {
				ShowToastNow(m, "✓ "+m.DialogState.Action)
			}
			clearDialogState(m)
		case keys.KeyEsc, keys.KeyN:
			clearDialogState(m)
		}
		return m, nil
	}

	if m.Mode == state.ModeExec {
		return handleExecDialogKeys(key, m)
	}

	if handleLogKeys(key, m) {
		return m, nil
	}

	if m.Mode == state.ModeConfirm {
		return handleConfirmKeys(key, m)
	}

	if m.Mode == state.ModeMark {
		return handleMarkMode(key, m)
	}

	if keys.IsSpace(key) {
		if m.Mode == state.ModeMark {
			cmds = append(cmds, RecordKeyStroke(m, key, keys.ActionLabelToggle))
			return doToggleMark(m)
		}
		cmds = append(cmds, RecordKeyStroke(m, key, keys.ActionLabelMark))
		return enterMarkMode(m)
	}

	// Nested views own additional cursor semantics and are resolved first.
	if m.ActivePanel == state.PanelImages && m.Images.ContainersViewID != "" {
		if mm, cmd := handleImagePanelKeys(key, m); mm != nil || cmd != nil {
			return mm, cmd
		}
	}
	if m.ActivePanel == state.PanelCompose && m.ComposeContainerViewID != "" {
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
		m.ViewportState.ToggleHeader()
		return m, tea.Batch(cmds...)
	}

	if key == keys.KeyC {
		cmds = append(cmds, RecordKeyStroke(m, key, keys.ActionLabelConn))
		return showConnectionInfo(m)
	}

	if key == keys.KeyO {
		if m.Mode == state.ModeMark {
			return m, nil
		}
		switch m.ActivePanel {
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
		if m.Mode == state.ModeMark {
			return m, nil
		}
		switch m.ActivePanel {
		case state.PanelContainers:
			m.Containers.SortAsc = !m.Containers.SortAsc
			m.Containers.Cursor = 0
			m.Containers.ViewOffset = 0
			cmds = append(cmds, RecordKeyStroke(m, key, keys.ActionLabelSort))
			return m, tea.Batch(cmds...)
		case state.PanelImages:
			m.Images.SortAsc = !m.Images.SortAsc
			m.Images.Cursor = 0
			m.Images.ViewOffset = 0
			cmds = append(cmds, RecordKeyStroke(m, key, keys.ActionLabelSort))
			return m, tea.Batch(cmds...)
		case state.PanelNetworks:
			m.Networks.SortAsc = !m.Networks.SortAsc
			m.Networks.Cursor = 0
			m.Networks.ViewOffset = 0
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
	if m == nil || m.Config == nil {
		return "", false
	}
	resolver := keys.NewResolver(keys.CompileBindings(m.Config.Keymap))
	return resolver.Resolve(key, keyContext(m))
}

func keyContext(m *state.AppModel) keys.Context {
	view := "containers"
	switch m.ActivePanel {
	case state.PanelImages:
		if m.Images.ContainersViewID != "" {
			view = "image-containers"
		} else {
			view = "images"
		}
	case state.PanelVolumes:
		view = "volumes"
	case state.PanelNetworks:
		view = "networks"
	case state.PanelCompose:
		if m.ComposeContainerViewID != "" {
			view = "compose-containers"
		} else {
			view = "compose"
		}
	case state.PanelHelp:
		view = "help"
	}
	return keys.Context{App: "app", Surface: keySurface(m.Mode), View: view, Mode: keyMode(m.Mode)}
}

func keySurface(mode state.AppMode) string {
	switch mode {
	case state.ModeFilter, state.ModeSearch, state.ModeImagePull, state.ModeCommand:
		return "input"
	case state.ModeConfirm, state.ModeExport, state.ModeDebug, state.ModeExec, state.ModeExecShell:
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
	default:
		return "normal"
	}
}

func handleFilterInput(key string, m *state.AppModel) (*state.AppModel, tea.Cmd) {
	m.FilterInput.Clamp()

	switch key {
	case keys.KeyEnter:
		ApplyFilter(m)
		focusFirstFilteredItem(m)
		m.Mode = state.ModeNormal
		m.NavigationState.ClearFilterExit()
	case keys.KeyEsc:
		if m.FilterExitPending {
			BackFromFilter(m)
			return m, nil
		}
		token := m.NavigationState.BeginFilterExit()
		ShowToastWarn(m, "Press Esc again within 5s to clear filter and exit")
		return m, tea.Tick(5*time.Second, func(time.Time) tea.Msg {
			return state.FilterExitTimeout{Token: token}
		})
	default:
		if handled, changed := editQueryInput(key, &m.FilterInput); handled && changed {
			ApplyFilter(m)
		}
	}
	return m, nil
}

func handleSearchInput(key string, m *state.AppModel) *state.AppModel {
	m.SearchInput.Clamp()
	switch key {
	case keys.KeyEnter:
		matches := m.LogState.ApplySearch(m.SearchInput.Text)
		m.Mode = state.ModeLogView
		if matches == 0 && m.LogSearchText != "" {
			ShowToastWarn(m, "No log matches for: "+m.LogSearchText)
		} else if matches > 0 {
			ShowToastNow(m, fmt.Sprintf("Log match 1/%d", matches))
		}
	case keys.KeyEsc:
		m.Mode = state.ModeLogView
		m.SearchInput.Set(m.LogSearchText)
	default:
		editQueryInput(key, &m.SearchInput)
	}
	return m
}

func focusFirstFilteredItem(m *state.AppModel) {
	if m == nil {
		return
	}
	switch m.ActivePanel {
	case state.PanelContainers:
		if m.Containers != nil {
			m.Containers.Cursor = 0
			m.Containers.ViewOffset = 0
		}
	case state.PanelImages:
		if m.Images != nil {
			m.Images.Cursor = 0
			m.Images.ViewOffset = 0
		}
	case state.PanelVolumes:
		if m.Volumes != nil {
			m.Volumes.Cursor = 0
			m.Volumes.ViewOffset = 0
		}
	case state.PanelNetworks:
		if m.Networks != nil {
			m.Networks.Cursor = 0
			m.Networks.ViewOffset = 0
		}
	case state.PanelCompose:
		if m.ComposeFocus == 1 {
			m.ComposeServiceCursor = 0
		} else {
			m.ComposeCursor = 0
			m.ComposeServiceCursor = 0
		}
	}
}
