package header

import (
	"fmt"
	"image/color"
	"strings"
	"time"

	"charm.land/lipgloss/v2"

	"github.com/elizabevil/docker-tui/internal/data/config"
	"github.com/elizabevil/docker-tui/internal/data/i18n"
	"github.com/elizabevil/docker-tui/internal/tui"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component/box"
	"github.com/elizabevil/docker-tui/internal/tui/ui/style"
	"github.com/elizabevil/docker-tui/internal/utils"
)

type headerWidths struct {
	host       int
	connection int
	keystroke  int
	logo       int
}

func resolveHeaderWidths(avail int, weights config.HeaderColumnWeights) headerWidths {
	total := weights.Host + weights.Connection + weights.Keystroke + weights.Logo
	host := avail * weights.Host / total
	connection := avail * weights.Connection / total
	keystroke := avail * weights.Keystroke / total
	return headerWidths{
		host:       host,
		connection: connection,
		keystroke:  keystroke,
		logo:       avail - host - connection - keystroke,
	}
}

// RenderHeight returns the number of terminal rows the header occupies.
// The header format is fixed at 4 rows (3 line breaks).
func RenderHeight() int { return 4 }

func Render(app *state.AppModel, usableW int) string {
	if !app.Viewport.HeaderVisible || app.Viewport.Width < 50 {
		return ""
	}

	eng := ""
	if app.Connection.Pool != nil {
		eng = app.Connection.Pool.ActiveName()
	}
	if eng == "" {
		eng = app.Connection.RuntimeType
	}
	if eng == "" {
		eng = app.Connection.ConnectionTarget
	}
	if eng == "" {
		eng = "no runtime"
	}
	hostStr := ""
	if app.Connection.Engine != nil {
		hostStr = app.Connection.Engine.Identity().Endpoint
	}
	var linkStatus string
	if app.Connection.Connected {
		linkStatus = component.LinkUp
	} else if app.Connection.Connecting {
		linkStatus = "..."
	} else {
		linkStatus = "—"
	}
	latencyStr := "—"
	if app.Connection.Pool != nil {
		if entry := app.Connection.Pool.Active(); entry != nil && entry.Latency > 0 {
			latencyStr = fmt.Sprintf("%dms", entry.Latency.Milliseconds())
		}
	}
	cpuC := cpuLoadColor(app.Metrics.HostCPU)
	headerBg := resolveHeaderBackground(app)
	tz := currentTimezone()
	lang := i18n.Current()
	if lang == "" {
		lang = "en"
	}

	// Column helpers — each label / value / stat cell is a self-contained
	// component with the header background applied, so padding spaces inside
	// later Width/Border sub-boxes pick up the header fill instead of the
	// terminal default.
	lbl := func(s string) string {
		return (&box.LabeledValue{
			Label: utils.PadVisible(s, 10), LabelStyle: box.StyleHeaderLabel,
			Background: headerBg,
		}).Render()
	}
	val := func(s string) string {
		return (&box.LabeledValue{
			Value: utils.PadVisible(s, 18), ValueStyle: box.StyleHeaderValue,
			Background: headerBg,
		}).Render()
	}

	// ── Col 1: Host dynamic info (15%) ────────────────────────
	pct := func(v float64, c color.Color) string {
		return (&box.Stat{
			Value: v, Unit: "%", ValueWidth: 8,
			Color: hexFromColor(c), Background: headerBg,
			WarningAt: 50, DangerAt: 80,
		}).Render()
	}
	memStr := fmt.Sprintf("%s/%s",
		utils.FormatBytes(float64(app.Metrics.HostMemUsed)),
		utils.FormatBytes(float64(app.Metrics.HostMemTotal)))
	colDyn := fmt.Sprintf("%s%s %dC\n%s%s\n%s%s\n%s%s",
		lbl("CPU"), pct(app.Metrics.HostCPU, cpuC), app.Metrics.HostCPUCores,
		lbl("Memory"), val(memStr),
		lbl("Disk"), val(app.Metrics.HostDisk),
		lbl("TimeZone"), val(tz),
	)

	// ── Col 2: Connection + App config (25%) ──────────────────
	// (resolved `eng` at col 1 carries the runtime name; TLS state is appended here.)
	if app.Connection.Pool != nil {
		if active := app.Connection.Pool.Active(); active != nil && active.TLS.Enabled {
			securityKey := "connection.tls_verified"
			if active.TLS.InsecureSkipVerify {
				securityKey = "connection.tls_insecure"
			}
			eng += " [" + i18n.T(securityKey) + "]"
		}
	}
	// Link status and latency share a single row so the user reads the
	// transport health (● / ○ / ...) and its round-trip cost together as
	// one cohesive "is this connection usable right now?" signal, instead
	// of as two unrelated stacked rows. Padding is tightened so the
	// 2:2:3:3 column ratios still leave room for both values once width
	// is truncated by narrow terminals.
	linkLatency := renderLinkCompact(linkStatus) + " " + renderLatencyCompact(latencyStr)
	// Four rows so col 2 lines up with col 1's host-metrics rows (CPU /
	// Memory / Disk / TimeZone). The Engine entry is split from the
	// version so the version has its own slot and the tail of "Engine"
	// never gets truncated mid-name.
	colConn := fmt.Sprintf("%s%s\n%s%s\n%s%s\n%s%s",
		lbl("Engine"), val(eng),
		lbl("Version"), val(app.Connection.EngineVersion),
		lbl("Socket"), val(hostStr),
		lbl("Link"), linkLatency,
	)

	// ── Column widths from config weights ─────────────────────
	avail := usableW - 4
	if avail < 10 {
		avail = app.Viewport.Width - 4
	}
	hCfg := app.Dependencies.Config.UI.Header
	widths := resolveHeaderWidths(avail, hCfg.Columns)
	dynW := widths.host
	connW := widths.connection
	keyW := widths.keystroke
	logoW := max(widths.logo, 10)

	// ── Col 3: Keystroke display (简化版，仅显示按键日志，无动画) ──
	colKeys := renderKeyStrokeColumn(app, keyW, hCfg.KeystrokeContentRatio)

	// ── Col 4: Logo ───────────────────────────────────────────
	verStr := app.Dependencies.AppVersion
	if verStr == "" {
		verStr = "dev"
	}
	logoLines := strings.Split(strings.TrimSpace(tui.DTUILogo), "\n")
	logoPart := strings.Join(logoLines[:min(3, len(logoLines))], "\n")
	colLogo := lipgloss.JoinVertical(lipgloss.Right,
		component.GetStyle(component.StylePanelTitle).Render(logoPart),
		component.GetStyle(component.StylePanelTitle).Render("v"+verStr),
	)
	fitColumn := func(content string, width int) string {
		lines := strings.Split(content, "\n")
		for i := range lines {
			lines[i] = utils.FitVisible(lines[i], width)
		}
		return strings.Join(lines, "\n")
	}

	rendered := lipgloss.JoinHorizontal(lipgloss.Top,
		fitColumn(colDyn, dynW),
		fitColumn(colConn, connW),
		fitColumn(colKeys, keyW),
		lipgloss.NewStyle().Width(logoW).Align(lipgloss.Right).Render(fitColumn(colLogo, logoW)),
	)

	headerStyle := lipgloss.NewStyle().Width(usableW)
	return headerStyle.Render(rendered)
}

