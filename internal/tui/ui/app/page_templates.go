package view

import (
	"strings"

	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/pages/compose"
	"github.com/elizabevil/docker-tui/internal/tui/ui/pages/containers"
	"github.com/elizabevil/docker-tui/internal/tui/ui/pages/detail"
	"github.com/elizabevil/docker-tui/internal/tui/ui/pages/help"
	"github.com/elizabevil/docker-tui/internal/tui/ui/pages/images"
	"github.com/elizabevil/docker-tui/internal/tui/ui/pages/logs"
	"github.com/elizabevil/docker-tui/internal/tui/ui/pages/networks"
	"github.com/elizabevil/docker-tui/internal/tui/ui/pages/volumes"
)

type pageTemplateKind string

const (
	listPageTemplate   pageTemplateKind = "list"
	splitPageTemplate  pageTemplateKind = "split"
	detailPageTemplate pageTemplateKind = "detail"
	logPageTemplate    pageTemplateKind = "log"
	helpPageTemplate   pageTemplateKind = "help"
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
		if m.Detail.DetailSourceType == "yaml" {
			view.title = strings.Replace(view.title, "Detail:", "YAML:", 1)
			view.title = strings.Replace(view.title, "详情:", "YAML:", 1)
		} else if m.Detail.DetailSourceType == "json" {
			view.title = strings.Replace(view.title, "Detail:", "JSON:", 1)
			view.title = strings.Replace(view.title, "详情:", "JSON:", 1)
		}
		view.summary = ""
		if m.Navigation.Mode == state.ModeExecPassthrough {
			view.title = "Exec"
			view.content = renderExecPassthroughPanel(m, bodyHeight)
		} else {
			view.content = detail.RenderView(m, bodyHeight)
		}
	case helpPageTemplate:
		view.content = help.RenderView(bodyWidth, m)
	case splitPageTemplate:
		view.content = compose.RenderPanel(m, bodyWidth, bodyHeight)
	default:
		view.content = renderListPage(m, bodyHeight, bodyWidth)
	}
	return view
}

func renderListPage(m *state.AppModel, panelHeight, contentWidth int) string {
	selectionDisabled := m.Navigation.Mode == state.ModeFilter
	switch m.Navigation.ActivePanel {
	case state.PanelContainers:
		return containers.RenderList(m.Resources.Containers, contentWidth, panelHeight, m.Selection.MarkedIDs, selectionDisabled)
	case state.PanelImages:
		return images.RenderList(m.Resources.Images, m.Resources.Containers, contentWidth, panelHeight, m.Selection.MarkedIDs, selectionDisabled)
	case state.PanelVolumes:
		return volumes.RenderList(m.Resources.Volumes, m.Resources.Containers, contentWidth, panelHeight, m.Selection.MarkedIDs, selectionDisabled)
	case state.PanelNetworks:
		return networks.RenderList(m.Resources.Networks, contentWidth, panelHeight, m.Selection.MarkedIDs, selectionDisabled)
	default:
		return ""
	}
}

func componentShortID(id string) string {
	if len(id) > 12 {
		return id[:12]
	}
	return id
}
