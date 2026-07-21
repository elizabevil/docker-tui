package view

import (
	_ "embed"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"strings"
	"sync"

	"charm.land/lipgloss/v2"
	"github.com/elizabevil/docker-tui/internal/data/config"
	"github.com/elizabevil/docker-tui/internal/data/i18n"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
	"github.com/elizabevil/docker-tui/internal/tui/ui/widget/dialog"
	"github.com/elizabevil/docker-tui/internal/tui/ui/widget/footer"
	"github.com/elizabevil/docker-tui/internal/tui/ui/widget/header"
	"github.com/elizabevil/docker-tui/internal/tui/ui/widget/panel"
	"github.com/elizabevil/docker-tui/internal/tui/utils"
)

//go:embed app.jsonc
var appConfigData []byte

type appWindowConfig struct {
	MarginTopPct    int `json:"marginTopPct"`
	MarginBottomPct int `json:"marginBottomPct"`
	ContentWidthPct int `json:"contentWidthPct"`
}

var appCfg appWindowConfig

func init() {
	loader := component.ConfigLoader[appWindowConfig]{
		RawData: appConfigData,
		Fallback: appWindowConfig{
			MarginTopPct:    5,
			MarginBottomPct: 5,
			ContentWidthPct: 90,
		},
	}
	appCfg = loader.Load()
}

// imageColorCache caches per-row colors extracted from image files.
// Key is struct{path string; rows int} to avoid returning wrong-sized slices after resize.
var imageColorCache sync.Map

const (
	dialogModeExport = "export"
	dialogModeDebug  = "debug"
	dialogModeExec   = "exec"
)

// sectionBG resolves the effective background for a section by merging the global
// background config with any per-section override. Inherits from global when not overridden.
func sectionBG(global config.BackgroundConfig, section string) config.SectionBackground {
	sb := config.SectionBackground{
		Type:           global.Type,
		Color:          global.Color,
		StartColor:     global.StartColor,
		EndColor:       global.EndColor,
		Image:          global.Image,
		OverlayColor:   global.Overlay.Color,
		OverlayOpacity: global.Overlay.Opacity,
	}
	var ov *config.SectionBackground
	switch section {
	case "top":
		ov = global.Sections.Top
	case "middle":
		ov = global.Sections.Middle
	case "bottom":
		ov = global.Sections.Bottom
	}
	if ov != nil {
		if ov.Type != "" {
			sb.Type = ov.Type
		}
		if ov.Color != "" {
			sb.Color = ov.Color
		}
		if ov.StartColor != "" {
			sb.StartColor = ov.StartColor
		}
		if ov.EndColor != "" {
			sb.EndColor = ov.EndColor
		}
		if ov.Image.Src != "" {
			sb.Image = ov.Image
		}
		if ov.OverlayColor != "" {
			sb.OverlayColor = ov.OverlayColor
		}
		if ov.OverlayOpacity > 0 {
			sb.OverlayOpacity = ov.OverlayOpacity
		}
	}
	return sb
}

