package detail

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/bytedance/sonic"
	"github.com/elizabevil/docker-tui/internal/data/i18n"
	"github.com/elizabevil/docker-tui/internal/data/runtime/docker"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
	"gopkg.in/yaml.v3"
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

// detailSection represents a collapsible section in the detail view.
type detailSection struct {
	Title    string
	Subtitle string
	Lines    []string
}

// RenderView renders the detail panel content based on the current resource type.
func RenderView(m *state.AppModel, panelHeight int) string {
	lines := detailDocument(m)
	return renderDocument(m, lines, panelHeight)
}

func detailDocument(m *state.AppModel) []state.DetailDocumentLine {
	if cached, ok := m.Detail.Documents[m.Detail.DetailSourceType]; ok &&
		cached.Lines != nil &&
		cached.Revision == m.Detail.Revision {
		return cached.Lines
	}

	var lines []state.DetailDocumentLine
	if m.Detail.DetailSourceType == state.DetailSourceJSON ||
		m.Detail.DetailSourceType == state.DetailSourceYAML {
		lines = buildSourceDocument(m)
	} else {
		lines = buildSectionDocument(m)
	}
	if m.Detail.Documents == nil {
		m.Detail.Documents = make(map[state.DetailSource]state.DetailDocument, 3)
	}
	m.Detail.Documents[m.Detail.DetailSourceType] = state.DetailDocument{
		Revision: m.Detail.Revision,
		Source:   m.Detail.DetailSourceType,
		Lines:    lines,
	}
	return lines
}

func buildSectionDocument(m *state.AppModel) []state.DetailDocumentLine {
	content := m.Detail.ImageDetailContent
	var sections []detailSection

	// Structured domain data takes precedence over fallback text/source views.
	if m.Detail.ImageDetailData != nil {
		sections = buildImageDetailDataSections(m.Detail.ImageDetailData)
	} else if m.Detail.VolumeDetail != nil {
		sections = convertDockerSections(docker.BuildVolumeDetailSections(m.Detail.VolumeDetail))
	} else if m.Detail.NetworkDetail != nil {
		sections = convertDockerSections(docker.BuildNetworkDetailSections(m.Detail.NetworkDetail))
	} else if m.Detail.ContainerDetail != nil {
		dockerSections := docker.BuildContainerDetailSections(m.Detail.ContainerDetail)
		sections = convertDockerSections(dockerSections)
	} else {
		if content == "" {
			content = i18n.T("hint.loading_detail")
		}
		sections = buildDetailSections(content)
	}
	if len(sections) == 0 {
		sections = []detailSection{{Title: "Details", Lines: []string{content}}}
	}

	return flattenSections(sections)
}

func flattenSections(sections []detailSection) []state.DetailDocumentLine {
	bodyLines := make([]state.DetailDocumentLine, 0, len(sections)*3)
	for _, sec := range sections {
		bodyLines = append(bodyLines, state.DetailDocumentLine{Kind: state.DetailLineSection, Left: sec.Title})
		if sec.Subtitle != "" {
			bodyLines = append(bodyLines, state.DetailDocumentLine{Kind: state.DetailLineSubtitle, Left: sec.Subtitle})
		}

		for _, ln := range sec.Lines {
			trimmed := strings.TrimSpace(ln)
			if trimmed == "" {
				continue
			}
			if idx := strings.Index(trimmed, ":"); idx > 0 {
				label := strings.TrimSpace(trimmed[:idx])
				val := strings.TrimSpace(trimmed[idx+1:])
				bodyLines = append(bodyLines, state.DetailDocumentLine{Kind: state.DetailLineValue, Left: label, Right: val})
			} else {
				bodyLines = append(bodyLines, state.DetailDocumentLine{Kind: state.DetailLinePlain, Left: ln})
			}
		}
	}
	return bodyLines
}

