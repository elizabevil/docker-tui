package help

import (
	_ "embed"
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/elizabevil/docker-tui/internal/tui/ui/component"

	"github.com/elizabevil/docker-tui/internal/data/i18n"
	"github.com/elizabevil/docker-tui/internal/tui"
)

//go:embed help.jsonc
var helpDefaultData []byte

// helpConfig maps the JSONC structure for help panel defaults.
type helpConfig struct {
	AppDescription string `json:"appDescription"`
}

func (c *helpConfig) normalize() {
	if c.AppDescription == "" {
		c.AppDescription = "Docker & Podman TUI Manager"
	}
}

// DefaultHelpConfig returns default values parsed from the embedded help.jsonc.
func DefaultHelpConfig() helpConfig {
	loader := component.ConfigLoader[helpConfig]{
		RawData:  helpDefaultData,
		Fallback: helpConfig{AppDescription: "Docker & Podman TUI Manager"},
		Normalize: func(c *helpConfig) {
			c.normalize()
		},
	}
	return loader.Load()
}

//go:embed shortcuts.jsonc
var shortcutsData []byte

// shortcutsConfig maps the JSONC structure for help-panel shortcut sections.
type shortcutsConfig struct {
	Sections []sectionDef `json:"sections"`
}

type sectionDef struct {
	TitleI18n string    `json:"titleI18n"`
	Title     string    `json:"title"`
	Items     []itemDef `json:"items"`
}

type itemDef struct {
	Keys string `json:"keys"`
	Desc string `json:"desc"`
}

func (c *shortcutsConfig) normalize() {
	for i := range c.Sections {
		if c.Sections[i].Items == nil {
			c.Sections[i].Items = []itemDef{}
		}
	}
}

func loadShortcuts() shortcutsConfig {
	loader := component.ConfigLoader[shortcutsConfig]{
		RawData:  shortcutsData,
		Fallback: shortcutsConfig{},
		Normalize: func(c *shortcutsConfig) {
			c.normalize()
		},
	}
	return loader.Load()
}

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
	Sections() []helpSection
}

type defaultSectionProvider struct{}

func (defaultSectionProvider) Sections() []helpSection {
	cfg := loadShortcuts()
	sections := make([]helpSection, 0, len(cfg.Sections))
	for _, s := range cfg.Sections {
		title := s.Title
		if s.TitleI18n != "" {
			title = i18n.T(s.TitleI18n)
		}
		items := make([]helpItem, len(s.Items))
		for i, it := range s.Items {
			items[i] = helpItem{Keys: it.Keys, Desc: it.Desc}
		}
		sections = append(sections, helpSection{Title: title, Items: items})
	}
	return sections
}

// Renderer 绑定帮助页配置与渲染行为，减少散落函数。
type Renderer struct {
	cfg      helpConfig
	provider SectionProvider
}

func NewRenderer(provider SectionProvider) *Renderer {
	if provider == nil {
		provider = defaultSectionProvider{}
	}
	return &Renderer{cfg: DefaultHelpConfig(), provider: provider}
}

func (r *Renderer) Render(width int) string {
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
	for _, sec := range r.provider.Sections() {
		rightSb.WriteString("\n")
		rightSb.WriteString(component.GetStyle("header").Render(sec.Title))
		rightSb.WriteString("\n")
		for _, item := range sec.Items {
			rightSb.WriteString(fmt.Sprintf("  %s  %s\n",
				component.GetStyle("helpKey").Render(fmt.Sprintf("%-20s", item.Keys)),
				component.GetStyle("helpDesc").Render(item.Desc),
			))
		}
	}
	rightSb.WriteString("\n")
	rightSb.WriteString(component.GetStyle("dim").Render("Press '?' or Esc to close help"))

	leftBox := lipgloss.NewStyle().Width(leftW).Render(leftSb.String())
	rightBox := lipgloss.NewStyle().Width(rightW).Render(rightSb.String())

	return lipgloss.JoinHorizontal(lipgloss.Top, leftBox, rightBox)
}

var defaultRenderer = NewRenderer(nil)

func RenderView(width int) string {
	return defaultRenderer.Render(width)
}
