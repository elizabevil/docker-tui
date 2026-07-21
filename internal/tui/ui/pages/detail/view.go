package detail

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/bytedance/sonic"
	"github.com/elizabevil/docker-tui/internal/data/docker"
	"github.com/elizabevil/docker-tui/internal/data/i18n"
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
	// Handle source view modes (YAML/JSON)
	if m.DetailSourceType == "json" || m.DetailSourceType == "yaml" {
		return renderSourceView(m, panelHeight)
	}

	content := m.ImageDetailContent
	var sections []detailSection

	// Priority: structured image data > raw JSON (container/network/volume) > raw text
	if m.ImageDetailData != nil {
		sections = buildImageDetailDataSections(m.ImageDetailData)
	} else if len(m.DetailRawJSON) > 0 {
		// Use resource type to select the appropriate build function
		var dockerSections []docker.DetailSection
		switch m.DetailResourceType {
		case state.ResourceNetwork:
			dockerSections = docker.BuildNetworkDetailSections(m.DetailRawJSON)
		case state.ResourceVolume:
			dockerSections = docker.BuildVolumeDetailSections(m.DetailRawJSON)
		default: // state.ResourceContainer or empty
			dockerSections = docker.BuildContainerDetailSections(m.DetailRawJSON)
		}
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

	return renderSections(m, sections, panelHeight)
}

// renderSections renders detail sections with styles and scroll support.
func renderSections(m *state.AppModel, sections []detailSection, panelHeight int) string {
	sectionStyle := component.GetStyle("detailSection")
	labelStyle := component.GetStyle("detailLabel")
	valueStyle := component.GetStyle("detailValue")
	dimStyle := component.GetStyle("detailDim")

	bodyLines := make([]string, 0, len(sections)*3)
	for _, sec := range sections {
		bodyLines = append(bodyLines, sectionStyle.Render(fmt.Sprintf("  \u2500\u2500 %s ", sec.Title)))
		if sec.Subtitle != "" {
			bodyLines = append(bodyLines, "    "+dimStyle.Render(sec.Subtitle))
		}

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
	offset := m.DetailState.VisibleOffset(len(bodyLines), bodyHeight)

	visible := bodyLines[offset:]
	if len(visible) > bodyHeight {
		visible = visible[:bodyHeight]
	}
	for len(visible) < bodyHeight {
		visible = append(visible, "")
	}
	body := lipgloss.NewStyle().Height(bodyHeight).MaxHeight(bodyHeight).Render(strings.Join(visible, "\n"))

	footer := fmt.Sprintf(" %d-%d/%d", offset+1, offset+len(visible), len(bodyLines))
	if m.DetailHint != "" {
		footer += " \u2502 " + m.DetailHint
	}
	return lipgloss.JoinVertical(lipgloss.Top, body, dimStyle.Render(footer))
}

// renderSourceView renders raw JSON as JSON or YAML source code.
func renderSourceView(m *state.AppModel, panelHeight int) string {
	dimStyle := component.GetStyle("detailDim")
	codeStyle := component.GetStyle("detailValue")

	var text string
	if m.DetailSourceType == "yaml" {
		var obj interface{}
		if err := sonic.Unmarshal(m.DetailRawJSON, &obj); err == nil {
			if yamlBytes, err := yaml.Marshal(obj); err == nil {
				text = string(yamlBytes)
			}
		}
		if text == "" {
			text = string(m.DetailRawJSON)
		}
	} else {
		var obj interface{}
		if err := sonic.Unmarshal(m.DetailRawJSON, &obj); err == nil {
			if jsonBytes, err := json.MarshalIndent(obj, "", "  "); err == nil {
				text = string(jsonBytes)
			}
		}
		if text == "" {
			text = string(m.DetailRawJSON)
		}
	}

	lines := strings.Split(text, "\n")
	bodyHeight := panelHeight - 1
	if bodyHeight < 3 {
		bodyHeight = 3
	}
	offset := m.DetailState.VisibleOffset(len(lines), bodyHeight)

	visible := lines[offset:]
	if len(visible) > bodyHeight {
		visible = visible[:bodyHeight]
	}
	for len(visible) < bodyHeight {
		visible = append(visible, "")
	}

	rendered := make([]string, len(visible))
	for i, ln := range visible {
		rendered[i] = codeStyle.Render(ln)
	}
	body := lipgloss.NewStyle().Height(bodyHeight).MaxHeight(bodyHeight).Render(strings.Join(rendered, "\n"))

	footer := fmt.Sprintf(" %d-%d/%d", offset+1, offset+len(visible), len(lines))
	if m.DetailHint != "" {
		footer += " \u2502 " + m.DetailHint
	}
	return lipgloss.JoinVertical(lipgloss.Top, body, dimStyle.Render(footer))
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
