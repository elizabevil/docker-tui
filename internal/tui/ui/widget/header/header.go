package header

import (
	_ "embed"
	"fmt"
	"image/color"
	"strings"
	"time"

	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
	"github.com/elizabevil/docker-tui/internal/tui/ui/style"

	"charm.land/lipgloss/v2"

	"github.com/elizabevil/docker-tui/internal/data/i18n"
	"github.com/elizabevil/docker-tui/internal/tui"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/utils"
)

//go:embed header.jsonc
var headerDefaultData []byte

// ColumnDef defines one column in the header bar.
type ColumnDef struct {
	ID     string `json:"id"`
	Weight int    `json:"weight"`
}

// keystrokeSubConfig holds keystroke-column-specific settings.
type keystrokeSubConfig struct {
	ContentRatio    int `json:"contentRatio"`    // percentage of column width for content (default 80)
	DisplayDuration int `json:"displayDuration"` // 显示时长 (ticks, 默认 30 = 3s)
	AnimDuration    int `json:"animDuration"`    // 动画时长 (ticks, 默认 5 = 500ms)
}

// headerConfig maps the JSONC structure for header defaults.
type headerConfig struct {
	DefaultTitle string              `json:"defaultTitle"`
	Columns      []ColumnDef         `json:"columns"`
	Keystroke    *keystrokeSubConfig `json:"keystroke,omitempty"`
}

func (c *headerConfig) normalize() {
	if c.DefaultTitle == "" {
		c.DefaultTitle = "dtui"
	}
	if len(c.Columns) == 0 {
		c.Columns = []ColumnDef{
			{ID: "host", Weight: 3},
			{ID: "connection", Weight: 5},
			{ID: "keystroke", Weight: 7},
			{ID: "logo", Weight: 3},
		}
	}
	if c.Keystroke == nil {
		c.Keystroke = &keystrokeSubConfig{ContentRatio: 80, DisplayDuration: 30, AnimDuration: 5}
	}
	if c.Keystroke.ContentRatio <= 0 || c.Keystroke.ContentRatio > 100 {
		c.Keystroke.ContentRatio = 80
	}
	if c.Keystroke.DisplayDuration <= 0 {
		c.Keystroke.DisplayDuration = 30
	}
	if c.Keystroke.AnimDuration <= 0 {
		c.Keystroke.AnimDuration = 5
	}
}

// DefaultHeaderConfig returns default values parsed from the embedded header.jsonc.
func DefaultHeaderConfig() headerConfig {
	loader := component.ConfigLoader[headerConfig]{
		RawData: headerDefaultData,
		Fallback: headerConfig{
			DefaultTitle: "dtui",
			Columns: []ColumnDef{
				{ID: "host", Weight: 3},
				{ID: "connection", Weight: 5},
				{ID: "keystroke", Weight: 7},
				{ID: "logo", Weight: 3},
			},
		},
		Normalize: func(c *headerConfig) {
			c.normalize()
		},
	}
	return loader.Load()
}