func RenderApp(m *state.AppModel) string {
	if m.Width == 0 || m.Height == 0 {
		spinner := m.Spinner
		if spinner != nil && spinner.Active() {
			return spinner.Render()
		}
		return i18n.T("msg.loading")
	}
	if m.Width < minimumTerminalWidth || m.Height < minimumTerminalHeight {
		return terminalSizeMessage(m.Width, m.Height)
	}

	// Window margin: percentage of terminal height for top/bottom spacing
	mt := appCfg.MarginTopPct
	if mt < 0 {
		mt = 0
	}
	if mt > 15 {
		mt = 15
	}
	mb := appCfg.MarginBottomPct
	if mb < 0 {
		mb = 0
	}
	if mb > 15 {
		mb = 15
	}
	marginTop := m.Height * mt / 100
	marginBot := m.Height * mb / 100
	usableH := m.Height - marginTop - marginBot
	if usableH < 10 {
		usableH = m.Height
		marginTop = 0
		marginBot = 0
	}

	contentWidthPct := appCfg.ContentWidthPct
	if contentWidthPct <= 0 || contentWidthPct > 100 {
		contentWidthPct = 90
	}
	usableW := m.Width * contentWidthPct / 100
	if usableW < 50 {
		usableW = m.Width
	}
	padH := (m.Width - usableW) / 2

	rails := calculateRailHeights(usableH)
	totalH := rails.total()

	// Layer 1+2: compute global image colors with overlay
	bgCfg := m.Config.Layout.Background
	var globalColors []string
	if bgCfg.Enable {
		switch bgCfg.Type {
		case "image":
			if bgCfg.Image.Src != "" {
				globalColors = precomputeGlobalImageColors(m, totalH)
				if globalColors != nil {
					for i, c := range globalColors {
						if bgCfg.Overlay.Color != "" && bgCfg.Overlay.Opacity > 0 {
							c = blendColors(c, bgCfg.Overlay.Color, bgCfg.Overlay.Opacity)
						}
						op := bgCfg.Image.Opacity
						if op <= 0 {
							op = 100
						}
						if op < 100 {
							c = blendColors("#0d1117", c, op)
						}
						globalColors[i] = c
					}
				}
			}
		case "solid":
			if bgCfg.Color != "" {
				globalColors = make([]string, totalH)
				for i := range globalColors {
					globalColors[i] = bgCfg.Color
				}
				if bgCfg.Overlay.Color != "" && bgCfg.Overlay.Opacity > 0 {
					for i, c := range globalColors {
						globalColors[i] = blendColors(c, bgCfg.Overlay.Color, bgCfg.Overlay.Opacity)
					}
				}
			}
		}
	}
	headerColors := sliceColors(globalColors, 0, rails.header)
	messageColors := sliceColors(globalColors, rails.header, rails.message)
	queryColors := sliceColors(globalColors, rails.header+rails.message, rails.query)
	panelColors := sliceColors(globalColors, rails.header+rails.message+rails.query, rails.panel)
	footerColors := sliceColors(globalColors, totalH-rails.footer, rails.footer)

	// Helper: wrap content with background colors (ANSI reset barrier)
	wrap := func(text string, colors []string) string {
		if globalColors == nil {
			return text
		}
		return renderContentLayer(text, colors)
	}

	padLeft := func(text string) string {
		if padH <= 0 {
			return text
		}
		lines := strings.Split(text, "\n")
		space := strings.Repeat(" ", padH)
		for i := range lines {
			lines[i] = space + lines[i]
		}
		return strings.Join(lines, "\n")
	}

	pad := strings.Repeat("\n", marginTop)

	headerRendered := fitRailHeight(header.Render(m, usableW), rails.header)
	messageRendered := fitRailHeight(renderMessageRail(m, usableW), rails.message)
	queryRendered := fitRailHeight(renderQueryRail(m, usableW), rails.query)
	panelRendered := fitRailHeight(renderMiddlePanel(m, rails.panel, usableW), rails.panel)
	footerRendered := fitRailHeight(footer.Render(m, usableW), rails.footer)

	result := pad + padLeft(lipgloss.JoinVertical(lipgloss.Top,
		wrap(headerRendered, headerColors),
		wrap(messageRendered, messageColors),
		wrap(queryRendered, queryColors),
		wrap(panelRendered, panelColors),
		wrap(footerRendered, footerColors)))

	overlayColor := m.Config.UI.DialogOverlayColor
	if overlayColor == "" {
		overlayColor = "#0d1117cc"
	}
	if m.Mode == state.ModeConfirm {
		return component.PlaceOverlay(m.Width, m.Height, component.RenderConfirmMsg(m.ConfirmMessage, m.ConfirmTarget, m.Width, m.Height, overlayColor), overlayColor)
	}
	if m.Mode == state.ModeExecShell {
		return component.PlaceOverlay(m.Width, m.Height, component.RenderShellDialog(m.FilterText, m.Width, m.Height, overlayColor), overlayColor)
	}
	if m.Mode == state.ModeExport {
		return dialog.RenderOverlay(result, m, dialogModeExport)
	}
	if m.Mode == state.ModeDebug {
		return dialog.RenderOverlay(result, m, dialogModeDebug)
	}
	if m.Mode == state.ModeExec {
		return dialog.RenderExecOverlay(result, m)
	}
	return result
}

func renderMiddlePanel(m *state.AppModel, panelH int, panelW int) string {
	borderLabel := ""
	if m.Mode != state.ModeFilter && m.Mode != state.ModeSearch && m.Mode != state.ModeImagePull && m.Mode != state.ModeCommand {
		if f := currentTableFilterLabel(m); f != "" {
			borderLabel = "Filter: " + f
		}
	}

	bodyH := panelBodyHeight(panelH)
	contentW := panelW - 4
	page := projectPage(m, bodyH, contentW)
	return panel.Panel{Title: page.title, Info: page.summary, Content: page.content, Breadcrumb: page.breadcrumb, BorderLabel: borderLabel, Width: panelW, Height: panelH}.Render()
}

