package detail

import (
	_ "embed"
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/elizabevil/docker-tui/internal/data/i18n"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
)

//go:embed detail.jsonc
var detailDefaultData []byte

// detailConfig maps the JSONC structure for detail view defaults.
type detailConfig struct {
	FoldKey string `json:"foldKey"`
}

func (c *detailConfig) normalize() {
	if c.FoldKey == "" {
		c.FoldKey = "space"
	}
}

// DefaultDetailConfig returns default values parsed from the embedded detail.jsonc.
func DefaultDetailConfig() detailConfig {
	loader := component.ConfigLoader[detailConfig]{
		RawData:  detailDefaultData,
		Fallback: detailConfig{FoldKey: "space"},
		Normalize: func(c *detailConfig) {
			c.normalize()
		},
	}
	return loader.Load()
}

type detailSection struct {
	Title string
	Lines []string
}

func RenderView(m *state.AppModel, panelHeight int) string {
	content := m.ImageDetailContent
	if content == "" {
		content = i18n.T("hint.loading_detail")
	}

	sections := buildDetailSections(content)
	if len(sections) == 0 {
		sections = []detailSection{{Title: "Details", Lines: []string{content}}}
	}

	// 柔和配色样式 — 护眼设计，无背景高亮
	sectionStyle := component.GetStyle("detailSection")
	labelStyle := component.GetStyle("detailLabel")
	valueStyle := component.GetStyle("detailValue")
	dimStyle := component.GetStyle("detailDim")

	bodyLines := make([]string, 0, len(sections)*3)
	for _, sec := range sections {
		// 区段标题 — 永远展开，无折叠指示器
		bodyLines = append(bodyLines, sectionStyle.Render(fmt.Sprintf("  \u2500\u2500 %s ", sec.Title)))

		for _, ln := range sec.Lines {
			trimmed := strings.TrimSpace(ln)
			if trimmed == "" {
				continue
			}
			if idx := strings.Index(trimmed, ":"); idx > 0 {
				label := strings.TrimSpace(trimmed[:idx])
				val := strings.TrimSpace(trimmed[idx+1:])
				bodyLines = append(bodyLines, fmt.Sprintf("    %s: %s",
					labelStyle.Render(label), valueStyle.Render(val)))
			} else {
				bodyLines = append(bodyLines, "    "+dimStyle.Render(ln))
			}
		}
	}

	bodyHeight := panelHeight - 1
	if bodyHeight < 3 {
		bodyHeight = 3
	}
	if m.DetailOffset < 0 {
		m.DetailOffset = 0
	}
	maxOffset := len(bodyLines) - bodyHeight
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.DetailOffset > maxOffset {
		m.DetailOffset = maxOffset
	}

	visible := bodyLines[m.DetailOffset:]
	if len(visible) > bodyHeight {
		visible = visible[:bodyHeight]
	}
	for len(visible) < bodyHeight {
		visible = append(visible, "")
	}
	body := lipgloss.NewStyle().Height(bodyHeight).MaxHeight(bodyHeight).Render(strings.Join(visible, "\n"))

	footer := fmt.Sprintf(" %d-%d/%d", m.DetailOffset+1, m.DetailOffset+len(visible), len(bodyLines))
	if m.DetailHint != "" {
		footer += " \u2502 " + m.DetailHint
	}
	return lipgloss.JoinVertical(lipgloss.Top, body, dimStyle.Render(footer))
}

func buildDetailSections(content string) []detailSection {
	if isImageDetailContent(content) {
		return buildImageDetailSections(content)
	}

	lines := strings.Split(content, "\n")
	sections := make([]detailSection, 0, 8)
	current := detailSection{Title: "Summary", Lines: make([]string, 0, 16)}

	flush := func() {
		if current.Title == "" && len(current.Lines) == 0 {
			return
		}
		if current.Title == "" {
			current.Title = "Summary"
		}
		sections = append(sections, current)
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "\u2500\u2500") && strings.HasSuffix(trimmed, "\u2500\u2500") {
			flush()
			title := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(trimmed, "\u2500\u2500"), "\u2500\u2500"))
			current = detailSection{Title: title, Lines: make([]string, 0, 16)}
			continue
		}
		current.Lines = append(current.Lines, line)
	}
	flush()

	filtered := make([]detailSection, 0, len(sections))
	for _, sec := range sections {
		nonEmpty := false
		for _, line := range sec.Lines {
			if strings.TrimSpace(line) != "" {
				nonEmpty = true
				break
			}
		}
		if nonEmpty {
			filtered = append(filtered, sec)
		}
	}
	return filtered
}