// renderLinkCompact renders the link indicator without the 5-char padding
// that renderLink uses, so it can sit alongside the latency value on the
// same row inside the narrow Connection column. The glyph is still styled
// with the success / danger foreground.
func renderLinkCompact(status string) string {
	switch status {
	case component.LinkUp:
		return style.ApplyForeground(lipgloss.NewStyle(), style.Colors.Success).Render(status)
	case component.BulletEmpty:
		return style.ApplyForeground(lipgloss.NewStyle(), style.Colors.Danger).Render(status)
	default:
		return lipgloss.NewStyle().Render(status)
	}
}

// renderLatencyCompact renders the latency value with no surrounding padding
// so it sits next to the link indicator without expanding the Connection
// row past the column width.
func renderLatencyCompact(latency string) string {
	return component.GetStyle(box.StyleHeaderValue).Render(latency)
}

// renderLink 渲染 link 状态字符: ● 绿色(Success) / ○ 红色(Danger) / 文字(中性)
func renderLink(status string) string {
	switch status {
	case component.LinkUp:
		up := style.ApplyForeground(lipgloss.NewStyle(), style.Colors.Success)
		return component.GetStyle(component.StyleHeaderBar).Render(utils.PadVisible(
			up.Render(status), 5))
	case component.BulletEmpty:
		empty := style.ApplyForeground(lipgloss.NewStyle(), style.Colors.Danger)
		return component.GetStyle(component.StyleHeaderBar).Render(utils.PadVisible(
			empty.Render(status), 5))
	default:
		return component.GetStyle(component.StyleHeaderBar).Render(utils.PadVisible(status, 5))
	}
}

