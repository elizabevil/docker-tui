package keyboard

import (
	"sort"
	"strings"

	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

func switchPanel(m *state.AppModel, direction int) {
	panels := state.PanelList()
	current := -1
	for i, p := range panels {
		if p == m.ActivePanel {
			current = i
			break
		}
	}
	if current == -1 {
		current = 0
	}
	next := (current + direction + len(panels)) % len(panels)
	m.ActivePanel = panels[next]
	m.StatsActive = false
}

func moveCursor(m *state.AppModel, delta int) {
	var items int
	switch m.ActivePanel {
	case state.PanelContainers:
		items = m.Containers.Len()
		m.Containers.Cursor = clamp(m.Containers.Cursor+delta, items)
	case state.PanelImages:
		if m.Images.ContainersViewID != "" {
			m.Images.ContainerCursor += delta
			if m.Images.ContainerCursor < 0 {
				m.Images.ContainerCursor = 0
			}
		} else {
			items = m.Images.Len()
			m.Images.Cursor = clamp(m.Images.Cursor+delta, items)
		}
	case state.PanelVolumes:
		items = m.Volumes.Len()
		m.Volumes.Cursor = clamp(m.Volumes.Cursor+delta, items)
	case state.PanelNetworks:
		items = m.Networks.Len()
		m.Networks.Cursor = clamp(m.Networks.Cursor+delta, items)
	case state.PanelCompose:
		if m.ComposeFocus == 1 {
			// 右栏：服务列表
			names := composeServiceNames(m, currentComposeProject(m))
			items = len(names)
			m.ComposeServiceCursor = clamp(m.ComposeServiceCursor+delta, items)
		} else {
			// 左栏：项目列表
			names := composeProjectNames(m)
			items = len(names)
			m.ComposeCursor = clamp(m.ComposeCursor+delta, items)
		}
	}
	m.StatsActive = false
}

// ApplyFilter applies the current resource filter to the active panel.
func ApplyFilter(m *state.AppModel) {
	if m == nil {
		return
	}
	if m.ActivePanel == state.PanelCompose {
		if m.ComposeFocus == 1 {
			m.ComposeServiceFilter = m.FilterInput.Text
		} else {
			m.ComposeProjectFilter = m.FilterInput.Text
		}
		return
	}
	if f := activeTableFilter(m); f != nil {
		f.SetFilter(m.FilterInput.Text)
	}
}

func activeTableFilter(m *state.AppModel) state.TableFilter {
	if m == nil {
		return nil
	}
	switch m.ActivePanel {
	case state.PanelContainers:
		return m.Containers
	case state.PanelImages:
		return m.Images
	case state.PanelVolumes:
		return m.Volumes
	case state.PanelNetworks:
		return m.Networks
	default:
		return nil
	}
}

func PanelFromMode(mode state.AppMode) state.PanelType {
	switch mode {
	case state.ModeHelp:
		return state.PanelHelp
	case state.ModeLogView:
		return state.PanelLogs
	case state.ModeDetail:
		return state.PanelDetail
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
	m.Mode = state.ModeNormal
	m.DialogTitle = ""
	m.DialogBody = ""
	m.DialogPreview = ""
	m.DialogAction = ""
	m.DialogFocus = 0
	m.DialogCursor = 0
}

func composeProjectNames(m *state.AppModel) []string {
	filter := ""
	if m != nil {
		filter = m.ComposeProjectFilter
	}
	set := make(map[string]bool)
	for _, c := range m.Containers.Items {
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
		filter = m.ComposeServiceFilter
	}
	set := make(map[string]bool)
	for _, c := range m.Containers.Items {
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
	if m.ComposeCursor >= len(names) {
		m.ComposeCursor = len(names) - 1
	}
	if m.ComposeCursor < 0 {
		m.ComposeCursor = 0
	}
	return names[m.ComposeCursor]
}

// FormatKeyForDisplay converts a raw key string to a display-friendly label.
func FormatKeyForDisplay(key string) string {
	switch {
	case key == " ":
		return "␣"
	case key == "enter":
		return "↵"
	case key == "esc":
		return "Esc"
	case key == "tab":
		return "⇥"
	case key == "backspace":
		return "⌫"
	case key == "up":
		return "↑"
	case key == "down":
		return "↓"
	case key == "left":
		return "←"
	case key == "right":
		return "→"
	case key == "pgup":
		return "Pg↑"
	case key == "pgdown":
		return "Pg↓"
	case strings.HasPrefix(key, "ctrl+"):
		return "^" + strings.ToUpper(strings.TrimPrefix(key, "ctrl+"))
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
	if len(m.KeyStrokeBuffer) >= 3 {
		m.KeyStrokeBuffer = m.KeyStrokeBuffer[1:]
	}
	m.KeyStrokeBuffer = append(m.KeyStrokeBuffer, evt)
	// 重置计时器：新按键直接显示，旧内容立即消失
	dur := m.KeyStrokeDuration
	if dur <= 0 {
		dur = 30
	}
	m.KeyStrokeTimer = dur
	m.KeyStrokeAnim = 0
	m.KeyStrokeDispTimer = 0
	m.LastKeyStroke = nil // 旧内容立即销毁
	return func() tea.Msg { return state.KeyStrokeTick{} }
}

// KeyStrokeActionLabel returns a short action label for a keys.
func KeyStrokeActionLabel(key string) string {
	label := map[string]string{
		"q": "Quit", "Q": "Quit",
		"?": "Help", "f1": "Help",
		"/": "Filter",
		"r": "Refresh", "R": "Restart",
		"s": "Start", "S": "Stop", "K": "Kill",
		"d": "Detail", "D": "Detail",
		"l": "Logs", "L": "Logs",
		"i": "Inspect", "I": "Inspect",
		"e": "Exec", "E": "Exec",
		"m": "Stats", "M": "Stats",
		"P": "Pull", "p": "Prune",
		"h": "Header", "H": "Header",
		"c": "Conn", "C": "Conn",
		"o": "Sort", "O": "Sort",
		"enter": "Open",
		"esc":   "Back",
		" ":     "Mark",
		"tab":   "Panel",
		"up":    "Up", "k": "Up",
		"down": "Down", "j": "Down",
		"ctrl+d": "Delete",
		"f2":     "Switch",
		":":      "Cmd",
	}
	if l, ok := label[key]; ok {
		return l
	}
	return key
}

// confirmAction sets the model to ModeConfirm with the given action, target, and message.
func confirmAction(m *state.AppModel, action, target, message string) {
	m.ConfirmAction = action
	m.ConfirmTarget = target
	m.ConfirmMessage = message
	m.Mode = state.ModeConfirm
}
