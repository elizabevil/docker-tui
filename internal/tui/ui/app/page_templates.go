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
	case m.Mode == state.ModeLogView || m.Mode == state.ModeSearch:
		return logPageTemplate
	case m.Mode == state.ModeDetail || m.Mode == state.ModeExecPassthrough:
		return detailPageTemplate
	case m.ActivePanel == state.PanelHelp:
		return helpPageTemplate
	case m.ActivePanel == state.PanelCompose:
		return splitPageTemplate
	default:
		return listPageTemplate
	}
}

func projectPage(m *state.AppModel, bodyHeight, bodyWidth int) pageView {
	view := pageView{
		template:   templateFor(m),
		title:      state.PanelLabel(m.ActivePanel),
		summary:    m.InfoMessage,
		breadcrumb: breadcrumb(m),
	}

	switch view.template {
	case logPageTemplate:
		view.title = state.PanelLabel(state.PanelLogs)
		view.summary = ""
		if m.LogContainerID != "" {
			view.summary = componentShortID(m.LogContainerID)
		}
		view.content = logs.RenderView(m, bodyHeight, bodyWidth)
	case detailPageTemplate:
		view.title = m.DetailTitle
		if view.title == "" {
			view.title = state.PanelLabel(state.PanelDetail)
		}
		// Adjust title based on source view mode
		if m.DetailSourceType == "yaml" {
			view.title = strings.Replace(view.title, "Detail:", "YAML:", 1)
			view.title = strings.Replace(view.title, "详情:", "YAML:", 1)
		} else if m.DetailSourceType == "json" {
			view.title = strings.Replace(view.title, "Detail:", "JSON:", 1)
			view.title = strings.Replace(view.title, "详情:", "JSON:", 1)
		}
		view.summary = ""
		if m.Mode == state.ModeExecPassthrough {
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
	selectionDisabled := m.Mode == state.ModeFilter
	switch m.ActivePanel {
	case state.PanelContainers:
		return containers.RenderList(m.Containers, contentWidth, panelHeight, m.MarkedIDs, selectionDisabled)
	case state.PanelImages:
		return images.RenderList(m.Images, m.Containers, contentWidth, panelHeight, m.MarkedIDs, selectionDisabled)
	case state.PanelVolumes:
		return volumes.RenderList(m.Volumes, m.Containers, contentWidth, panelHeight, m.MarkedIDs, selectionDisabled)
	case state.PanelNetworks:
		return networks.RenderList(m.Networks, contentWidth, panelHeight, m.MarkedIDs, selectionDisabled)
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