// renderKeyStrokeColumn 显示快捷键日志（简化版，仅收集期间显示）。
func renderKeyStrokeColumn(app *state.AppModel, colW, ratio int) string {
	var keyStyleName box.StyleName
	var bold bool
	var events []state.KeyStrokeEvent

	switch {
	case len(app.Feedback.KeyStrokeBuffer) > 0:
		keyStyleName = box.StyleHeaderKey
		bold = true
		events = app.Feedback.KeyStrokeBuffer
	case len(app.Feedback.LastKeyStroke) > 0:
		keyStyleName = box.StyleHeaderLast
		events = app.Feedback.LastKeyStroke
	default:
		return "" // 无按键时不留空白
	}

	badges := make([]box.Badge, len(events))
	for i, evt := range events {
		badges[i] = box.Badge{Key: evt.Key, Description: evt.Action, StyleName: keyStyleName, Bold: bold}
	}
	headerBg := resolveHeaderBackground(app)
	content := (&box.BadgeRow{
		Badges:     badges,
		Separator:  "  ",
		Background: headerBg,
	}).Render()

	// Horizontal padding between the rounded border and the first badge.
	// The KeystrokeContentRatio still drives the relative "fill" aim, but
	// the floor is bumped so the badges don't visually stick to the border
	// on wider terminals. The 0.6 ceiling keeps the box from hollowing out
	// on the narrowest layouts.
	padH := max((colW-2)*(100-ratio)/100+2, 3)
	if maxPad := (colW - 2) * 60 / 100; maxPad > 0 && padH > maxPad {
		padH = maxPad
	}

	boxed := (&box.BorderedBox{
		Content:     content,
		UseRounded:  true,
		BorderColor: hexFromColor(component.GetStyle(box.StylePanelTitle).GetForeground()),
		Background:  headerBg,
		Padding:     [2]int{0, padH},
		Width:       colW,
	}).Render()
	boxLines := strings.Count(boxed, "\n") + 1
	if boxLines < 4 {
		boxed += strings.Repeat("\n", 4-boxLines)
	}
	return boxed
}

// resolveHeaderBackground returns the header's component-level background,
// or "" when the header should be transparent. The transparent value is the
// fallback: the global renderAppBackground wrapper provides the fill, and
// any component that explicitly sets a Background overrides it.
func resolveHeaderBackground(app *state.AppModel) string {
	return ""
}

// hexFromColor returns the hex string for a color.Color produced by the
// style package, or "" when the value is nil / the transparent token.
func hexFromColor(c color.Color) string {
	if c == nil {
		return ""
	}
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("#%02x%02x%02x", r/257, g/257, b/257)
}

func cpuLoadColor(pct float64) color.Color {
	switch {
	case pct > 80:
		return style.Colors.Danger
	case pct > 50:
		return style.Colors.Warning
	default:
		return style.Colors.Success
	}
}

func currentTimezone() string {
	loc := time.Now().Location()
	name := loc.String()
	if name != "" && name != "Local" {
		return name
	}
	_, offset := time.Now().Zone()
	sign := "+"
	if offset < 0 {
		sign = "-"
		offset = -offset
	}
	return fmt.Sprintf("UTC%s%02d:%02d", sign, offset/3600, (offset%3600)/60)
}