// widthByID computes column widths from config weights.
func widthByID(avail int, cols []ColumnDef) map[string]int {
	totalW := 0
	for _, c := range cols {
		totalW += c.Weight
	}
	out := make(map[string]int, len(cols))
	used := 0
	for i, c := range cols {
		w := avail * c.Weight / totalW
		if i == len(cols)-1 {
			w = avail - used // remainder absorbs rounding
		}
		out[c.ID] = w
		used += w
	}
	return out
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
	if app.Connection.Docker != nil {
		hostStr = app.Connection.Docker.Host
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
	colDyn := fmt.Sprintf("%s%s %dC\n%s%s\n%s%s\n%s%s\n",
		lbl("CPU"), pct(app.Metrics.HostCPU, cpuC), app.Metrics.HostCPUCores,
		lbl("Memory"), utils.PadVisible(memStr, 18),
		lbl("Disk"), utils.PadVisible(app.Metrics.HostDisk, 18),
		lbl("TimeZone"), utils.PadVisible(tz, 18),
	)

	// ── Col 2: Connection + App config (25%) ──────────────────
	runtimeName := ""
	if app.Connection.Pool != nil {
		runtimeName = app.Connection.Pool.ActiveName()
	}
	if runtimeName == "" {
		runtimeName = app.Connection.ConnectionTarget
	}
	if runtimeName == "" {
		runtimeName = "-"
	}
	if app.Connection.Pool != nil {
		if active := app.Connection.Pool.Active(); active != nil && active.TLS.Enabled {
			securityKey := "connection.tls_verified"
			if active.TLS.InsecureSkipVerify {
				securityKey = "connection.tls_insecure"
			}
			runtimeName += " [" + i18n.T(securityKey) + "]"
		}
	}
	colConn := fmt.Sprintf("%s%s\n%s%s\n%s%s\n%s%s\n",
		lbl("Engine"), eng+" "+app.Connection.EngineVersion,
		lbl("Runtime"), runtimeName,
		lbl("Socket"), hostStr,
		lbl("Language"), lang,
	)

	// ── Column widths from config weights ─────────────────────
	avail := usableW - 4
	if avail < 10 {
		avail = app.Viewport.Width - 4
	}
	hCfg := DefaultHeaderConfig()
	widths := widthByID(avail, hCfg.Columns)
	dynW := widths["host"]
	connW := widths["connection"]
	keyW := widths["keystroke"]
	logoW := widths["logo"]
	if logoW < 10 {
		logoW = 10
	}

	// ── Col 3: Keystroke display (简化版，仅显示按键日志，无动画) ──
	colKeys := renderKeyStrokeColumn(app, keyW)

	// ── Col 4: Logo ───────────────────────────────────────────
	verStr := app.Dependencies.AppVersion
	if verStr == "" {
		verStr = "dev"
	}
	logoLines := strings.Split(strings.TrimSpace(tui.DTUILogo), "\n")
	logoPart := strings.Join(logoLines[:min(3, len(logoLines))], "\n")
	colLogo := lipgloss.JoinVertical(lipgloss.Right,
		component.GetStyle("panelTitle").Render(logoPart),
		component.GetStyle("dim").Render("v"+verStr),
	)

	rendered := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(dynW).Render(colDyn),
		lipgloss.NewStyle().Width(connW).Render(colConn),
		lipgloss.NewStyle().Width(keyW).Render(colKeys),
		lipgloss.NewStyle().Width(logoW).Align(lipgloss.Right).Render(colLogo),
	)

	return component.GetStyle("headerBar").Render(rendered)
}

// spinnerBorders cycles corner patterns for a subtle spinning box effect.
var spinnerBorders = []lipgloss.Border{
	{Top: "─", Bottom: "─", Left: "│", Right: "│", TopLeft: "╭", TopRight: "╮", BottomLeft: "╰", BottomRight: "╯"},
	{Top: "─", Bottom: "─", Left: "│", Right: "│", TopLeft: "╰", TopRight: "╮", BottomLeft: "╭", BottomRight: "╯"},
	{Top: "─", Bottom: "─", Left: "│", Right: "│", TopLeft: "╰", TopRight: "╯", BottomLeft: "╭", BottomRight: "╮"},
	{Top: "─", Bottom: "─", Left: "│", Right: "│", TopLeft: "╭", TopRight: "╯", BottomLeft: "╰", BottomRight: "╮"},
}

// renderKeyStrokeColumn 显示快捷键日志（简化版，仅收集期间显示）。
func renderKeyStrokeColumn(app *state.AppModel, colW int) string {
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

	ratio := 80
	if hCfg := DefaultHeaderConfig(); hCfg.Keystroke != nil {
		ratio = hCfg.Keystroke.ContentRatio
	}
	if ratio <= 0 || ratio > 100 {
		ratio = 80
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
