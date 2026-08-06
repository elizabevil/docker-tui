package view

import (
	"strings"

	"github.com/elizabevil/docker-tui/internal/data/config"
	"github.com/elizabevil/docker-tui/internal/data/i18n"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
	"github.com/elizabevil/docker-tui/internal/tui/ui/pages/audit"
	"github.com/elizabevil/docker-tui/internal/tui/ui/pages/compose"
	"github.com/elizabevil/docker-tui/internal/tui/ui/pages/containers"
	"github.com/elizabevil/docker-tui/internal/tui/ui/pages/detail"
	pageevents "github.com/elizabevil/docker-tui/internal/tui/ui/pages/events"
	"github.com/elizabevil/docker-tui/internal/tui/ui/pages/help"
	"github.com/elizabevil/docker-tui/internal/tui/ui/pages/history"
	"github.com/elizabevil/docker-tui/internal/tui/ui/pages/images"
	"github.com/elizabevil/docker-tui/internal/tui/ui/pages/logs"
	"github.com/elizabevil/docker-tui/internal/tui/ui/pages/networks"
	"github.com/elizabevil/docker-tui/internal/tui/ui/pages/processes"
	"github.com/elizabevil/docker-tui/internal/tui/ui/pages/volumes"
	"github.com/elizabevil/docker-tui/internal/utils"
)

type pageTemplateKind string

const (
	listPageTemplate   pageTemplateKind = "list"
	splitPageTemplate  pageTemplateKind = "split"
	detailPageTemplate pageTemplateKind = "detail"
	logPageTemplate    pageTemplateKind = "log"
	helpPageTemplate   pageTemplateKind = "help"
)

// Title prefixes used in the detail page header. The detail resource
// title (e.g. "Container Detail: %s (%s)") is rewritten by replacing
// these prefixes when the user toggles between Section / YAML / JSON
// views; English and 中文 source views both need replacement so the
// header advertises the active source instead of the parent resource.
const (
	titlePrefixDetailEN  = "Detail:"
	titlePrefixDetailZH  = "详情:"
	titlePrefixYAMLShort = "YAML:"
	titlePrefixJSONShort = "JSON:"
)

type pageView struct {
	template   pageTemplateKind
	title      string
	summary    string
	breadcrumb string
	content    string
}

func templateFor(m *state.AppModel) pageTemplateKind {
	switch {
	case m.Navigation.Mode == state.ModeLogView || m.Navigation.Mode == state.ModeSearch:
		return logPageTemplate
	case m.Navigation.Mode == state.ModeDetail || m.Navigation.Mode == state.ModeExecPassthrough:
		return detailPageTemplate
	case m.Navigation.Mode == state.ModeTop:
		return detailPageTemplate
	case m.Navigation.Mode == state.ModeAuditDetail:
		return detailPageTemplate
	case m.Navigation.Mode == state.ModeHistory:
		return detailPageTemplate
	case m.Navigation.Mode == state.ModeEvents:
		return detailPageTemplate
	case m.Navigation.ActivePanel == state.PanelHelp:
		return helpPageTemplate
	case m.Navigation.ActivePanel == state.PanelCompose:
		return splitPageTemplate
	default:
		return listPageTemplate
	}
}

