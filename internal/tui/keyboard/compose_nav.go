package keyboard

import (
	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/pages/compose"

	tea "charm.land/bubbletea/v2"
)

func handleComposePanelKeys(key string, m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.ActivePanel != state.PanelCompose {
		return nil, nil
	}

	// 容器子视图模式：Esc 返回
	if m.ComposeContainerViewID != "" {
		action, known := resolveAction(key, m)
		if action == keys.ActionBack || key == keys.KeyLeft {
			m.ComposeContainerViewID = ""
			m.ComposeContainerCursor = 0
			return m, RecordKeyStroke(m, key, keys.ActionLabelProjects)
		}
		if known && action == keys.ActionDown {
			m.ComposeContainerCursor++
			return m, nil
		}
		if known && action == keys.ActionUp {
			if m.ComposeContainerCursor > 0 {
				m.ComposeContainerCursor--
			}
			return m, nil
		}
		return nil, nil
	}

	// ← 切到左栏(项目)
	if key == keys.KeyLeft {
		m.ComposeFocus = 0
		return m, RecordKeyStroke(m, key, keys.ActionLabelProjects)
	}
	// → 切到右栏(服务)
	if key == keys.KeyRight {
		m.ComposeFocus = 1
		if m.ComposeServiceCursor < 0 {
			m.ComposeServiceCursor = 0
		}
		return m, RecordKeyStroke(m, key, keys.ActionLabelServices)
	}
	// Enter:
	//   左栏 → 切到右栏（同 →）
	//   右栏 → 进入容器子视图（Compose 面板内）
	switch key {
	case keys.KeyD:
		// 左栏按 d → 项目概览
		if m.ComposeFocus == 0 {
			content := compose.RenderProjectDetail(m)
			if content != "" {
				ToDetail(m, "Project: "+currentComposeProject(m), content)
			}
			return m, RecordKeyStroke(m, key, keys.ActionLabelDetail)
		}
		return nil, nil
	case keys.KeyS:
		mm, cmd := doComposeStart(m)
		return mm, tea.Batch(cmd, RecordKeyStroke(m, key, keys.ActionLabelStart))
	case keys.KeyCtrlS:
		mm, cmd := doComposeStop(m)
		return mm, tea.Batch(cmd, RecordKeyStroke(m, key, keys.ActionLabelStop))
	case keys.KeyL:
		mm, cmd := doComposeLogs(m)
		return mm, tea.Batch(cmd, RecordKeyStroke(m, key, keys.ActionLabelLogs))
	case keys.KeyCtrlD:
		mm, cmd := doComposeDown(m)
		return mm, tea.Batch(cmd, RecordKeyStroke(m, key, keys.ActionLabelDown))
	}
	return nil, nil
}

func doComposeEnter(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.ComposeFocus == 0 {
		m.ComposeFocus = 1
		if m.ComposeServiceCursor < 0 {
			m.ComposeServiceCursor = 0
		}
		return m, RecordKeyStroke(m, keys.KeyEnter, keys.ActionLabelServices)
	}
	project := currentComposeProject(m)
	services := composeServiceNames(m, project)
	if m.ComposeServiceCursor >= 0 && m.ComposeServiceCursor < len(services) {
		m.ComposeContainerViewID = services[m.ComposeServiceCursor]
		m.ComposeContainerCursor = 0
	}
	return m, RecordKeyStroke(m, keys.KeyEnter, keys.ActionLabelContainers)
}