func renderExecPassthroughPanel(m *state.AppModel, bodyH int) string {
	var lines []string
	if m.ExecBuf != nil {
		lines = m.ExecBuf.View(bodyH, m.ExecScroll)
	}
	if len(lines) == 0 {
		lines = make([]string, bodyH)
		if bodyH >= 3 {
			lines[0] = "  Connecting to container shell..."
			lines[1] = ""
			lines[2] = "  Press Esc to return."
		}
	}
	panelW := m.Width - 8
	if m.ExecBuf != nil {
		if vr, vc := m.ExecBuf.CursorVisible(bodyH, m.ExecScroll); vr >= 0 {
			runes := []rune(lines[vr])
			if vc > len(runes) {
				vc = len(runes)
			}
			before := string(runes[:vc])
			after := string(runes[vc:])
			lines[vr] = before + "\u2588" + after
		}
	}
	for i, line := range lines {
		if len(line) > panelW {
			lines[i] = line[:panelW]
		}
	}
	return lipgloss.NewStyle().Width(panelW).Render(strings.Join(lines, "\n"))
}

// sliceColors safely slices a colors array. Returns nil if colors is nil or bounds are invalid.
func sliceColors(colors []string, start, count int) []string {
	if colors == nil || start < 0 || count <= 0 || start+count > len(colors) {
		return nil
	}
	return colors[start : start+count]
}

func insertCursor(text string, cursor int) string {
	r := []rune(text)
	if cursor < 0 {
		cursor = 0
	}
	if cursor > len(r) {
		cursor = len(r)
	}
	return string(r[:cursor]) + "\u2588" + string(r[cursor:])
}

func currentTableFilterLabel(m *state.AppModel) string {
	if m == nil {
		return ""
	}
	switch m.ActivePanel {
	case state.PanelContainers:
		if m.Containers != nil {
			return m.Containers.Filter
		}
	case state.PanelImages:
		if m.Images != nil {
			return m.Images.Filter
		}
	case state.PanelVolumes:
		if m.Volumes != nil {
			return m.Volumes.Filter
		}
	case state.PanelNetworks:
		if m.Networks != nil {
			return m.Networks.Filter
		}
	case state.PanelCompose:
		if m.ComposeFocus == 1 {
			if m.ComposeServiceFilter == "" {
				return ""
			}
			return "服务: " + m.ComposeServiceFilter
		}
		if m.ComposeProjectFilter == "" {
			return ""
		}
		return "项目: " + m.ComposeProjectFilter
	}
	return ""
}

// ── Background rendering ────────────────────────────────────

func hexToRGB(hex string) (int, int, int) { return utils.HexToRGB(hex) }
func rgbToHex(r, g, b int) string         { return utils.RGBToHex(r, g, b) }
func clamp(v int) int {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return v
}
func interpolateColor(start, end string, t float64) string {
	return utils.InterpolateColor(start, end, t)
}

type imgCacheKey struct {
	path string
	rows int
}

// loadImageRowColors loads an image and extracts the average color per terminal row.
// sampleRate controls horizontal sampling (% of image width, 0-100).
// position controls vertical alignment (center/top/bottom).
// Results are cached by (path, rows) to handle terminal resize safely.
func loadImageRowColors(path string, targetRows int, sampleRate int, position string) ([]color.Color, error) {
	key := imgCacheKey{path: path, rows: targetRows}
	if cached, ok := imageColorCache.Load(key); ok {
		return cached.([]color.Color), nil
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open image %s: %w", path, err)
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("decode image %s: %w", path, err)
	}
	bounds := img.Bounds()
	imgW := bounds.Dx()
	imgH := bounds.Dy()

	// Horizontal sampling range: center-aligned strip, width = sampleRate% of image
	if sampleRate <= 0 {
		sampleRate = 1
	}
	if sampleRate > 100 {
		sampleRate = 100
	}
	sampleW := imgW * sampleRate / 100
	if sampleW < 1 {
		sampleW = 1
	}
	xStart := bounds.Min.X + (imgW-sampleW)/2
	xEnd := xStart + sampleW
	if xEnd > bounds.Max.X {
		xEnd = bounds.Max.X
	}

	// Vertical position mapping
	var srcY func(row, total int) int
	switch position {
	case "top":
		srcY = func(row, total int) int { return bounds.Min.Y + row*imgH/total/2 }
	case "bottom":
		srcY = func(row, total int) int { return bounds.Max.Y - 1 - (total-1-row)*imgH/total/2 }
	default: // center
		srcY = func(row, total int) int { return bounds.Min.Y + row*imgH/total }
	}

	colors := make([]color.Color, targetRows)
	for row := 0; row < targetRows; row++ {
		y := srcY(row, targetRows)
		if y > bounds.Max.Y-1 {
			y = bounds.Max.Y - 1
		}
		if y < bounds.Min.Y {
			y = bounds.Min.Y
		}
		// Average colors across the horizontal sample range
		var rSum, gSum, bSum int64
		count := 0
		for x := xStart; x < xEnd; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			rSum += int64(r >> 8)
			gSum += int64(g >> 8)
			bSum += int64(b >> 8)
			count++
		}
		if count == 0 {
			colors[row] = img.At(xStart, y)
		} else {
			// Return averaged color
			colors[row] = color.RGBA{uint8(rSum / int64(count)), uint8(gSum / int64(count)), uint8(bSum / int64(count)), 255}
		}
	}
	imageColorCache.Store(key, colors)
	return colors, nil
}

