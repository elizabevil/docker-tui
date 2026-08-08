package compose

import (
	"fmt"
	"sort"
	"strings"

	"charm.land/lipgloss/v2"
	dockerclient "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/keys"

	"github.com/elizabevil/docker-tui/internal/data/i18n"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/tables"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
	"github.com/elizabevil/docker-tui/internal/tui/ui/style"
	"github.com/elizabevil/docker-tui/internal/utils"
)

var tc = tables.MustLoad("compose")

type composeProj struct {
	name    string
	total   int
	running int
	svcs    map[string]composeSvc
}

type composeSvc struct {
	count   int
	image   string
	running int
	total   int
}

func gatherComposeProjects(m *state.AppModel) ([]composeProj, map[string]*composeProj) {
	projects := map[string]*composeProj{}
	for _, c := range m.Resources.Containers.Items {
		p := c.ComposeProject
		if p == "" {
			continue
		}
		s := c.ComposeService
		if s == "" {
			s = "unknown"
		}
		item, ok := projects[p]
		if !ok {
			item = &composeProj{name: p, svcs: map[string]composeSvc{}}
			projects[p] = item
		}
		item.total++
		if c.State == state.ContainerStateRunning {
			item.running++
		}
		info := item.svcs[s]
		info.count++
		info.total++
		if c.State == state.ContainerStateRunning {
			info.running++
		}
		if info.image == "" {
			info.image = c.Image
		}
		item.svcs[s] = info
	}
	names := make([]string, 0, len(projects))
	for name := range projects {
		names = append(names, name)
	}
	sort.Strings(names)
	ordered := make([]composeProj, 0, len(names))
	for _, name := range names {
		ordered = append(ordered, *projects[name])
	}
	return ordered, projects
}

func RenderPanel(m *state.AppModel, panelWidth int, panelHeight int) string {
	// 容器子视图模式
	if m.Compose.ComposeContainerViewID != "" {
		return renderComposeContainers(m, panelWidth, panelHeight)
	}

	ordered, _ := gatherComposeProjects(m)
	ordered = filterProjects(ordered, m.Compose.ComposeProjectFilter)
	if len(ordered) == 0 {
		if m.Compose.ComposeProjectFilter != "" {
			return component.GetStyle(component.StyleDim).Render(i18n.T("compose.empty_projects_filter"))
		}
		return component.GetStyle(component.StyleDim).Render(i18n.T("compose.empty_projects"))
	}
	projectCursor := m.Compose.ProjectCursor(len(ordered))

	totalW := panelWidth

	ratioL, ratioR := 4, 6
	if tc.Layout != nil && tc.Layout.Ratio.Left > 0 && tc.Layout.Ratio.Right > 0 {
		ratioL = tc.Layout.Ratio.Left
		ratioR = tc.Layout.Ratio.Right
	}
	leftW := totalW * ratioL / (ratioL + ratioR)
	rightW := totalW - leftW - 1
	if leftW < 10 {
		leftW = 10
	}
	if rightW < 10 {
		rightW = 10
	}

	// topBar 占 1 行，表格内容高度减 1
	bodyH := panelHeight - 1
	if bodyH < 3 {
		bodyH = 3
	}

	left := renderProjectList(m, ordered, leftW, bodyH)
	right := renderServicePanel(m, ordered[projectCursor], rightW, bodyH)

	// 顶部焦点色条：聚焦面板上方显示彩色横线
	focus := tc.Layout.Focus
	activeColor := keys.FocusActive
	inactiveColor := keys.FocusInactive
	if focus.ActiveColor != "" {
		activeColor = focus.ActiveColor
	}
	if focus.InactiveColor != "" {
		inactiveColor = focus.InactiveColor
	}

	barColor := style.Color(activeColor)
	noFocus := style.Color(inactiveColor)
	topBar := func(focused bool, w int) string {
		c := noFocus
		if focused {
			c = barColor
		}
		return style.ApplyForeground(lipgloss.NewStyle(), c).Render(strings.Repeat(component.BorderLineHorizontal, w))
	}
	leftBar := topBar(m.Compose.ComposeFocus == 0, leftW)
	rightBar := topBar(m.Compose.ComposeFocus == 1, rightW)

	left = leftBar + "\n" + left
	right = rightBar + "\n" + right

	// 顶部竖线分隔（仅首行对齐，不撑高）
	sep := component.GetStyle(component.StylePanelTitle).Render(component.BorderLineVertical)

	return lipgloss.JoinHorizontal(lipgloss.Top,
		component.GetStyle(component.StylePanel).Width(leftW).Render(left),
		component.GetStyle(component.StylePanel).Width(1).Render(sep),
		component.GetStyle(component.StylePanel).Width(rightW).Render(right),
	)
}

