package keyboard

import (
	"github.com/elizabevil/docker-tui/internal/tui/keys"

	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

func HandleKeyPress(msg tea.KeyPressMsg, m *state.AppModel) (*state.AppModel, tea.Cmd) {
	key := msg.String()
	// 不区分大小写：统一转为小写匹配
	if len(key) == 1 && key >= "A" && key <= "Z" {
		key = string(key[0] + 32)
	}
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
		key = keys.KeyEsc
		BackFromHelp(m)
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
		return handleFilterInput(key, m), nil
	}

	if m.Mode == state.ModeCommand {
		return handleCommandInput(key, m)
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

	if mm, cmd := handleImagePanelKeys(key, m); mm != nil || cmd != nil {
		return mm, cmd
	}
	if mm, cmd := handleComposePanelKeys(key, m); mm != nil || cmd != nil {
		return mm, cmd
	}

	if key == keys.KeyCtrlD {
		cmds = append(cmds, RecordKeyStroke(m, key, keys.ActionLabelDelete))
		return doDeleteAction(m)
	}

	if key == keys.KeyF2 || key == "F2" {
		cmds = append(cmds, RecordKeyStroke(m, key, keys.ActionLabelSwitch))
		return doSwitchRuntime(m)
	}

	if key == keys.KeyH {
		cmds = append(cmds, RecordKeyStroke(m, key, keys.ActionLabelHeader))
		m.HeaderVisible = !m.HeaderVisible
		return m, tea.Batch(cmds...)
	}

	if mm, cmd := handleVolumeEnter(key, m); mm != nil || cmd != nil {
		return mm, cmd
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

	km := keys.DefaultKeyMapping()
	action, known := km[key]
	if !known {
		return m, nil
	}
	cmds = append(cmds, RecordKeyStroke(m, key, KeyStrokeActionLabel(key)))
	return handleAction(action, m, cmds)
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
	case keys.KeyLeft:
		if m.FilterCursor > 0 {
			m.FilterCursor--
		}
	case keys.KeyRight:
		t := []rune(m.FilterText)
		if m.FilterCursor < len(t) {
			m.FilterCursor++
		}
	case "ctrl+a", keys.KeyHome:
		m.FilterCursor = 0
	case "ctrl+e", keys.KeyEnd:
		m.FilterCursor = len([]rune(m.FilterText))
	case keys.KeyBackspace:
		t := []rune(m.FilterText)
		if len(t) > 0 && m.FilterCursor > 0 {
			idx := m.FilterCursor
			m.FilterText = string(append(t[:idx-1], t[idx:]...))
			m.FilterCursor--
			if !inLogSearch {
				ApplyFilter(m)
			}
		}
	case keys.KeyDelete:
		t := []rune(m.FilterText)
		if len(t) > 0 && m.FilterCursor < len(t) {
			idx := m.FilterCursor
			m.FilterText = string(append(t[:idx], t[idx+1:]...))
			if !inLogSearch {
				ApplyFilter(m)
			}
		}
	default:
		if len(key) == 1 && key != keys.KeyBackspace {
			t := []rune(m.FilterText)
			idx := m.FilterCursor
			if idx < 0 {
				idx = 0
			}
			if idx > len(t) {
				idx = len(t)
			}
			insert := []rune(key)
			m.FilterText = string(append(append(t[:idx], insert...), t[idx:]...))
			m.FilterCursor += len(insert)
			if !inLogSearch {
				ApplyFilter(m)
			}
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