func renderDocument(m *state.AppModel, bodyLines []state.DetailDocumentLine, panelHeight int) string {
	sectionStyle := component.GetStyle("detailSection")
	labelStyle := component.GetStyle("detailLabel")
	valueStyle := component.GetStyle("detailValue")
	dimStyle := component.GetStyle("detailDim")

	bodyHeight := panelHeight - 1
	if bodyHeight < 3 {
		bodyHeight = 3
	}
	offset := m.Detail.ClampVisibleOffset(len(bodyLines), bodyHeight)

	visible := bodyLines[offset:]
	if len(visible) > bodyHeight {
		visible = visible[:bodyHeight]
	}
	visibleEnd := offset + len(visible)
	rendered := make([]string, 0, bodyHeight)
	for _, line := range visible {
		switch line.Kind {
		case state.DetailLineSection:
			rendered = append(rendered, sectionStyle.Render(fmt.Sprintf("  \u2500\u2500 %s ", line.Left)))
		case state.DetailLineSubtitle:
			rendered = append(rendered, "    "+dimStyle.Render(line.Left))
		case state.DetailLineValue:
			rendered = append(rendered, fmt.Sprintf("    %s: %s",
				labelStyle.Render(line.Left), valueStyle.Render(line.Right)))
		default:
			if m.Detail.DetailSourceType == state.DetailSourceSection {
				rendered = append(rendered, "    "+dimStyle.Render(line.Left))
			} else {
				rendered = append(rendered, valueStyle.Render(line.Left))
			}
		}
	}
	for len(rendered) < bodyHeight {
		rendered = append(rendered, "")
	}
	body := lipgloss.NewStyle().Height(bodyHeight).MaxHeight(bodyHeight).Render(strings.Join(rendered, "\n"))

	footer := fmt.Sprintf(" %d-%d/%d", offset+1, visibleEnd, len(bodyLines))
	if m.Detail.DetailHint != "" {
		footer += " \u2502 " + m.Detail.DetailHint
	}
	return lipgloss.JoinVertical(lipgloss.Top, body, dimStyle.Render(footer))
}

// buildSourceDocument formats raw data once per detail revision and source.
func buildSourceDocument(m *state.AppModel) []state.DetailDocumentLine {
	if !m.Detail.HasRawSource() {
		return []state.DetailDocumentLine{{
			Kind: state.DetailLinePlain,
			Left: i18n.T("hint.source_unavailable"),
		}}
	}
	var text string
	if m.Detail.DetailSourceType == state.DetailSourceYAML {
		var obj interface{}
		if err := sonic.Unmarshal(m.Detail.DetailRawJSON, &obj); err == nil {
			if yamlBytes, err := yaml.Marshal(obj); err == nil {
				text = string(yamlBytes)
			}
		}
		if text == "" {
			text = string(m.Detail.DetailRawJSON)
		}
	} else {
		var obj interface{}
		if err := sonic.Unmarshal(m.Detail.DetailRawJSON, &obj); err == nil {
			if jsonBytes, err := json.MarshalIndent(obj, "", "  "); err == nil {
				text = string(jsonBytes)
			}
		}
		if text == "" {
			text = string(m.Detail.DetailRawJSON)
		}
	}

	lines := strings.Split(text, "\n")
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		lines = []string{i18n.T("hint.source_unavailable")}
	}
	document := make([]state.DetailDocumentLine, len(lines))
	for i, line := range lines {
		document[i] = state.DetailDocumentLine{Kind: state.DetailLinePlain, Left: line}
	}
	return document
}

// convertDockerSections converts docker.DetailSection to detailSection.
func convertDockerSections(dockerSections []docker.DetailSection) []detailSection {
	if len(dockerSections) == 0 {
		return nil
	}
	result := make([]detailSection, len(dockerSections))
	for i, ds := range dockerSections {
		result[i] = detailSection{
			Title:    ds.Title,
			Subtitle: ds.Subtitle,
			Lines:    ds.Lines,
		}
	}
	return result
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