func renderProjectList(m *state.AppModel, ordered []composeProj, w, panelHeight int) string {
	colsDef := tc.Columns.Get("default")
	if len(colsDef) == 0 {
		return ""
	}

	total := len(ordered)
	rowHeight := component.CalcRowHeight(panelHeight)
	rows := make([][]string, 0, rowHeight)
	for i := 0; i < total && len(rows) < rowHeight; i++ {
		p := ordered[i]
		sts := "partial"
		if p.running == p.total {
			sts = state.ContainerStateRunning
		}
		if p.running == 0 {
			sts = state.ContainerStateStopped
		}
		cells := make([]string, len(colsDef))
		for j, cd := range colsDef {
			switch cd.Key {
			case "project":
				cells[j] = p.name
			case "status":
				cells[j] = sts
			case "services":
				cells[j] = fmt.Sprintf("%d", len(p.svcs))
			case "pods":
				cells[j] = fmt.Sprintf("%d", p.total)
			}
		}
		rows = append(rows, cells)
	}

	projectCursor := m.Compose.ProjectCursor(total)
	banner := ordered[projectCursor].name

	colStyles := component.GetColumnStyles(colsDef)
	bannerW := tableBannerWidth(w)
	return component.RenderTable(component.TableData{
		Cols:       colsDef,
		Rows:       rows,
		Selected:   projectCursor,
		Total:      total,
		Limit:      rowHeight,
		BodyHeight: panelHeight,
		Banner:     banner,
		BannerW:    bannerW,
		FooterHint: i18n.T("compose.footer_projects", total),
		ColStyles:  colStyles,
	})
}

func renderServicePanel(m *state.AppModel, proj composeProj, w, panelHeight int) string {
	sts := "partial"
	if proj.running == proj.total {
		sts = state.ContainerStateRunning
	}
	if proj.running == 0 {
		sts = state.ContainerStateStopped
	}
	rawTitle := state.PanelLabel(state.PanelCompose) + " > " + proj.name + " > " + i18n.T("key.services")
	rawSummary := i18n.T("compose.summary", sts, len(proj.svcs), proj.total)
	// header combines the breadcrumb (left) with the running / service / pod
	// counts (right). The summary is treated as the canonical "stats" slot for
	// this view; any duplicate shortcut hints are intentionally omitted so the
	// header stays compact and the kebab shortcuts elsewhere remain the source
	// of truth.
	title := component.GetStyle(component.StylePanelTitle).Render(component.TruncateVisible(rawTitle, w))
	summary := component.GetStyle(component.StyleDim).Render(rawSummary)
	header := composePanelHeader(title, summary, w)

	svcNames := make([]string, 0, len(proj.svcs))
	for name := range proj.svcs {
		if m.Compose.ComposeServiceFilter != "" && !strings.Contains(name, m.Compose.ComposeServiceFilter) {
			continue
		}
		svcNames = append(svcNames, name)
	}
	sort.Strings(svcNames)
	if len(svcNames) == 0 {
		msg := i18n.T("compose.empty_services")
		if m.Compose.ComposeServiceFilter != "" {
			msg = i18n.T("compose.empty_services_filter")
		}
		return lipgloss.JoinVertical(lipgloss.Top,
			header,
			"",
			component.GetStyle(component.StyleDim).Render(msg),
		)
	}
	serviceCursor := m.Compose.ServiceCursor(len(svcNames))

	colsDef := tc.Columns.Get("services_sub")
	total := len(svcNames)
	rowHeight := component.CalcRowHeight(panelHeight - 2)
	rows := make([][]string, 0, rowHeight)
	for i := 0; i < total && len(rows) < rowHeight; i++ {
		name := svcNames[i]
		info := proj.svcs[name]
		cells := make([]string, len(colsDef))
		for j, cd := range colsDef {
			switch cd.Key {
			case "service":
				cells[j] = name
			case "image":
				cells[j] = info.image
			case "pods":
				cells[j] = fmt.Sprintf("%d", info.count)
			}
		}
		rows = append(rows, cells)
	}

	colStyles := component.GetColumnStyles(colsDef)
	bannerW := tableBannerWidth(w)
	table := component.RenderTable(component.TableData{
		Cols:       colsDef,
		Rows:       rows,
		Selected:   serviceCursor,
		Total:      total,
		Limit:      rowHeight,
		BodyHeight: panelHeight - 2,
		BannerW:    bannerW,
		ColStyles:  colStyles,
	})

	return lipgloss.JoinVertical(lipgloss.Top,
		header,
		"",
		table,
	)
}