func projectPage(m *state.AppModel, bodyHeight, bodyWidth int) pageView {
	view := pageView{
		template:   templateFor(m),
		title:      state.PanelLabel(m.Navigation.ActivePanel),
		summary:    m.Feedback.InfoMessage,
		breadcrumb: breadcrumb(m),
	}

	switch view.template {
	case logPageTemplate:
		view.title = state.PanelLabel(state.PanelLogs)
		view.summary = ""
		if m.Log.LogContainerID != "" {
			view.summary = componentShortID(m.Log.LogContainerID)
		}
		view.content = logs.RenderView(m, bodyHeight, bodyWidth)
	case detailPageTemplate:
		view.title = m.Detail.DetailTitle
		if view.title == "" {
			view.title = state.PanelLabel(state.PanelDetail)
		}
		// Adjust title based on source view mode
		switch m.Detail.DetailSourceType {
		case state.DetailSourceYAML:
			view.title = strings.Replace(view.title, titlePrefixDetailEN, titlePrefixYAMLShort, 1)
			view.title = strings.Replace(view.title, titlePrefixDetailZH, titlePrefixYAMLShort, 1)
		case state.DetailSourceJSON:
			view.title = strings.Replace(view.title, titlePrefixDetailEN, titlePrefixJSONShort, 1)
			view.title = strings.Replace(view.title, titlePrefixDetailZH, titlePrefixJSONShort, 1)
		}
		view.summary = ""
		if m.Navigation.Mode == state.ModeExecPassthrough {
			view.title = "Exec"
			view.content = renderExecPassthroughPanel(m, bodyHeight)
		} else if m.Navigation.Mode == state.ModeTop {
			view.title = "Top: " + m.Processes.ContainerName
			view.content = component.WrapActionWindow(
				processes.Render(m.Processes, bodyWidth-4, bodyHeight-2),
				config.OperationScopeContainer,
				bodyWidth,
			)
		} else if m.Navigation.Mode == state.ModeAuditDetail {
			view.title = "Audit Detail"
			view.content = audit.RenderDetail(m.Audit.DetailRecord, bodyWidth, bodyHeight)
		} else if m.Navigation.Mode == state.ModeHistory {
			view.title = i18n.T("history.title")
			view.summary = i18n.T("history.summary", m.History.ImageRef, len(m.History.Layers))
			view.content = component.WrapActionWindow(
				history.RenderView(m, bodyHeight-2, bodyWidth-4),
				config.OperationScopeImage,
				bodyWidth,
			)
		} else if m.Navigation.Mode == state.ModeEvents {
			view.title = i18n.T("events.title")
			view.content = pageevents.RenderView(&m.EventPanel, bodyHeight, bodyWidth)
		} else if m.Detail.DetailResourceType == state.ResourceComposeProject {
			view.content = compose.RenderProjectDetailTable(m, bodyWidth, bodyHeight)
		} else {
			view.content = detail.RenderView(m, bodyHeight, bodyWidth)
		}
	case helpPageTemplate:
		view.content = help.RenderView(bodyWidth, m)
	case splitPageTemplate:
		view.content = compose.RenderPanel(m, bodyWidth, bodyHeight)
	default:
		if m.Navigation.ActivePanel == state.PanelImages && m.Resources.Images.ContainersViewID != "" {
			view.summary = "Image: " + m.Resources.Images.ContainersViewRef
		}
		view.content = renderListPage(m, bodyHeight, bodyWidth)
	}
	return view
}

func renderListPage(m *state.AppModel, panelHeight, contentWidth int) string {
	selectionDisabled := m.Navigation.Mode == state.ModeFilter
	switch m.Navigation.ActivePanel {
	case state.PanelContainers:
		return containers.RenderList(m.Resources.Containers, contentWidth, panelHeight, m.Selection.PanelMarks[state.PanelContainers], selectionDisabled)
	case state.PanelImages:
		return images.RenderList(m.Resources.Images, m.Resources.Containers, contentWidth, panelHeight, m.Selection.PanelMarks[state.PanelImages], selectionDisabled)
	case state.PanelVolumes:
		return volumes.RenderList(m.Resources.Volumes, m.Resources.Containers, contentWidth, panelHeight, m.Selection.PanelMarks[state.PanelVolumes], selectionDisabled)
	case state.PanelNetworks:
		return networks.RenderList(m.Resources.Networks, contentWidth, panelHeight, m.Selection.PanelMarks[state.PanelNetworks], selectionDisabled)
	case state.PanelAudit:
		return audit.RenderList(&m.Audit, contentWidth, panelHeight)
	default:
		return ""
	}
}

func componentShortID(id string) string {
	return utils.ShortID(id)
}
