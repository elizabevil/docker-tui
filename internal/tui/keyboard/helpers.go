package keyboard

import (
	"sort"
	"strings"

	"github.com/elizabevil/docker-tui/internal/data/audit"
	"github.com/elizabevil/docker-tui/internal/tui/filter"
	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

func switchPanel(m *state.AppModel, direction int) {
	panels := state.PanelList()
	current := -1
	for i, p := range panels {
		if p == m.Navigation.ActivePanel {
			current = i
			break
		}
	}
	if current == -1 {
		current = 0
	}
	next := (current + direction + len(panels)) % len(panels)
	m.Navigation.ActivePanel = panels[next]
	m.Metrics.StatsActive = false
	// 切换页面清除已有选择:不同页面(如 compose 与容器)的 marks 会相互
	// 影响,且 mark 的 id 仅在所属页面上下文有意义,切页后必须失效。
	m.Selection.ClearMarks()
}

func moveCursor(m *state.AppModel, delta int) {
	var items int
	switch m.Navigation.ActivePanel {
	case state.PanelContainers:
		items = len(m.Resources.Containers.FilteredItems())
		m.Resources.Containers.Cursor = clamp(m.Resources.Containers.Cursor+delta, items)
	case state.PanelImages:
		if m.Resources.Images.ContainersViewID != "" {
			m.Resources.Images.ContainerCursor += delta
			if m.Resources.Images.ContainerCursor < 0 {
				m.Resources.Images.ContainerCursor = 0
			}
		} else {
			items = len(m.Resources.Images.FilteredItems())
			m.Resources.Images.Cursor = clamp(m.Resources.Images.Cursor+delta, items)
		}
	case state.PanelVolumes:
		items = len(m.Resources.Volumes.FilteredItems())
		m.Resources.Volumes.Cursor = clamp(m.Resources.Volumes.Cursor+delta, items)
	case state.PanelNetworks:
		items = len(m.Resources.Networks.FilteredItems())
		m.Resources.Networks.Cursor = clamp(m.Resources.Networks.Cursor+delta, items)
	case state.PanelCompose:
		if m.Compose.ComposeFocus == 1 {
			// 右栏：服务列表
			names := composeServiceNames(m, currentComposeProject(m))
			items = len(names)
			m.Compose.ComposeServiceCursor = clamp(m.Compose.ComposeServiceCursor+delta, items)
		} else {
			// 左栏：项目列表
			names := composeProjectNames(m)
			items = len(names)
			m.Compose.ComposeCursor = clamp(m.Compose.ComposeCursor+delta, items)
		}
	case state.PanelAudit:
		items = m.Audit.Total()
		m.Audit.Cursor = clamp(m.Audit.Cursor+delta, items)
	}
	m.Metrics.StatsActive = false
}

// ApplyFilter applies the current resource filter to the active panel.
// Thin wrapper around filter.Controller.ApplyCurrent for backward
// compatibility with the existing call sites in keyboard and tests.
func ApplyFilter(m *state.AppModel) {
	filter.New(m).ApplyCurrent()
}
func PanelFromMode(mode state.AppMode) state.PanelType {
	switch mode {
	case state.ModeHelp:
		return state.PanelHelp
	case state.ModeLogView:
		return state.PanelLogs
	case state.ModeDetail:
		return state.PanelDetail
	case state.ModeAuditDetail:
		return state.PanelAudit
	default:
		return state.PanelContainers
	}
}

func clamp(val, max int) int {
	if val < 0 {
		return 0
	}
	if max == 0 {
		return 0
	}
	if val >= max {
		return max - 1
	}
	return val
}

func clearDialogState(m *state.AppModel) {
	m.Navigation.Mode = state.ModeNormal
	m.Dialog.Close()
}

func composeProjectNames(m *state.AppModel) []string {
	filter := ""
	if m != nil {
		filter = m.Compose.ComposeProjectFilter
	}
	set := make(map[string]bool)
	for _, c := range m.Resources.Containers.Items {
		if c.ComposeProject == "" {
			continue
		}
		if filter != "" && !strings.Contains(c.ComposeProject, filter) {
			continue
		}
		set[c.ComposeProject] = true
	}
	names := make([]string, 0, len(set))
	for name := range set {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func composeServiceNames(m *state.AppModel, project string) []string {
	filter := ""
	if m != nil {
		filter = m.Compose.ComposeServiceFilter
	}
	set := make(map[string]bool)
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
		set[name] = true
	}
	names := make([]string, 0, len(set))
	for name := range set {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func currentComposeProject(m *state.AppModel) string {
	names := composeProjectNames(m)
	if len(names) == 0 {
		return ""
	}
	if m.Compose.ComposeCursor >= len(names) {
		m.Compose.ComposeCursor = len(names) - 1
	}
	if m.Compose.ComposeCursor < 0 {
		m.Compose.ComposeCursor = 0
	}
	return names[m.Compose.ComposeCursor]
}

// FormatKeyForDisplay converts a raw key string to a display-friendly label.
func FormatKeyForDisplay(key string) string {
	switch {
	case key == " ":
		return "␣"
	case key == keys.KeyEnter:
		return "↵"
	case key == keys.KeyEsc:
		return keys.KEsc
	case key == keys.KeyTab:
		return "⇥"
	case key == keys.KeyBackspace:
		return "⌫"
	case key == keys.KeyUp:
		return keys.KUp
	case key == keys.KeyDown:
		return keys.KDown
	case key == keys.KeyLeft:
		return keys.KLeft
	case key == keys.KeyRight:
		return keys.KRight
	case key == keys.KeyPgUp:
		return keys.KPgUp + keys.KUp
	case key == keys.KeyPgDn:
		return keys.KPgDn + keys.KDown
	case strings.HasPrefix(key, keys.KeyCtrlPrefix):
		return keys.DisplayCtrl + strings.ToUpper(strings.TrimPrefix(key, keys.KeyCtrlPrefix))
	case len(key) == 1:
		return strings.ToUpper(key)
	default:
		return key
	}
}

// RecordKeyStroke appends a key event to the buffer (max 3, FIFO).
// 新按键直接替换旧显示，重置 3s 计时器。
func RecordKeyStroke(m *state.AppModel, key, action string) tea.Cmd {
	evt := state.KeyStrokeEvent{Key: FormatKeyForDisplay(key), Action: action}
	// FIFO: keep max 3 events
	if len(m.Feedback.KeyStrokeBuffer) >= 3 {
		m.Feedback.KeyStrokeBuffer = m.Feedback.KeyStrokeBuffer[1:]
	}
	m.Feedback.KeyStrokeBuffer = append(m.Feedback.KeyStrokeBuffer, evt)
	// 重置计时器：新按键直接显示，旧内容立即消失
	dur := m.Feedback.KeyStrokeDuration
	if dur <= 0 {
		dur = 30
	}
	m.Feedback.KeyStrokeTimer = dur
	m.Feedback.KeyStrokeAnim = 0
	m.Feedback.KeyStrokeDispTimer = 0
	m.Feedback.LastKeyStroke = nil // 旧内容立即销毁
	return func() tea.Msg { return state.KeyStrokeTick{} }
}

// KeyStrokeActionLabel returns a short action label for a keys.
func KeyStrokeActionLabel(key string) string {
	label := map[string]string{
		keys.KeyQ: keys.ActionLabelQuit, keys.KeyQUpper: keys.ActionLabelQuit,
		keys.KeyQmark: keys.ActionLabelHelp, keys.KeyF1: keys.ActionLabelHelp,
		keys.KeySlash: keys.ActionLabelFilter,
		keys.KeyR:     keys.ActionLabelRefresh, keys.KeyRUpper: keys.ActionLabelRestart,
		keys.KeyS: keys.ActionLabelStart, keys.KeySUpper: keys.ActionLabelStop, keys.KeyKUpper: keys.ActionLabelKill,
		keys.KeyD: keys.ActionLabelDetail, keys.KeyDUpper: keys.ActionLabelDetail,
		keys.KeyL: keys.ActionLabelLogs, keys.KeyLUpper: keys.ActionLabelLogs,
		keys.KeyI: keys.ActionLabelInspect, keys.KeyIUpper: keys.ActionLabelInspect,
		keys.KeyE: keys.ActionLabelExec, keys.KeyEUpper: keys.ActionLabelExec,
		keys.KeyM: keys.ActionLabelStats, keys.KeyMUpper: keys.ActionLabelStats,
		keys.KeyPUpper: keys.ActionLabelPull, keys.KeyP: keys.ActionLabelPrune,
		keys.KeyH: keys.ActionLabelHistory, keys.KeyHUpper: keys.ActionLabelHistory,
		keys.KeyC: keys.ActionLabelConn, keys.KeyCUpper: keys.ActionLabelConn,
		keys.KeyO: keys.ActionLabelSort, keys.KeyOUpper: keys.ActionLabelSort,
		keys.KeyEnter: keys.ActionLabelOpen,
		keys.KeyEsc:   keys.ActionLabelBack,
		keys.KeySpace: keys.ActionLabelMark,
		keys.KeyTab:   keys.ActionLabelPanel,
		keys.KeyUp:    keys.ActionLabelUp, keys.KeyK: keys.ActionLabelUp,
		keys.KeyDown: keys.ActionLabelDown, keys.KeyJ: keys.ActionLabelDown,
		keys.KeyCtrlD: keys.ActionLabelDelete,
		keys.KeyF2:    keys.ActionLabelSwitch,
		":":           keys.ActionLabelCmd,
	}
	if l, ok := label[key]; ok {
		return l
	}
	return key
}

// confirmAction sets the model to ModeConfirm with the given action, target, and message.
func confirmAction(m *state.AppModel, action, target, message string) {
	m.Confirm.Open(action, target, message, audit.Trace{})
	m.Navigation.Mode = state.ModeConfirm
}
