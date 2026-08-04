package panel

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/elizabevil/docker-tui/internal/tui/ui/style"

	"github.com/elizabevil/docker-tui/internal/tui"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
)

// Panel 渲染中间主区域的统一带边框容器。
type Panel struct {
	Title       string
	Info        string
	Content     string
	Breadcrumb  string
	SearchText  string // 兼容字段：搜索框已迁移到 header 与 panel 之间
	BorderLabel string // 面板上边框标签（如：搜索内容）
	Width       int
	Height      int
}

// Render 输出带边框的面板：标题行 + 内容区。
func (p Panel) Render() string {
	if p.Width <= 0 || p.Height <= 0 {
		return ""
	}
	if p.Height < 4 {
		p.Height = 4
	}

	// 标题行
	titleLine := renderTitle(p)

	innerH := p.Height - 2 // 去掉边框
	contentH := innerH - 1 // 去掉标题行
	if contentH < 1 {
		contentH = 1
	}
	body := lipgloss.NewStyle().MaxHeight(contentH).Height(contentH).Background(style.Colors.BG).Render(p.Content)
	inner := lipgloss.JoinVertical(lipgloss.Top, titleLine, body)
	boxed := tui.ActiveBorderStyle.Width(p.Width - 2).Render(inner)
	if p.BorderLabel != "" {
		boxed = applyBorderLabel(boxed, p.BorderLabel, p.Width-2)
	}
	return boxed
}

// renderTitle 标题行：面板名 + 面包屑(右侧)
func renderTitle(p Panel) string {
	title := p.Title
	if p.Info != "" {
		title += " \u2502 " + p.Info
	}
	titleLine := component.GetStyle("panelTitle").Render(title)
	if p.Breadcrumb != "" {
		lineW := p.Width - 6
		if lineW < 10 {
			lineW = 10
		}
		right := component.GetStyle("dim").Render(p.Breadcrumb)
		titleLine = component.JustifyBetween(titleLine, right, lineW)
	}
	return titleLine
}

func applyBorderLabel(boxed, label string, width int) string {
	if boxed == "" || label == "" {
		return boxed
	}
	lines := strings.Split(boxed, "\n")
	if len(lines) == 0 {
		return boxed
	}
	lines[0] = buildBorderLabelLine(label, width)
	return strings.Join(lines, "\n")
}

func buildBorderLabelLine(label string, width int) string {
	if width < 4 {
		return ""
	}
	inner := width - 2
	lab := " " + label + " "
	if component.VisibleLen(lab) > inner {
		lab = component.TruncateVisible(lab, inner)
	}
	remain := inner - component.VisibleLen(lab)
	if remain < 0 {
		remain = 0
	}
	left := remain / 2
	right := remain - left
	return "╭" + strings.Repeat("─", left) + lab + strings.Repeat("─", right) + "╮"
}

// RenderNoBorder 渲染无外层边框的面板。
func (p Panel) RenderNoBorder() string {
	if p.Width <= 0 {
		return ""
	}
	title := p.Title
	if p.Info != "" {
		title += " \u2502 " + p.Info
	}
	return lipgloss.JoinVertical(lipgloss.Top,
		component.GetStyle("panelTitle").Render(title),
		p.Content)
}

// AsPanel 快捷创建带边框面板。
func AsPanel(title, content string, w, h int) string {
	return Panel{Title: title, Content: content, Width: w, Height: h}.Render()
}

// Breadcrumb 构建面包屑字符串。
func Breadcrumb(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, " > ")
}
