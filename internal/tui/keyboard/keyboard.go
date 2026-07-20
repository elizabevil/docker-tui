package keyboard

import (
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
		m.ErrorMessage = ""
		m.ErrorCount = 0
	}

	if m.Mode == state.ModeHelp {
		if action, known := resolveAction(key, m); known && (action == keys.ActionHelp || action == keys.ActionBack) {
			BackFromHelp(m)
		}
		return m, nil
	}

	if m.Mode == state.ModeExecPassthrough {
		if key == keys.KeyEsc {
			if m.ExecConn != nil {
				m.ExecConn.Close()
			}
			m.ExecConn = nil
			m.ExecID = ""
			m.ExecCh = nil
			m.ExecDone = nil
			if m.ExecBuf != nil {
				m.ExecBuf.Reset()
			}
			m.ExecScroll = 0
			m.Mode = state.ModeNormal
			return m, nil
		}
		if m.ExecConn != nil {
			m.ExecConn.Write(mapKeyToTerm(key))
		}
		return m, nil
	}

	if m.Mode == state.ModeFilter {
		return handleFilterInput(normalizeInputKey(rawKey), m), nil
	}

	if m.Mode == state.ModeCommand {
		return handleCommandInput(normalizeInputKey(rawKey), m)
	}

	if handleDetailKeys(key, m) {
		return m, nil
	}

	if m.Mode == state.ModeExport || m.Mode == state.ModeDebug {
		switch key {
		case keys.KeyTab:
			m.DialogFocus = 1 - m.DialogFocus // toggle 0↔1
		case keys.KeyEnter:
			if m.DialogFocus == 0 { // confirm
				ShowToastNow(m, "✓ "+m.DialogAction)
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
		m.HeaderVisible = !m.HeaderVisible
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
	case state.ModeFilter, state.ModeCommand:
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

func handleFilterInput(key string, m *state.AppModel) *state.AppModel {
	inLogSearch := m.Mode == state.ModeFilter && m.LogContainerID != "" && m.LogContent != nil
	clampFilterCursor(m)

	switch key {
	case keys.KeyEnter:
		if inLogSearch {
			m.LogSearchText = m.FilterText
			m.LogSearchMatch = 0
			scrollToMatch(m)
			m.Mode = state.ModeNormal
			m.FilterText = ""
			m.FilterCursor = 0
			return m
		}
		ApplyFilter(m)
		focusFirstFilteredItem(m)
		m.Mode = state.ModeNormal
		m.FilterCursor = 0
	case keys.KeyEsc:
		m.Mode = state.ModeNormal
		m.FilterText = ""
		m.FilterCursor = 0
		if inLogSearch {
			m.LogSearchText = ""
		} else {
			ApplyFilter(m)
		}
	default:
		if handled, changed := editTextInput(key, &m.FilterText, &m.FilterCursor); handled && changed && !inLogSearch {
			ApplyFilter(m)
		}
	}
	return m
}

func clampFilterCursor(m *state.AppModel) {
	if m == nil {
		return
	}
	max := len([]rune(m.FilterText))
	if m.FilterCursor < 0 {
		m.FilterCursor = 0
		return
	}
	if m.FilterCursor > max {
		m.FilterCursor = max
	}
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