// composePanelHeader lays out title (left) and summary (right) on one line so
// the breadcrumb + services count share the panel title row, matching the
// convention used by other Compose sub-views. width <= 0 is treated as
// unlimited so the rendered strings can include their full styling.
func composePanelHeader(title, summary string, width int) string {
	if width <= 0 || utils.DisplayWidth(title)+utils.DisplayWidth(summary)+1 >= width {
		return lipgloss.JoinVertical(lipgloss.Top, title, summary)
	}
	gap := width - utils.DisplayWidth(title) - utils.DisplayWidth(summary)
	return lipgloss.NewStyle().Width(width).Render(
		lipgloss.JoinHorizontal(lipgloss.Top, title, strings.Repeat(" ", gap), summary),
	)
}

// RenderProjectDetailTable renders a Compose project as a read-only detail table.
func RenderProjectDetailTable(m *state.AppModel, width, panelHeight int) string {
	ordered, _ := gatherComposeProjects(m)
	ordered = filterProjects(ordered, m.Compose.ComposeProjectFilter)
	if len(ordered) == 0 {
		return component.GetStyle(component.StyleDim).Render(i18n.T("compose.empty_detail"))
	}
	proj := ordered[m.Compose.ProjectCursor(len(ordered))]
	svcNames := make([]string, 0, len(proj.svcs))
	for name := range proj.svcs {
		svcNames = append(svcNames, name)
	}
	sort.Strings(svcNames)
	if len(svcNames) == 0 {
		return component.GetStyle(component.StyleDim).Render(i18n.T("compose.empty_services"))
	}

	w := max(40, width-4)
	colsDef := tc.Columns.Get("detail")
	if len(colsDef) == 0 {
		return component.GetStyle(component.StyleDim).Render(i18n.T("compose.empty_detail_columns"))
	}
	rowHeight := component.CalcTableRowHeight(panelHeight, false)
	offset := m.Detail.ClampVisibleOffset(len(svcNames), rowHeight)
	rows := make([][]string, 0, rowHeight)
	for i := offset; i < len(svcNames) && len(rows) < rowHeight; i++ {
		name := svcNames[i]
		info := proj.svcs[name]
		status := state.ContainerStateStopped
		if info.running == info.total {
			status = state.ContainerStateRunning
		} else if info.running > 0 {
			status = "partial"
		}
		cells := make([]string, len(colsDef))
		for j, column := range colsDef {
			switch column.Key {
			case "service":
				cells[j] = name
			case "image":
				cells[j] = info.image
			case "status":
				cells[j] = status
			case "pods":
				cells[j] = fmt.Sprintf("%d/%d", info.running, info.total)
			}
		}
		rows = append(rows, cells)
	}
	return component.RenderTable(component.TableData{
		Cols:       colsDef,
		Rows:       rows,
		Selected:   -1,
		Total:      len(svcNames),
		Offset:     offset,
		Limit:      rowHeight,
		BannerW:    w,
		FooterHint: i18n.T("compose.detail_footer", proj.running, proj.total, len(proj.svcs)),
		ColStyles:  component.GetColumnStyles(colsDef),
	})
}

