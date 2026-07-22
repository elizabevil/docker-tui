package keyboard

import (
	"fmt"

	"github.com/elizabevil/docker-tui/internal/data/audit"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

func beginAudit(m *state.AppModel, action string, target audit.Target, message string) audit.Trace {
	if m == nil || m.Dependencies.Audit == nil {
		return audit.Trace{}
	}
	trace := m.Dependencies.Audit.Begin(action, target, audit.RuntimeContext{
		Type: m.Connection.RuntimeType,
		Name: runtimeName(m),
		Host: runtimeHost(m),
	}, audit.UIContext{
		Surface: "main",
		View:    panelName(m.Navigation.ActivePanel),
		Mode:    modeLabel(m.Navigation.Mode),
	}, message)
	syncAuditProjection(m)
	return trace
}

func runtimeName(m *state.AppModel) string {
	if m != nil && m.Connection.Pool != nil {
		return m.Connection.Pool.ActiveName()
	}
	return ""
}

func FinishAudit(m *state.AppModel, trace audit.Trace, result audit.Result, message string, details audit.Details) {
	if m == nil || m.Dependencies.Audit == nil || !trace.Valid() {
		return
	}
	m.Dependencies.Audit.Finish(trace, result, message, details)
	syncAuditProjection(m)
}

func publishUIMessage(m *state.AppModel, level audit.Level, message string) {
	if m != nil && m.Dependencies.Audit != nil {
		m.Dependencies.Audit.PublishUI(level, message)
		syncAuditProjection(m)
	}
}

func syncAuditProjection(m *state.AppModel) {
	if m == nil || m.Dependencies.Audit == nil {
		return
	}
	if notification := m.Dependencies.Audit.ConsumeNotification(); notification != nil {
		m.Feedback.ShowToast(notification.Message, toastLevel(notification.Level), 30)
	}
	if operation := m.Dependencies.Audit.CurrentOperation(); operation != nil {
		m.Feedback.AuditOperationMessage = fmt.Sprintf("%s: %s", operation.Action, operation.Message)
	}
	m.Audit.Records = m.Dependencies.Audit.RecentOperations()
}

func toastLevel(level audit.Level) state.NotificationLevel {
	switch level {
	case audit.LevelError:
		return state.NotificationError
	case audit.LevelWarn:
		return state.NotificationWarning
	default:
		return state.NotificationInfo
	}
}

func runtimeHost(m *state.AppModel) string {
	if m != nil && m.Connection.Docker != nil {
		return m.Connection.Docker.Host
	}
	return ""
}

func modeLabel(mode state.AppMode) string {
	switch mode {
	case state.ModeDetail:
		return "detail"
	case state.ModeLogView, state.ModeSearch:
		return "logs"
	case state.ModeFilter:
		return "filter"
	case state.ModeConfirm:
		return "confirm"
	case state.ModeTop:
		return "top"
	case state.ModeRename:
		return "rename"
	default:
		return "normal"
	}
}

func panelName(panel state.PanelType) string {
	switch panel {
	case state.PanelContainers:
		return "containers"
	case state.PanelImages:
		return "images"
	case state.PanelVolumes:
		return "volumes"
	case state.PanelNetworks:
		return "networks"
	case state.PanelCompose:
		return "compose"
	case state.PanelLogs:
		return "logs"
	case state.PanelDetail:
		return "detail"
	case state.PanelHelp:
		return "help"
	default:
		return "unknown"
	}
}

func withContainerAudit(cmd tea.Cmd, trace audit.Trace) tea.Cmd {
	if !trace.Valid() {
		return cmd
	}
	return func() tea.Msg {
		msg := cmd()
		if actioned, ok := msg.(state.ContainerActioned); ok {
			actioned.Audit = trace
			return actioned
		}
		return msg
	}
}

func withImageAudit(cmd tea.Cmd, trace audit.Trace) tea.Cmd {
	if !trace.Valid() {
		return cmd
	}
	return func() tea.Msg {
		msg := cmd()
		if actioned, ok := msg.(state.ImageActioned); ok {
			actioned.Audit = trace
			return actioned
		}
		return msg
	}
}

func withGenericAudit(cmd tea.Cmd, trace audit.Trace) tea.Cmd {
	if !trace.Valid() {
		return cmd
	}
	return func() tea.Msg {
		msg := cmd()
		if actioned, ok := msg.(state.GenericActioned); ok {
			actioned.Audit = trace
			return actioned
		}
		return msg
	}
}

func containerTarget(m *state.AppModel, id string) audit.ContainerTarget {
	target := audit.ContainerTarget{ID: id, Name: id}
	if m != nil && m.Resources.Containers != nil {
		for _, item := range m.Resources.Containers.Items {
			if item.ID == id {
				target.Name, target.Meta.Image, target.Meta.State = item.Name, item.Image, item.State
				break
			}
		}
	}
	return target
}

func imageTarget(m *state.AppModel, id string) audit.ImageTarget {
	target := audit.ImageTarget{ID: id, Name: id}
	if m != nil && m.Resources.Images != nil {
		for _, item := range m.Resources.Images.Items {
			if item.ID == id {
				target.Name, target.Meta.RepoTags = firstTag(item.RepoTags), append([]string(nil), item.RepoTags...)
				break
			}
		}
	}
	return target
}

func bulkResourceName(panel state.PanelType) string {
	switch panel {
	case state.PanelContainers:
		return string(state.ResourceContainer)
	case state.PanelImages:
		return string(state.ResourceImage)
	case state.PanelVolumes:
		return string(state.ResourceVolume)
	case state.PanelNetworks:
		return string(state.ResourceNetwork)
	default:
		return "resource"
	}
}

func bulkTarget(m *state.AppModel) audit.Target {
	name := fmt.Sprintf("%d items", len(m.Selection.MarkedIDs))
	switch m.Navigation.ActivePanel {
	case state.PanelContainers:
		return audit.ContainerTarget{ID: "bulk", Name: name}
	case state.PanelImages:
		return audit.ImageTarget{ID: "bulk", Name: name}
	case state.PanelVolumes:
		return audit.VolumeTarget{Name: name}
	case state.PanelNetworks:
		return audit.NetworkTarget{ID: "bulk", Name: name}
	default:
		return audit.ComposeTarget{Name: name}
	}
}

func firstTag(tags []string) string {
	if len(tags) > 0 {
		return tags[0]
	}
	return ""
}
