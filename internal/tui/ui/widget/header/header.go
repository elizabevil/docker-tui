package header

import (
	"fmt"
	"image/color"
	"strings"
	"time"

	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
	"github.com/elizabevil/docker-tui/internal/tui/ui/style"

	"charm.land/lipgloss/v2"

	"github.com/elizabevil/docker-tui/internal/data/config"
	"github.com/elizabevil/docker-tui/internal/data/i18n"
	"github.com/elizabevil/docker-tui/internal/tui"
	"github.com/elizabevil/docker-tui/internal/tui/state"
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
	cpuC := cpuLoadColor(app.Metrics.HostCPU)
	lb := component.GetStyle("headerLabel")
	tz := currentTimezone()
	lang := i18n.Current()
	if lang == "" {
		lang = "en"
	}

	// Column helpers
	lbl := func(s string) string { return lb.Render(utils.PadVisible(s, 10)) }

	// ── Col 1: Host dynamic info (15%) ────────────────────────
	pct := func(v float64, c color.Color) string {
		return lipgloss.NewStyle().Foreground(c).Render(utils.PadVisible(utils.FormatPercent(v), 8))
	}
	memStr := fmt.Sprintf("%s/%s",
		utils.FormatBytes(float64(app.Metrics.HostMemUsed)),
		utils.FormatBytes(float64(app.Metrics.HostMemTotal)))
	colDyn := fmt.Sprintf("%s%s %dC\n%s%s\n%s%s\n%s%s",
		lbl("CPU"), pct(app.Metrics.HostCPU, cpuC), app.Metrics.HostCPUCores,
		lbl("Memory"), utils.PadVisible(memStr, 18),
		lbl("Disk"), utils.PadVisible(app.Metrics.HostDisk, 18),
		lbl("TimeZone"), utils.PadVisible(tz, 18),
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
	colConn := fmt.Sprintf("%s%s\n%s%s\n%s%s",
		lbl("Engine"), eng+" "+app.Connection.EngineVersion,
		lbl("Socket"), hostStr,
		lbl("Language"), lang,
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
	logoW := widths.logo
	if logoW < 10 {
		logoW = 10
	}

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
		component.GetStyle("panelTitle").Render(logoPart),
		component.GetStyle("panelTitle").Render("v"+verStr),
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

	return component.GetStyle("headerBar").Width(usableW).Render(rendered)
}

// renderKeyStrokeColumn 显示快捷键日志（简化版，仅收集期间显示）。
func renderKeyStrokeColumn(app *state.AppModel, colW, ratio int) string {
	var content string

	switch {
	case len(app.Feedback.KeyStrokeBuffer) > 0:
		content = joinKeyBadges(app.Feedback.KeyStrokeBuffer)
		content = component.GetStyle("keyBadge").Background(style.Colors.Blue).Padding(0, 1).Bold(true).Render(content)
	case len(app.Feedback.LastKeyStroke) > 0:
		content = joinKeyBadges(app.Feedback.LastKeyStroke)
		content = component.GetStyle("keyLast").Render(content)
	default:
		return "" // 无按键时不留空白
	}

	padH := (colW - 2) * (100 - ratio) / 200
	if padH < 0 {
		padH = 0
	}

	boxed := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true, true, true, true).
		BorderForeground(style.Colors.Blue).
		Padding(0, padH).
		Render(content)
	boxLines := strings.Count(boxed, "\n") + 1
	if boxLines < 4 {
		boxed += strings.Repeat("\n", 4-boxLines)
	}
	return boxed
}

// joinKeyBadges renders a slice of KeyStrokeEvents as badge strings.
// Each badge looks like "[R] Restart".
func joinKeyBadges(events []state.KeyStrokeEvent) string {
	var parts []string
	for _, evt := range events {
		parts = append(parts, fmt.Sprintf("[%s] %s", evt.Key, evt.Action))
	}
	return strings.Join(parts, "  ")
}

func cpuLoadColor(pct float64) color.Color {
	switch {
	case pct > 80:
		return style.Colors.Red
	case pct > 50:
		return style.Colors.Yellow
	default:
		return style.Colors.Green
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
