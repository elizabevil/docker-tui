package footer

import (
	_ "embed"
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/elizabevil/docker-tui/internal/tui/keys"

	"github.com/elizabevil/docker-tui/internal/data/i18n"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
)

//go:embed footer.jsonc
var footerDefaultData []byte

// footerConfig maps the JSONC structure for footer default values.
type footerConfig struct {
	MarkSymbol string `json:"markSymbol"`
}

func (c *footerConfig) normalize() {
	if c.MarkSymbol == "" {
		c.MarkSymbol = "☑"
	}
}

// DefaultFooterConfig returns default values parsed from the embedded footer.jsonc.
func DefaultFooterConfig() footerConfig {
	loader := component.ConfigLoader[footerConfig]{
		RawData:  footerDefaultData,
		Fallback: footerConfig{MarkSymbol: "☑"},
		Normalize: func(c *footerConfig) {
			c.normalize()
		},
	}
	return loader.Load()
}

var defaultCfg = DefaultFooterConfig()

// ShortcutProvider 抽象快捷键来源，统一模式/全局/页面特定差异。
type ShortcutProvider interface {
	ModeShortcuts(app *state.AppModel) []keyShortcut
	GlobalShortcuts() []keyShortcut
	PageShortcuts(panel state.PanelType, marked map[string]bool) []keyShortcut
}

type defaultShortcutProvider struct{}

func (defaultShortcutProvider) ModeShortcuts(app *state.AppModel) []keyShortcut {
	if app.Mode == state.ModeMark {
		return ks("Esc Cancel", "Space/Enter Toggle", "Ctrl+D Delete marked")
	}
	if app.Mode == state.ModeConfirm {
		return ks("y Confirm", "n Cancel")
	}
	if app.Mode == state.ModeLogView {
		return ks("Esc Back", "j/k Scroll", "PgUp/Dn Page", "g/G Top/Bot")
	}
	if app.Mode == state.ModeDetail {
		return ks("Esc/Enter Back", "j/k Scroll", "Space/PgDn Page")
	}
	if app.Mode == state.ModeHelp {
		return ks("? Close")
	}
	return nil
}

func (defaultShortcutProvider) GlobalShortcuts() []keyShortcut {
	return []keyShortcut{
		{keys.KJDown, i18n.T("key.down")},
		{keys.KKUp, i18n.T("key.up")},
		{keys.KTab, i18n.T("key.panel")},
		{keys.KeySlash, i18n.T("key.filter")},
		{keys.KeyQmark, i18n.T("key.help")},
		{keys.KeyHUpper, i18n.T("key.header")},
		{keys.KeyCUpper, i18n.T("key.connect")},
		{keys.KeyQ, i18n.T("key.quit")},
	}
}

func (defaultShortcutProvider) PageShortcuts(panel state.PanelType, marked map[string]bool) []keyShortcut {
	return pageShortcuts(panel, marked)
}

var shortcuts ShortcutProvider = defaultShortcutProvider{}

func StatusBar(app *state.AppModel) string {
	if app.Mode == state.ModeMark {
		status := fmt.Sprintf("%s Mark mode", component.GetStyle("toastWarning").Render("●"))
		return component.GetStyle("statusBar").Render(status)
	}
	engineLabel := app.RuntimeType
	if engineLabel == "" {
		engineLabel = "docker"
	}
	hostStr := ""
	if app.Docker != nil {
		hostStr = app.Docker.Host
	}
	status := fmt.Sprintf("%s %s", component.GetStyle("toastSuccess").Render("●"), engineLabel)
	if app.Connecting {
		status = fmt.Sprintf("%s connecting...", component.GetStyle("toastWarning").Render("○"))
	} else if !app.Connected {
		status = fmt.Sprintf("%s disconnected", component.GetStyle("toastError").Render("○"))
	}
	if hostStr != "" {
		status += " │ " + hostStr
	}
	if op := operationLogStatus(app); op != "" {
		status += " │ " + op
	}
	return component.GetStyle("statusBar").Render(status)
}

func Shortcuts(app *state.AppModel) string {
	// 统一操作日志行：固定占位一行，默认空，避免内容出现时 footer 高度跳变。
	logLine := component.GetStyle("shortcutBar").Render(OperationLogLine(app))

	// 关键节点：先处理模式级快捷键，再组合全局+页面特定。
	if modeHints := shortcuts.ModeShortcuts(app); len(modeHints) > 0 {
		return lipgloss.JoinVertical(lipgloss.Top, renderKeyShortcuts(modeHints), logLine)
	}

	// 双行快捷键栏：全局行 + 页面特定行
	globalRow := renderGlobalShortcuts()
	pageRow := renderPageShortcuts(app.ActivePanel, app.MarkedIDs)
	if pageRow == "" {
		return lipgloss.JoinVertical(lipgloss.Top, globalRow, logLine)
	}
	return lipgloss.JoinVertical(lipgloss.Top, globalRow, pageRow, logLine)
}

