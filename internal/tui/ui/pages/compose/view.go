package compose

import (
	"fmt"
	"sort"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/elizabevil/docker-tui/internal/tui/keys"

	"github.com/elizabevil/docker-tui/internal/data/i18n"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/tables"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"

	dockerclient "github.com/elizabevil/docker-tui/internal/data/docker"
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
		if c.State == "running" {
			item.running++
		}
		info := item.svcs[s]
		info.count++
		info.total++
		if c.State == "running" {
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
			return component.GetStyle("dim").Render("(no compose projects match filter)")
		}
		return component.GetStyle("dim").Render("(no compose projects found)")
	}
	projectCursor := m.Compose.ProjectCursor(len(ordered))

	totalW := panelWidth - 4

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

	barColor := lipgloss.Color(activeColor)
	noFocus := lipgloss.Color(inactiveColor)
	topBar := func(focused bool, w int) string {
		c := noFocus
		if focused {
			c = barColor
		}
		return lipgloss.NewStyle().Foreground(c).Render(strings.Repeat("\u2500", w))
	}
	leftBar := topBar(m.Compose.ComposeFocus == 0, leftW)
	rightBar := topBar(m.Compose.ComposeFocus == 1, rightW)

	left = leftBar + "\n" + left
	right = rightBar + "\n" + right

	// 顶部竖线分隔（仅首行对齐，不撑高）
	sep := component.GetStyle("panelTitle").Render("\u2502")

	return lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(leftW).Render(left),
		lipgloss.NewStyle().Width(1).Render(sep),
		lipgloss.NewStyle().Width(rightW).Render(right),
	)
}

