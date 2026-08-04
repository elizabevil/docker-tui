package help

import (
	_ "embed"
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/elizabevil/docker-tui/internal/data/i18n"
	"github.com/elizabevil/docker-tui/internal/tui/ui/action"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
	"github.com/elizabevil/docker-tui/internal/tui/ui/style"

	"github.com/elizabevil/docker-tui/internal/tui"
	"github.com/elizabevil/docker-tui/internal/tui/state"
)

//go:embed about.txt
var aboutText string

type helpSection struct {
	Title string
	Items []helpItem
}

type helpItem struct {
	Keys string
	Desc string
}

// SectionProvider 为帮助页提供分组数据。
// 通过接口解耦"渲染流程"和"内容来源"，便于后续替换为动态来源。
type SectionProvider interface {
	Sections(*state.AppModel) []helpSection
}

type defaultSectionProvider struct{}

func (defaultSectionProvider) Sections(app *state.AppModel) []helpSection {
	registrySections := action.Sections(app)
	sections := make([]helpSection, 0, len(registrySections))
	for _, section := range registrySections {
		items := make([]helpItem, len(section.Shortcuts))
		for i, shortcut := range section.Shortcuts {
			items[i] = helpItem{Keys: shortcut.Key, Desc: shortcut.Description}
		}
		sections = append(sections, helpSection{Title: section.Title, Items: items})
	}
	return sections
}

// Renderer 绑定帮助页配置与渲染行为，减少散落函数。
type Renderer struct {
	provider SectionProvider
}

func NewRenderer(provider SectionProvider) *Renderer {
	if provider == nil {
		provider = defaultSectionProvider{}
	}
	return &Renderer{provider: provider}
}

func (r *Renderer) Render(width int, app *state.AppModel) string {
	if width < 40 {
		width = 40
	}

	// Left column: logo + about text
	leftW := width * 35 / 100
	if leftW < 20 {
		leftW = 20
	}
	rightW := width - leftW - 2

	var leftSb strings.Builder
	leftSb.WriteString(component.GetStyle("panelTitle").Render(tui.DTUILogo))
	leftSb.WriteString("\n\n")
	leftSb.WriteString(component.GetStyle("dim").Render(strings.TrimSpace(aboutText)))

	// Right column: all shortcut sections
	var rightSb strings.Builder
	for _, sec := range r.provider.Sections(app) {
		rightSb.WriteString("\n")
		rightSb.WriteString(component.GetStyle("header").Render(sec.Title))
		rightSb.WriteString("\n")
		for _, item := range sec.Items {
			keyColumn := component.PadVisible(component.TruncateVisible(item.Keys, 20), 20)
			rightSb.WriteString(fmt.Sprintf("  %s  %s\n",
				component.GetStyle("helpKey").Render(keyColumn),
				component.GetStyle("helpDesc").Render(item.Desc),
			))
		}
	}
	rightSb.WriteString("\n")
	closeKeys := "Esc"
	for _, shortcut := range action.ForMode(app) {
		if shortcut.Description == "Close" || shortcut.Description == i18n.T("key.close_help") {
			closeKeys = shortcut.Key
			break
		}
	}
	rightSb.WriteString(component.GetStyle("dim").Render(i18n.T("help.press_close", closeKeys)))

	leftBox := lipgloss.NewStyle().Width(leftW).Background(style.Colors.BG).Render(leftSb.String())
	rightBox := lipgloss.NewStyle().Width(rightW).Background(style.Colors.BG).Render(rightSb.String())

	return lipgloss.JoinHorizontal(lipgloss.Top, leftBox, rightBox)
}

var defaultRenderer = NewRenderer(nil)

func RenderView(width int, app ...*state.AppModel) string {
	var model *state.AppModel
	if len(app) > 0 {
		model = app[0]
	}
	return defaultRenderer.Render(width, model)
}
