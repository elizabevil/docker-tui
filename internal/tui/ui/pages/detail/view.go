package detail

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/elizabevil/docker-tui/internal/data/i18n"
	"github.com/elizabevil/docker-tui/internal/data/runtime/docker"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
	"github.com/elizabevil/docker-tui/internal/utils"
)

// detailSection represents a collapsible section in the detail view.
type detailSection struct {
	Title    string
	Subtitle string
	Lines    []string
}

// RenderView renders the detail panel content based on the current resource type.
func RenderView(m *state.AppModel, panelHeight, panelWidth int) string {
	lines := detailDocument(m, panelWidth)
	return renderDocument(m, lines, panelHeight, panelWidth)
}

func detailDocument(m *state.AppModel, panelWidth int) []state.DetailDocumentLine {
	if cached, ok := m.Detail.Documents[m.Detail.DetailSourceType]; ok &&
		cached.Lines != nil &&
		cached.Revision == m.Detail.Revision && cached.Width == panelWidth {
		return cached.Lines
	}

	var lines []state.DetailDocumentLine
	if m.Detail.DetailSourceType == state.DetailSourceJSON ||
		m.Detail.DetailSourceType == state.DetailSourceYAML {
		lines = buildSourceDocument(m)
	} else {
		lines = buildSectionDocument(m)
	}
	lines = wrapDocument(lines, panelWidth, m.Detail.DetailSourceType != state.DetailSourceSection)
	if m.Detail.Documents == nil {
		m.Detail.Documents = make(map[state.DetailSource]state.DetailDocument, 3)
	}
	m.Detail.Documents[m.Detail.DetailSourceType] = state.DetailDocument{
		Revision: m.Detail.Revision,
		Source:   m.Detail.DetailSourceType,
		Width:    panelWidth,
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
		sections = []detailSection{{Title: i18n.T("inspect.image.details"), Lines: []string{content}}}
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

func renderDocument(m *state.AppModel, bodyLines []state.DetailDocumentLine, panelHeight, panelWidth int) string {
	sectionStyle := component.GetStyle("detailSection")
	labelStyle := component.GetStyle("detailLabel")
	valueStyle := component.GetStyle("detailValue")
	dimStyle := component.GetStyle("detailDim")
	selectionStyle := component.GetStyle("detailSelection")

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
		var renderedLine string
		switch line.Kind {
		case state.DetailLineSection:
			renderedLine = sectionStyle.Render(fmt.Sprintf("  \u2500\u2500 %s ", line.Left))
		case state.DetailLineSubtitle:
			renderedLine = "    " + dimStyle.Render(line.Left)
		case state.DetailLineValue:
			renderedLine = fmt.Sprintf("    %s: %s", labelStyle.Render(line.Left), valueStyle.Render(line.Right))
		case state.DetailLineContinuation:
			renderedLine = "    " + valueStyle.Render(line.Left)
		default:
			if m.Detail.DetailSourceType == state.DetailSourceSection {
				renderedLine = "    " + dimStyle.Render(line.Left)
			} else {
				renderedLine = valueStyle.Render(line.Left)
			}
		}
		if m.Detail.SourceSelected && m.Detail.DetailSourceType != state.DetailSourceSection {
			renderedLine = selectionStyle.Width(max(1, panelWidth)).Render(renderedLine)
		}
		rendered = append(rendered, renderedLine)
	}
	for len(rendered) < bodyHeight {
		rendered = append(rendered, "")
	}
	body := lipgloss.NewStyle().Height(bodyHeight).MaxHeight(bodyHeight).Render(strings.Join(rendered, "\n"))

	footer := fmt.Sprintf(" %d-%d/%d", offset+1, visibleEnd, len(bodyLines))
	if m.Detail.DetailHint != "" {
		footer += " \u2502 " + m.Detail.DetailHint
	}
	if m.Detail.SourceSelected {
		footer += " \u2502 " + i18n.T("hint.source_selected")
	}
	return lipgloss.JoinVertical(lipgloss.Top, body, dimStyle.Render(footer))
}

func wrapDocument(lines []state.DetailDocumentLine, panelWidth int, source bool) []state.DetailDocumentLine {
	if panelWidth <= 0 {
		return lines
	}
	wrapped := make([]state.DetailDocumentLine, 0, len(lines))
	for _, line := range lines {
		switch {
		case line.Kind == state.DetailLineValue:
			parts := utils.WrapCells(line.Right, max(1, panelWidth-6-utils.DisplayWidth(line.Left)))
			line.Right = parts[0]
			wrapped = append(wrapped, line)
			for _, part := range parts[1:] {
				for _, continuation := range utils.WrapCells(part, max(1, panelWidth-4)) {
					wrapped = append(wrapped, state.DetailDocumentLine{Kind: state.DetailLineContinuation, Left: continuation})
				}
			}
		case source && line.Kind == state.DetailLinePlain:
			for _, part := range utils.WrapCells(line.Left, panelWidth) {
				wrapped = append(wrapped, state.DetailDocumentLine{Kind: state.DetailLinePlain, Left: part})
			}
		case line.Kind == state.DetailLinePlain:
			for _, part := range utils.WrapCells(line.Left, max(1, panelWidth-4)) {
				wrapped = append(wrapped, state.DetailDocumentLine{Kind: state.DetailLinePlain, Left: part})
			}
		default:
			wrapped = append(wrapped, line)
		}
	}
	return wrapped
}

// buildSourceDocument formats raw data once per detail revision and source.
func buildSourceDocument(m *state.AppModel) []state.DetailDocumentLine {
	if !m.Detail.HasRawSource() {
		return []state.DetailDocumentLine{{
			Kind: state.DetailLinePlain,
			Left: i18n.T("hint.source_unavailable"),
		}}
	}
	text := m.Detail.SourceText()

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
	current := detailSection{Title: i18n.T("inspect.image.summary"), Lines: make([]string, 0, 16)}

	flush := func() {
		if current.Title == "" && len(current.Lines) == 0 {
			return
		}
		if current.Title == "" {
			current.Title = i18n.T("inspect.image.summary")
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
