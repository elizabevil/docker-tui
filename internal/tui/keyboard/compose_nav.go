package keyboard

import (
	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

func handleComposePanelKeys(key string, m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Navigation.ActivePanel != state.PanelCompose {
		return nil, nil
	}

	// 容器子视图模式：Esc 返回
	if m.Compose.ComposeContainerViewID != "" {
		action, known := resolveAction(key, m)
		if action == keys.ActionBack || key == keys.KeyLeft {
			m.Compose.ComposeContainerViewID = ""
			m.Compose.ComposeContainerCursor = 0
			return m, RecordKeyStroke(m, key, keys.ActionLabelProjects)
		}
		if known && action == keys.ActionDown {
			m.Compose.ComposeContainerCursor++
			return m, nil
		}
		if known && action == keys.ActionUp {
			if m.Compose.ComposeContainerCursor > 0 {
				m.Compose.ComposeContainerCursor--
			}
			return m, nil
		}
		return nil, nil
	}

	// ← 切到左栏(项目)
	if key == keys.KeyLeft {
		m.Compose.ComposeFocus = 0
		return m, RecordKeyStroke(m, key, keys.ActionLabelProjects)
	}
	// → 切到右栏(服务)
	if key == keys.KeyRight {
		m.Compose.ComposeFocus = 1
		if m.Compose.ComposeServiceCursor < 0 {
			m.Compose.ComposeServiceCursor = 0
		}
		return m, RecordKeyStroke(m, key, keys.ActionLabelServices)
	}
	// Enter:
	//   左栏 → 切到右栏（同 →）
	//   右栏 → 进入容器子视图（Compose 面板内）
	switch key {
	case keys.KeyD:
		// 左栏按 d → 项目概览
		if m.Compose.ComposeFocus == 0 {
			project := currentComposeProject(m)
			if project != "" {
				ToDetail(m, "Compose Detail: "+project, "")
				m.Detail.DetailResourceType = state.ResourceComposeProject
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
	if m.Compose.ComposeFocus == 0 {
		m.Compose.ComposeFocus = 1
		if m.Compose.ComposeServiceCursor < 0 {
			m.Compose.ComposeServiceCursor = 0
		}
		return m, RecordKeyStroke(m, keys.KeyEnter, keys.ActionLabelServices)
	}
	project := currentComposeProject(m)
	services := composeServiceNames(m, project)
	if m.Compose.ComposeServiceCursor >= 0 && m.Compose.ComposeServiceCursor < len(services) {
		m.Compose.ComposeContainerViewID = services[m.Compose.ComposeServiceCursor]
		m.Compose.ComposeContainerCursor = 0
	}
	return m, RecordKeyStroke(m, keys.KeyEnter, keys.ActionLabelContainers)
}