// renderGlobalShortcuts 返回全局通用快捷键（所有面板一致，不区分大小写）。
func renderGlobalShortcuts() string {
	hints := shortcuts.GlobalShortcuts()
	return renderKeyShortcuts(hints)
}

// renderPageShortcuts 返回当前面板的特定快捷键。
func renderPageShortcuts(p state.PanelType, marked map[string]bool) string {
	s := shortcuts.PageShortcuts(p, marked)
	return renderKeyShortcuts(s)
}

// renderKeyShortcuts 将 keyShortcut 列表渲染为一行。
func renderKeyShortcuts(kk []keyShortcut) string {
	if len(kk) == 0 {
		return ""
	}
	var parts []string
	for _, x := range kk {
		parts = append(parts, component.GetStyle("hintKey").Render(x.key)+component.GetStyle("hintDesc").Render(" "+x.desc))
	}
	return component.GetStyle("shortcutBar").Render(strings.Join(parts, component.GetStyle("hintSep").Render(" │ ")))
}

type keyShortcut struct{ key, desc string }

func ks(pairs ...string) []keyShortcut {
	var s []keyShortcut
	for _, p := range pairs {
		if i := strings.IndexByte(p, ' '); i > 0 {
			s = append(s, keyShortcut{p[:i], p[i+1:]})
		}
	}
	return s
}

func pageShortcuts(p state.PanelType, marked map[string]bool) []keyShortcut {
	var all []keyShortcut
	hasMarked := len(marked) > 0
	switch p {
	case state.PanelContainers:
		if hasMarked {
			all = append(all,
				keyShortcut{fmt.Sprintf("%s %s %d", keys.KSpace, defaultCfg.MarkSymbol, len(marked)), i18n.T("key.toggle")},
				keyShortcut{keys.KeyS, i18n.T("key.batch_start")}, keyShortcut{keys.KCtrlS, i18n.T("key.batch_stop")},
				keyShortcut{keys.KCtrlR, i18n.T("key.batch_restart")}, keyShortcut{keys.KCtrlD, i18n.T("key.batch_delete")})
		} else {
			all = append(all,
				keyShortcut{keys.KSpace, i18n.T("key.mark")}, keyShortcut{keys.KeyS, i18n.T("key.start")}, keyShortcut{keys.KCtrlS, i18n.T("key.stop")},
				keyShortcut{keys.KCtrlR, i18n.T("key.restart")}, keyShortcut{keys.KeyL, i18n.T("key.logs")}, keyShortcut{keys.KeyD, i18n.T("key.detail")},
				keyShortcut{keys.KCtrlD, i18n.T("key.delete")})
		}
	case state.PanelImages:
		if hasMarked {
			all = append(all,
				keyShortcut{fmt.Sprintf("%s %s %d", keys.KSpace, defaultCfg.MarkSymbol, len(marked)), i18n.T("key.toggle")},
				keyShortcut{keys.KCtrlP, i18n.T("key.batch_pull")}, keyShortcut{keys.KeyP, i18n.T("key.batch_prune")},
				keyShortcut{keys.KCtrlD, i18n.T("key.batch_delete")})
		} else {
			all = append(all,
				keyShortcut{keys.KSpace, i18n.T("key.mark")}, keyShortcut{keys.KRight, i18n.T("key.expand")}, keyShortcut{keys.KeyY, i18n.T("key.copy")},
				keyShortcut{keys.KeyD, i18n.T("key.detail")}, keyShortcut{keys.KCtrlB, i18n.T("key.debug")}, keyShortcut{keys.KCtrlE, i18n.T("key.export")},
				keyShortcut{keys.KCtrlP, i18n.T("key.pull")}, keyShortcut{keys.KeyP, i18n.T("key.prune")}, keyShortcut{keys.KeyO, i18n.T("key.sort")},
				keyShortcut{keys.KCtrlD, i18n.T("key.delete")})
		}
	case state.PanelVolumes, state.PanelNetworks:
		if hasMarked {
			all = append(all,
				keyShortcut{fmt.Sprintf("%s %s %d", keys.KSpace, defaultCfg.MarkSymbol, len(marked)), i18n.T("key.toggle")},
				keyShortcut{keys.KCtrlD, i18n.T("key.batch_delete")})
		} else {
			all = append(all,
				keyShortcut{keys.KSpace, i18n.T("key.mark")}, keyShortcut{keys.KEnter, i18n.T("key.expand")}, keyShortcut{keys.KCtrlD, i18n.T("key.delete")})
		}
	case state.PanelCompose:
		all = append(all,
			keyShortcut{keys.KeyS, i18n.T("key.start")}, keyShortcut{keys.KCtrlS, i18n.T("key.stop")},
			keyShortcut{keys.KeyL, i18n.T("key.logs")}, keyShortcut{keys.KCtrlD, i18n.T("key.down")},
			keyShortcut{keys.KRight, i18n.T("key.detail")}, keyShortcut{keys.KEnter, i18n.T("key.expand")})
	}
	return all
}