func renderProjectList(m *state.AppModel, ordered []composeProj, w, panelHeight int) string {
	widths := tc.ColumnWidths("default", w)
	colsDef := tc.Columns["default"]
	if widths == nil {
		return ""
	}

	total := len(ordered)
	rowHeight := component.CalcRowHeight(panelHeight)
	rows := make([][]string, 0, rowHeight)
	for i := 0; i < total && len(rows) < rowHeight; i++ {
		p := ordered[i]
		sts := "partial"
		if p.running == p.total {
			sts = "running"
		}
		if p.running == 0 {
			sts = "stopped"
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

	colStyles := component.GetPageColumnStyles("container", colsDef)
	ts := tc.EffectiveTableStyle()
	bannerW := tableBannerWidth(w, ts)
	return component.RenderTable(component.TableData{
		Cols:              colsDef,
		Widths:            widths,
		Rows:              rows,
		Selected:          projectCursor,
		Total:             total,
		Limit:             rowHeight,
		BodyHeight:        panelHeight,
		Banner:            banner,
		BannerW:           bannerW,
		FooterHint:        fmt.Sprintf("%d projects", total),
		RowPrefix:         ts.RowPrefix,
		RowPrefixSelected: ts.RowPrefixSelected,
		ColStyles:         colStyles,
	})
}

func renderServicePanel(m *state.AppModel, proj composeProj, w, panelHeight int) string {
	sts := "partial"
	if proj.running == proj.total {
		sts = "running"
	}
	if proj.running == 0 {
		sts = "stopped"
	}
	rawTitle := "Services: " + proj.name
	rawSummary := fmt.Sprintf("Status: %s  Services: %d  Pods: %d", sts, len(proj.svcs), proj.total)
	title := component.GetStyle("panelTitle").Render(component.TruncateVisible(rawTitle, w))
	summary := component.GetStyle("dim").Render(component.TruncateVisible(rawSummary, w))

	svcNames := make([]string, 0, len(proj.svcs))
	for name := range proj.svcs {
		if m.Compose.ComposeServiceFilter != "" && !strings.Contains(name, m.Compose.ComposeServiceFilter) {
			continue
		}
		svcNames = append(svcNames, name)
	}
	sort.Strings(svcNames)
	if len(svcNames) == 0 {
		msg := "(no services found)"
		if m.Compose.ComposeServiceFilter != "" {
			msg = "(no services match filter)"
		}
		return lipgloss.JoinVertical(lipgloss.Top,
			title,
			summary,
			"",
			component.GetStyle("dim").Render(msg),
		)
	}
	serviceCursor := m.Compose.ServiceCursor(len(svcNames))

	widths := tc.ColumnWidths("services_sub", w)
	colsDef := tc.Columns["services_sub"]
	total := len(svcNames)
	rowHeight := component.CalcRowHeight(panelHeight - 3)
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

	colStyles := component.GetPageColumnStyles("container", colsDef)
	ts := tc.EffectiveTableStyle()
	bannerW := tableBannerWidth(w, ts)
	table := component.RenderTable(component.TableData{
		Cols:              colsDef,
		Widths:            widths,
		Rows:              rows,
		Selected:          serviceCursor,
		Total:             total,
		Limit:             rowHeight,
		BodyHeight:        panelHeight - 3,
		BannerW:           bannerW,
		FooterHint:        "s:start S:stop l:logs",
		RowPrefix:         ts.RowPrefix,
		RowPrefixSelected: ts.RowPrefixSelected,
		ColStyles:         colStyles,
	})

	return lipgloss.JoinVertical(lipgloss.Top,
		title,
		summary,
		"",
		table,
	)
}

// RenderProjectDetail 从模型获取当前项目数据，生成概览文本供 ModeDetail 使用。
func RenderProjectDetail(m *state.AppModel) string {
	ordered, _ := gatherComposeProjects(m)
	if len(ordered) == 0 {
		return ""
	}
	return BuildProjectDetail(ordered[m.Compose.ProjectCursor(len(ordered))])
}

// BuildProjectDetail 生成 compose 项目概览文本，供给 ModeDetail 渲染。
// 格式使用 ── 标题 ── 分隔，被 detail/buildDetailSections 自动解析为多节。
func BuildProjectDetail(proj composeProj) string {
	var b strings.Builder

	b.WriteString("\u2500\u2500 \u9879\u76ee\u6982\u51b5 \u2500\u2500\n")
	b.WriteString(fmt.Sprintf("Project: %s\n", proj.name))
	b.WriteString(fmt.Sprintf("Containers: %d/%d running\n", proj.running, proj.total))

	b.WriteString("\n\u2500\u2500 \u670d\u52a1\u5217\u8868 \u2500\u2500\n")
	svcNames := make([]string, 0, len(proj.svcs))
	for name := range proj.svcs {
		svcNames = append(svcNames, name)
	}
	sort.Strings(svcNames)
	for _, name := range svcNames {
		info := proj.svcs[name]
		ico := "\u25cf"
		if info.running == 0 {
			ico = "\u25cb"
		}
		b.WriteString(fmt.Sprintf("  %s %s  %d/%d  image: %s\n", ico, name, info.running, info.total, info.image))
	}

	b.WriteString("\n\u2500\u2500 \u5feb\u6377\u64cd\u4f5c \u2500\u2500\n")
	b.WriteString("  s  Start all     S  Stop all\n")
	b.WriteString("  l  Service logs    Ctrl+D  Compose down\n")
	b.WriteString("  Enter \u2192 Container list   d  This view\n")

	return b.String()
}

// renderComposeContainers 渲染 compose 服务下的容器子视图。
func renderComposeContainers(m *state.AppModel, panelWidth int, panelHeight int) string {
	widths := tc.ColumnWidths("pods_sub", panelWidth-4)
	colsDef := tc.Columns["pods_sub"]
	if widths == nil {
		return component.GetStyle("dim").Render("(no container data)")
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
		return component.GetStyle("dim").Render("(no containers for " + service + ")")
	}

	rowHeight := component.CalcRowHeight(panelHeight - 2) // title + breadcrumb
	containerCursor := m.Compose.ContainerCursor(total)
	viewOffset := 0
	component.EnsureVisible(&viewOffset, containerCursor, rowHeight, total)

	rows := component.BuildRows(matched, colsDef, viewOffset, rowHeight,
		func(c dockerclient.ContainerSummary, col tables.ColumnDef, idx int) string {
			switch col.Key {
			case "id":
				return c.ID[:12]
			case "name":
				return c.Name
			case "state":
				return component.RenderStateText(c.State)
			case "ip":
				if len(c.IPs) > 0 {
					return c.IPs[0]
				}
				return "\u2014"
			default:
				return ""
			}
		})

	rawTitle := component.TruncateVisible(project+"/"+service, panelWidth-4)
	title := component.GetStyle("panelTitle").Render(rawTitle)
	bc := component.GetStyle("dim").Render("Esc " + i18n.T("key.back"))
	ts := tc.EffectiveTableStyle()

	return lipgloss.JoinVertical(lipgloss.Top,
		title,
		component.RenderTable(component.TableData{
			Cols:              colsDef,
			Widths:            widths,
			Rows:              rows,
			Selected:          containerCursor - viewOffset,
			Total:             total,
			Limit:             rowHeight,
			BodyHeight:        panelHeight - 2,
			BannerW:           tableBannerWidth(panelWidth-4, ts),
			RowPrefix:         ts.RowPrefix,
			RowPrefixSelected: ts.RowPrefixSelected,
			FooterHint:        fmt.Sprintf("Esc back  |  l:logs  d:detail"),
		}),
		"",
		bc,
	)
}

func tableBannerWidth(totalW int, ts tables.TableStyle) int {
	prefixW := component.VisibleLen(ts.RowPrefix)
	selPrefixW := component.VisibleLen(ts.RowPrefixSelected)
	if selPrefixW > prefixW {
		prefixW = selPrefixW
	}
	bw := totalW - prefixW
	if bw < 10 {
		bw = 10
	}
	return bw
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