// renderComposeContainers 渲染 compose 服务下的容器子视图。
func renderComposeContainers(m *state.AppModel, panelWidth int, panelHeight int) string {
	colsDef := tc.Columns.Get("pods_sub")
	if len(colsDef) == 0 {
		return component.GetStyle(component.StyleDim).Render(i18n.T("compose.empty_container_columns"))
	}
	// §4.3 decision C: docker runtime 不渲染 Pod 列, 列宽回收给其它列。
	if !podScopeSupported(m) {
		colsDef = filterOutColumn(colsDef, "pod")
	}

	// 收集该服务的容器
	project := currentProjectName(m)
	service := m.Compose.ComposeContainerViewID
	matched := make([]dockerclient.ContainerSummary, 0, 8)
	for _, c := range m.Resources.Containers.Items {
		if c.ComposeProject == project && c.ComposeService == service {
			matched = append(matched, c)
		}
	}

	total := len(matched)
	if total == 0 {
		return component.GetStyle(component.StyleDim).Render(i18n.T("compose.empty_containers_for", service))
	}

	rowHeight := component.CalcRowHeight(panelHeight - 2) // title + breadcrumb
	containerCursor := m.Compose.ContainerCursor(total)
	viewOffset := 0
	component.EnsureVisible(&viewOffset, containerCursor, rowHeight, total)

	rows := component.BuildRows(matched, colsDef, viewOffset, rowHeight,
		func(c dockerclient.ContainerSummary, col tables.ColumnDef, idx int) string {
			switch col.Key {
			case "id":
				return utils.ShortID(c.ID)
			case "name":
				return c.Name
			case "state":
				return component.RenderStateText(c.State)
			case "pod":
				if c.CoLocatedGroupID == "" {
					return component.StrDash
				}
				return c.CoLocatedGroupID
			case "ip":
				if len(c.IPs) > 0 {
					return c.IPs[0]
				}
				return component.StrDash
			default:
				return ""
			}
		})

	rawTitle := component.TruncateVisible(project+"/"+service, panelWidth)
	title := component.GetStyle(component.StylePanelTitle).Render(rawTitle)
	bc := component.GetStyle(component.StyleDim).Render("Esc " + i18n.T("key.back"))
	return lipgloss.JoinVertical(lipgloss.Top,
		title,
		component.RenderTable(component.TableData{
			Cols:       colsDef,
			Rows:       rows,
			Selected:   containerCursor - viewOffset,
			Total:      total,
			Limit:      rowHeight,
			BodyHeight: panelHeight - 2,
			BannerW:    tableBannerWidth(panelWidth),
			ColStyles:  component.GetColumnStyles(colsDef),
			FooterHint: i18n.T("compose.footer_hint_containers"),
		}),
		"",
		bc,
	)
}

func tableBannerWidth(totalW int) int {
	return max(10, totalW)
}

// currentProjectName 返回当前 Compose 光标所在的项目名。
func currentProjectName(m *state.AppModel) string {
	ordered, _ := gatherComposeProjects(m)
	ordered = filterProjects(ordered, m.Compose.ComposeProjectFilter)
	if len(ordered) == 0 {
		return ""
	}
	return ordered[m.Compose.ProjectCursor(len(ordered))].name
}

func filterProjects(projects []composeProj, query string) []composeProj {
	if query == "" {
		return projects
	}
	out := make([]composeProj, 0, len(projects))
	for _, p := range projects {
		if strings.Contains(p.name, query) {
			out = append(out, p)
		}
	}
	return out
}

// podScopeSupported returns true when the active engine advertises
// CapabilityComposePodScope at any non-Unsupported level. nil engine
// returns false (fail closed — do not show the pod column when no
// runtime is connected).
func podScopeSupported(m *state.AppModel) bool {
	if m == nil || m.Connection.Engine == nil {
		return false
	}
	return m.Connection.Engine.Capabilities().Supports(dockerclient.CapabilityComposePodScope)
}

// filterOutColumn returns cols minus any ColumnDef whose Key matches
// the supplied target. Order is preserved.
func filterOutColumn(cols []tables.ColumnDef, key string) []tables.ColumnDef {
	out := make([]tables.ColumnDef, 0, len(cols))
	for _, c := range cols {
		if c.Key == key {
			continue
		}
		out = append(out, c)
	}
	return out
}
