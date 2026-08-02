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
	dockerclient "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
	"github.com/elizabevil/docker-tui/internal/tui/ui/widget/actionbar"
	"github.com/elizabevil/docker-tui/internal/tui/ui/widget/dialog"
	"github.com/elizabevil/docker-tui/internal/tui/ui/widget/footer"
	"github.com/elizabevil/docker-tui/internal/tui/ui/widget/header"
	"github.com/elizabevil/docker-tui/internal/tui/ui/widget/panel"
	"github.com/elizabevil/docker-tui/internal/utils"
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
	if m.Viewport.Width == 0 || m.Viewport.Height == 0 {
		return i18n.T("msg.loading")
	}
	switch ClassifyTerminal(m.Viewport.Width, m.Viewport.Height) {
	case TerminalUnsupported:
		return renderTerminalError(m.Viewport.Width, m.Viewport.Height)
	case TerminalCompact:
		return renderActionBar(m, renderCompactApp(m))
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
	marginTop := m.Viewport.Height * mt / 100
	marginBot := m.Viewport.Height * mb / 100
	usableH := m.Viewport.Height - marginTop - marginBot
	if usableH < 10 {
		usableH = m.Viewport.Height
		marginTop = 0
	}

	contentWidthPct := appCfg.ContentWidthPct
	if contentWidthPct <= 0 || contentWidthPct > 100 {
		contentWidthPct = 90
	}
	usableW := m.Viewport.Width * contentWidthPct / 100
	if usableW < 50 {
		usableW = m.Viewport.Width
	}
	padH := (m.Viewport.Width - usableW) / 2

	rails := calculateRailHeights(usableH)
	totalH := rails.total()

	// Layer 1+2: compute global image colors with overlay
	bgCfg := m.Dependencies.Config.Layout.Background
	var globalColors []string
	if bgCfg.Enable {
		switch bgCfg.Type {
		case "image":
			if bgCfg.Image.Src != "" {
				globalColors = precomputeGlobalImageColors(m, totalH)
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

	overlayColor := m.Dependencies.Config.UI.DialogOverlayColor
	if overlayColor == "" {
		overlayColor = "#0d1117cc"
	}
	rep := ResolveLayout(m)
	panelBody := dialog.PanelBody{
		Left:  rep.Panel.bodyLeft,
		Top:   rep.Panel.bodyTop,
		Width: rep.Panel.bodyWidth,
		Rows:  rep.Panel.bodyRows,
	}
	if m.Navigation.Mode == state.ModeActionBar {
		return actionbar.RenderBar(m, result, panelBody)
	}
	if m.Navigation.Mode == state.ModeConfirm {
		return dialog.RenderChoiceOverlayInPanel(result, m, panelBody)
	}
	if m.Navigation.Mode == state.ModeExecShell {
		shellOverlay := component.RenderShellDialogBox(m.Dialog.Input.Text, panelBody.Width, overlayColor)
		return dialog.CenterOnPanelDefault(result, shellOverlay, panelBody, m.Viewport.Width, m.Viewport.Height)
	}
	if m.Navigation.Mode == state.ModeRename || m.Navigation.Mode == state.ModeResourceCreate ||
		m.Navigation.Mode == state.ModeImageWorkflow {
		inputOverlay := component.RenderTextInputBox(m.Dialog.Title, m.Dialog.Input.Text,
			m.Dialog.Input.Cursor, panelBody.Width, overlayColor)
		return dialog.CenterOnPanelDefault(result, inputOverlay, panelBody, m.Viewport.Width, m.Viewport.Height)
	}
	if m.Navigation.Mode == state.ModeImageTransfer {
		progressOverlay := component.RenderProgressDialogBox(m.Dialog.Title, m.Dialog.Body, imageTransferStatus(m.ImageTransfer.Progress.Status),
			m.ImageTransfer.Progress.Current, m.ImageTransfer.Progress.Total, panelBody.Width, overlayColor)
		return dialog.CenterOnPanelDefault(result, progressOverlay, panelBody, m.Viewport.Width, m.Viewport.Height)
	}
	if m.Navigation.Mode == state.ModeContainerForm {
		return dialog.RenderContainerFormOverlayInPanel(result, m, panelBody)
	}
	if m.Dialog.Kind.IsSelection() {
		return dialog.RenderOverlayInPanel(result, m, panelBody)
	}
	if m.Dialog.Kind == state.DialogExec {
		return dialog.RenderExecOverlayInPanel(result, m, panelBody)
	}
	if m.Navigation.Mode == state.ModeRuntimeSelect {
		return dialog.CenterOnPanelDefault(result, renderRuntimeSelector(m), panelBody, m.Viewport.Width, m.Viewport.Height)
	}
	return result
}

func renderActionBar(m *state.AppModel, content string) string {
	if m == nil || m.Navigation.Mode != state.ModeActionBar {
		return content
	}
	report := ResolveLayout(m)
	body := dialog.PanelBody{
		Left: report.Panel.bodyLeft, Top: report.Panel.bodyTop,
		Width: report.Panel.bodyWidth, Rows: report.Panel.bodyRows,
	}
	return actionbar.RenderBar(m, content, body)
}

func imageTransferStatus(status string) string {
	switch status {
	case "starting", "saving", "loading":
		return i18n.T("image.transfer." + status)
	default:
		return status
	}
}

type runtimeSelectorUI struct {
	Title   string
	Entries []runtimeSelectorEntry
	Footer  string
	Border  lipgloss.Border
	Padding [2]int
	Width   int
}

type runtimeSelectorEntry struct {
	Marker string
	Name   string
	Info   string
}

func (ui runtimeSelectorUI) Render() string {
	rows := []string{ui.Title, ""}
	for _, e := range ui.Entries {
		rows = append(rows, fmt.Sprintf("%s%s", e.Marker, e.Name))
		rows = append(rows, e.Info)
	}
	rows = append(rows, "", ui.Footer)
	return lipgloss.NewStyle().Border(ui.Border).Padding(ui.Padding[0], ui.Padding[1]).Width(ui.Width).Render(strings.Join(rows, "\n"))
}

func renderRuntimeSelector(m *state.AppModel) string {
	if m.Connection.Pool == nil {
		return "runtime selection unavailable"
	}
	hostNames := m.Connection.Pool.KnownHostNames()
	cursor := m.Connection.RuntimeSelectorCursor
	errors := m.Connection.RuntimeSelectorError

	entries := make([]runtimeSelectorEntry, len(hostNames))
	for i, name := range hostNames {
		e := m.Connection.Pool.Get(name)
		marker := "  "
		if i == cursor {
			marker = "> "
		}
		status := "disconnected"
		runtime := "-"
		version := "-"
		security := ""
		latency := "-"
		if e != nil && e.State == dockerclient.StateConnected {
			status = "connected"
			if e.Latency > 0 {
				latency = e.Latency.String()
			}
		}
		if e != nil && e.Engine != nil {
			identity := e.Engine.Identity()
			if identity.Version != "" {
				version = identity.Version
			}
		}
		if e != nil && e.Runtime != "" {
			runtime = string(e.Runtime)
		}
		if e != nil && e.TLS.Enabled {
			security = " " + i18n.T("connection.tls_configured")
			if e.State == dockerclient.StateConnected {
				securityKey := "connection.tls_verified"
				if e.TLS.InsecureSkipVerify {
					securityKey = "connection.tls_insecure"
				}
				security = " " + i18n.T(securityKey)
			}
		}
		if failure, ok := errors[name]; ok {
			status = "error: " + i18n.ConnectionFailureMessage(string(failure.Kind))
		}
		entries[i] = runtimeSelectorEntry{
			Marker: marker,
			Name:   name,
			Info: fmt.Sprintf("    [%s v%s @ %s] %s latency=%s%s",
				runtime, version, socketDisplay(e), status, latency, security),
		}
	}

	ui := runtimeSelectorUI{
		Title:   "Select runtime connection",
		Entries: entries,
		Footer:  "Enter connect  Esc cancel  R refresh",
		Border:  lipgloss.RoundedBorder(),
		Padding: [2]int{1, 2},
		Width:   100,
	}
	return ui.Render()
}

func socketDisplay(e *dockerclient.PoolEntry) string {
	if e == nil || e.Host == "" {
		return "-"
	}
	return e.Host
}

func renderMiddlePanel(m *state.AppModel, panelH int, panelW int) string {
	borderLabel := ""
	if !filterMode(m.Navigation.Mode) {
		if f := currentTableFilterLabel(m); f != "" {
			borderLabel = "Filter: " + f
		}
	}

	bodyH := panelBodyHeight(panelH)
	contentW := panelW - 4
	page := projectPage(m, bodyH, contentW)
	return panel.Panel{
		Title:       page.title,
		Info:        page.summary,
		Content:     page.content,
		Breadcrumb:  page.breadcrumb,
		BorderLabel: borderLabel,
		Width:       panelW,
		Height:      panelH,
	}.Render()
}

// filterMode reports whether the user is typing into a filter/input field
// that should hide the table border label.
func filterMode(mode state.AppMode) bool {
	switch mode {
	case state.ModeFilter, state.ModeSearch, state.ModeImagePull, state.ModeCommand:
		return true
	default:
		return false
	}
}

func renderExecPassthroughPanel(m *state.AppModel, bodyH int) string {
	var lines []string
	if m.Exec.ExecBuf != nil {
		lines = m.Exec.ExecBuf.View(bodyH, m.Exec.ExecScroll)
	}
	if len(lines) == 0 {
		lines = make([]string, bodyH)
		if bodyH >= 3 {
			lines[0] = "  Connecting to container shell..."
			lines[1] = ""
			lines[2] = "  Press Esc to return."
		}
	}
	panelW := m.Viewport.Width - 8
	if m.Exec.ExecBuf != nil {
		if vr, vc := m.Exec.ExecBuf.CursorVisible(bodyH, m.Exec.ExecScroll); vr >= 0 {
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
	switch m.Navigation.ActivePanel {
	case state.PanelContainers:
		if m.Resources.Containers != nil {
			return m.Resources.Containers.Filter
		}
	case state.PanelImages:
		if m.Resources.Images != nil {
			return m.Resources.Images.Filter
		}
	case state.PanelVolumes:
		if m.Resources.Volumes != nil {
			return m.Resources.Volumes.Filter
		}
	case state.PanelNetworks:
		if m.Resources.Networks != nil {
			return m.Resources.Networks.Filter
		}
	case state.PanelCompose:
		if m.Compose.ComposeFocus == 1 {
			if m.Compose.ComposeServiceFilter == "" {
				return ""
			}
			return "服务: " + m.Compose.ComposeServiceFilter
		}
		if m.Compose.ComposeProjectFilter == "" {
			return ""
		}
		return "项目: " + m.Compose.ComposeProjectFilter
	}
	return ""
}

// ── Background rendering ────────────────────────────────────

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
		return cached.([]color.Color), nil //nolint:errcheck // cache holds only []color.Color; assert cannot fail.
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open image %s: %w", path, err)
	}
	defer func() { _ = f.Close() }() //nolint:errcheck // image file fully decoded before close.
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
	bg := m.Dependencies.Config.Layout.Background
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

func blendColors(base, top string, topPct int) string { return utils.BlendColors(base, top, topPct) }

// breadcrumb derives the navigation path from the current model state.
// Used in the title bar area. Each page + sub-view combination produces a path.
func breadcrumb(m *state.AppModel) string {
	items := buildBreadcrumbItems(m)
	if len(items) == 0 {
		return ""
	}
	bw := m.Viewport.Width - 6
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
		Label: state.PanelLabel(m.Navigation.ActivePanel),
		ID:    fmt.Sprintf("%d", m.Navigation.ActivePanel),
	}}

	if m.Navigation.ActivePanel == state.PanelImages && m.Resources.Images != nil && m.Resources.Images.ContainersViewID != "" {
		items = append(items, component.BreadcrumbItem{Label: "containers", ID: "images-containers"})
	}
	if m.Navigation.ActivePanel == state.PanelVolumes && m.Resources.Volumes != nil && m.Resources.Volumes.DetailName != "" {
		items = append(items, component.BreadcrumbItem{Label: "containers", ID: "volumes-containers"})
	}
	if m.Navigation.ActivePanel == state.PanelCompose && m.Compose.ComposeContainerViewID != "" {
		items = append(items, component.BreadcrumbItem{Label: "containers", ID: "compose-containers"})
	}

	switch m.Navigation.Mode {
	case state.ModeLogView:
		items = append(items, component.BreadcrumbItem{Label: "logs", ID: "logs"})
	case state.ModeDetail:
		items = append(items, component.BreadcrumbItem{Label: "detail", ID: "detail"})
	case state.ModeTop:
		items = append(items, component.BreadcrumbItem{Label: "top", ID: "top"})
	case state.ModeExecPassthrough:
		items = append(items, component.BreadcrumbItem{Label: "exec", ID: "exec"})
	case state.ModeHelp:
		items = append(items, component.BreadcrumbItem{Label: "help", ID: "help"})
	case state.ModeHistory:
		items = append(items, component.BreadcrumbItem{Label: i18n.T("history.title"), ID: "history"})
	case state.ModeEvents:
		items = append(items, component.BreadcrumbItem{Label: i18n.T("events.title"), ID: "events"})
	}

	return items
}