// precomputeGlobalImageColors loads the background image once and returns per-row
// hex colors for ALL terminal rows. Sections slice into this array for continuity.
func precomputeGlobalImageColors(m *state.AppModel, totalRows int) []string {
	bg := m.Config.Layout.Background
	if bg.Type != "image" || bg.Image.Src == "" {
		return nil
	}
	imgPath := bg.Image.Src
	pos := bg.Image.Position
	if pos == "" {
		pos = "center"
	}
	sampleRate := bg.Image.SampleRate
	if sampleRate <= 0 {
		sampleRate = 80 // high default for smooth gradient
	}
	imgColors, err := loadImageRowColors(imgPath, totalRows, sampleRate, pos)
	if err != nil {
		return nil
	}
	hex := make([]string, len(imgColors))
	for i, c := range imgColors {
		r, g, b, _ := c.RGBA()
		hex[i] = fmt.Sprintf("#%02x%02x%02x", r/257, g/257, b/257)
	}
	return hex
}

// renderContentLayer is Layer 3: overlays pre-styled content text on top of
// pre-computed background row colors (which already include Layer 1 + Layer 2 blending).
// Uses the ANSI reset barrier: every \033[0m in content is followed by a background
// restore code, preventing theme style resets from clearing the image background.
// When globalColors is empty, returns the text unchanged (passthrough).
func renderContentLayer(text string, rowColors []string) string {
	if text == "" || len(rowColors) == 0 {
		return text
	}
	lines := strings.Split(text, "\n")
	for i := range lines {
		if i >= len(rowColors) {
			break
		}
		br, bgC, bb := utils.HexToRGB(rowColors[i])
		bgAnsi := fmt.Sprintf("\033[48;2;%d;%d;%dm", br, bgC, bb)
		line := strings.ReplaceAll(lines[i], "\033[0m", "\033[0m"+bgAnsi)
		lines[i] = bgAnsi + line + "\033[0m"
	}
	return strings.Join(lines, "\n")
}

func resolveOverlay(color string, opacity int) string {
	if color == "" || opacity <= 0 {
		return ""
	}
	return color
}

func blendColors(base, top string, topPct int) string { return utils.BlendColors(base, top, topPct) }

// breadcrumb derives the navigation path from the current model state.
// Used in the title bar area. Each page + sub-view combination produces a path.
func breadcrumb(m *state.AppModel) string {
	items := buildBreadcrumbItems(m)
	if len(items) == 0 {
		return ""
	}
	bw := m.Width - 6
	if bw < 10 {
		bw = 10
	}
	return component.RenderBreadcrumb(items, " > ", bw)
}

func buildBreadcrumbItems(m *state.AppModel) []component.BreadcrumbItem {
	if m == nil {
		return nil
	}
	items := []component.BreadcrumbItem{{
		Label: state.PanelLabel(m.ActivePanel),
		ID:    fmt.Sprintf("%d", m.ActivePanel),
	}}

	if m.ActivePanel == state.PanelImages && m.Images != nil && m.Images.ContainersViewID != "" {
		items = append(items, component.BreadcrumbItem{Label: "containers", ID: "images-containers"})
	}
	if m.ActivePanel == state.PanelVolumes && m.Volumes != nil && m.Volumes.DetailName != "" {
		items = append(items, component.BreadcrumbItem{Label: "containers", ID: "volumes-containers"})
	}
	if m.ActivePanel == state.PanelCompose && m.ComposeContainerViewID != "" {
		items = append(items, component.BreadcrumbItem{Label: "containers", ID: "compose-containers"})
	}

	switch m.Mode {
	case state.ModeLogView:
		items = append(items, component.BreadcrumbItem{Label: "logs", ID: "logs"})
	case state.ModeDetail:
		items = append(items, component.BreadcrumbItem{Label: "detail", ID: "detail"})
	case state.ModeExecPassthrough:
		items = append(items, component.BreadcrumbItem{Label: "exec", ID: "exec"})
	case state.ModeHelp:
		items = append(items, component.BreadcrumbItem{Label: "help", ID: "help"})
	}

	return items
}
