package component

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/elizabevil/docker-tui/internal/data/config"
	"github.com/elizabevil/docker-tui/internal/tui/ui/style"
	"github.com/elizabevil/docker-tui/internal/utils"
)

type globalStyleRefs struct {
	SearchBar            styleRef `json:"searchBar"`
	SearchCursor         styleRef `json:"searchCursor"`
	SearchHint           styleRef `json:"searchHint"`
	CommandPrefix        styleRef `json:"commandPrefix"`
	Breadcrumb           styleRef `json:"breadcrumb"`
	BreadcrumbActive     styleRef `json:"breadcrumbActive"`
	ToastSuccess         styleRef `json:"toastSuccess"`
	ToastError           styleRef `json:"toastError"`
	ToastInfo            styleRef `json:"toastInfo"`
	ToastWarning         styleRef `json:"toastWarning"`
	HintKey              styleRef `json:"hintKey"`
	HintDescription      styleRef `json:"hintDesc"`
	HintSeparator        styleRef `json:"hintSep"`
	DialogTitle          styleRef `json:"dialogTitle"`
	DialogBody           styleRef `json:"dialogBody"`
	DialogOption         styleRef `json:"dialogOption"`
	DialogOptionDisabled styleRef `json:"dialogOptionDisabled"`
	DialogBodyBackground styleRef `json:"dialogBodyBackground"`
	Panel                styleRef `json:"panel"`
	PanelTitle           styleRef `json:"panelTitle"`
	Dim                  styleRef `json:"dim"`
	HelpKey              styleRef `json:"helpKey"`
	HelpDescription      styleRef `json:"helpDesc"`
	DetailSection        styleRef `json:"detailSection"`
	DetailLabel          styleRef `json:"detailLabel"`
	DetailValue          styleRef `json:"detailValue"`
	DetailDim            styleRef `json:"detailDim"`
	DetailSelection      styleRef `json:"detailSelection"`
	LogTimestamp         styleRef `json:"logTimestamp"`
	LogText              styleRef `json:"logText"`
	LogStderr            styleRef `json:"logStderr"`
	LogHighlight         styleRef `json:"logHighlightBg"`
	Header               styleRef `json:"header"`
	HeaderBar            styleRef `json:"headerBar"`
	HeaderLabel          styleRef `json:"headerLabel"`
	KeyBadge             styleRef `json:"keyBadge"`
	KeyLast              styleRef `json:"keyLast"`
	Footer               styleRef `json:"footer"`
	ShortcutBar          styleRef `json:"shortcutBar"`
	ActionBar            styleRef `json:"actionBar"`
	MessageRail          styleRef `json:"messageRail"`
	QueryBar             styleRef `json:"queryBar"`
	SelectedRow          styleRef `json:"selectedRow"`
	DialogConfirm        styleRef `json:"dialogConfirm"`
	DialogError          styleRef `json:"dialogError"`
	DialogWarning        styleRef `json:"dialogWarning"`
}

// styleRef 定义 JSONC 中单一样式条目的属性。
// 字段与 styles.jsonc 中的 "styles" 对象一一对应。
type styleRef struct {
	Color      string `json:"color,omitempty"`
	Background string `json:"background,omitempty"`
	Bold       bool   `json:"bold,omitempty"`
	Faint      bool   `json:"faint,omitempty"`
}

// rawStyles 存储从 styles.jsonc 加载的原始样式引用。
// 延迟解析以避开 init() 顺序依赖（style.Colors 在 ApplyTheme 中填充）。
var rawStyles globalStyleRefs

func loadComponentStyles() {
	rawStyles = globalStyleRefs{
		SearchCursor:  styleRef{Bold: true},
		SearchHint:    styleRef{Faint: true},
		Breadcrumb:    styleRef{Faint: true},
		HintKey:       styleRef{Bold: true},
		DialogTitle:   styleRef{Bold: true},
		Dim:           styleRef{Faint: true},
		DetailSection: styleRef{Faint: true},
		DetailLabel:   styleRef{Faint: true},
		DetailValue:   styleRef{Faint: true},
		DetailDim:     styleRef{Faint: true},
		LogTimestamp:  styleRef{Faint: true},
		LogText:       styleRef{Faint: true},
		LogStderr:     styleRef{Faint: true},
		Header:        styleRef{Bold: true},
		KeyBadge:      styleRef{Bold: true},
		SelectedRow:   styleRef{Bold: true},
	}
	ApplyThemeStyles(config.DefaultTheme())
}

// ApplyThemeStyles projects fixed Theme scopes into the component rendering
// styles. The assignments are explicit so a theme cannot introduce arbitrary
// selectors or component names.
func ApplyThemeStyles(theme *config.Theme) {
	if theme == nil {
		return
	}
	resolve := theme.ResolveColor
	resolveBackground := func(ref config.ColorRef) string {
		return flattenThemeBackground(theme, ref)
	}
	rawStyles.SearchBar.Color = resolve(theme.Text.HelpDescription)
	rawStyles.SearchCursor.Color = resolve(theme.Text.HelpDescription)
	rawStyles.SearchHint.Color = resolve(theme.Text.Dim)
	rawStyles.CommandPrefix.Color = resolve(theme.Text.Success)
	rawStyles.Breadcrumb.Color = resolve(theme.Text.Dim)
	rawStyles.BreadcrumbActive.Color = resolve(theme.Text.Info)
	rawStyles.ToastSuccess = styleRef{Color: resolve(theme.Toast.Success), Background: resolveBackground(theme.Toast.Background)}
	rawStyles.ToastError = styleRef{Color: resolve(theme.Toast.Error), Background: resolveBackground(theme.Toast.Background)}
	rawStyles.ToastInfo = styleRef{Color: resolve(theme.Text.Info), Background: resolveBackground(theme.Toast.Background)}
	rawStyles.ToastWarning = styleRef{Color: resolve(theme.Text.Warning), Background: resolveBackground(theme.Toast.Background)}
	rawStyles.HintKey.Color = resolve(theme.Footer.Key)
	rawStyles.HintDescription.Color = resolve(theme.Footer.Description)
	rawStyles.HintSeparator.Color = resolve(theme.Footer.Separator)
	rawStyles.DialogTitle.Color = resolve(theme.Dialog.Title)
	rawStyles.DialogBody.Color = resolve(theme.Dialog.Body)
	rawStyles.DialogBodyBackground = styleRef{Background: resolveBackground(theme.Dialog.BodyBackground)}
	rawStyles.DialogOption.Color = resolve(theme.Dialog.OptionActive)
	rawStyles.DialogOptionDisabled.Color = resolve(theme.Dialog.OptionInactive)
	rawStyles.Panel = styleRef{Background: resolveBackground(theme.Main.PanelBackground)}
	rawStyles.PanelTitle.Color = resolve(theme.Main.Title)
	rawStyles.Dim.Color = resolve(theme.Text.Dim)
	rawStyles.HelpKey.Color = resolve(theme.Text.HelpKey)
	rawStyles.HelpDescription.Color = resolve(theme.Text.HelpDescription)
	rawStyles.DetailLabel.Color = resolve(theme.Text.Info)
	rawStyles.DetailSection.Color = resolve(theme.Text.HelpDescription)
	rawStyles.DetailValue.Color = resolve(theme.Text.HelpDescription)
	rawStyles.DetailDim.Color = resolve(theme.Text.Dim)
	rawStyles.DetailSelection = styleRef{Color: resolve(theme.Main.RowText), Background: resolveBackground(theme.Main.RowSelected), Bold: true}
	rawStyles.LogTimestamp.Color = resolve(theme.Text.Info)
	rawStyles.LogText.Color = resolve(theme.Text.HelpDescription)
	rawStyles.LogStderr.Color = resolve(theme.Text.Error)
	rawStyles.LogHighlight = styleRef{Color: resolveBackground(theme.Header.Background), Background: resolveBackground(theme.Text.Warning)}
	rawStyles.Header = styleRef{Color: resolve(theme.Main.TableHeader), Bold: true}
	rawStyles.HeaderBar = styleRef{Color: resolve(theme.Header.Value), Background: resolveBackground(theme.Header.Background)}
	rawStyles.HeaderLabel = styleRef{Color: resolve(theme.Header.Label)}
	rawStyles.KeyBadge = styleRef{Color: resolve(theme.Header.Value), Background: resolveBackground(theme.Header.Background), Bold: true}
	rawStyles.Footer = styleRef{Color: resolve(theme.Main.Footer)}
	rawStyles.ShortcutBar = styleRef{Background: resolveBackground(theme.Footer.ShortcutBackground)}
	rawStyles.ActionBar = styleRef{Background: resolveBackground(theme.Main.ActionBarBackground)}
	rawStyles.MessageRail = styleRef{Background: resolveBackground(theme.Main.MessageRailBackground)}
	rawStyles.QueryBar = styleRef{Background: resolveBackground(theme.Main.QueryBarBackground)}
	rawStyles.SelectedRow = styleRef{Color: resolve(theme.Main.RowText), Background: resolveBackground(theme.Main.RowSelected)}
	rawStyles.DialogConfirm = styleRef{Color: resolve(theme.Text.Success), Background: resolveBackground(theme.Toast.Background), Bold: true}
	rawStyles.DialogError = styleRef{Color: resolve(theme.Text.Error), Background: resolveBackground(theme.Toast.Background)}
	rawStyles.DialogWarning = styleRef{Color: resolve(theme.Text.Warning), Background: resolveBackground(theme.Toast.Background)}

	tableCfg.RowStyles.Selected.Color = resolve(theme.Main.RowText)
	tableCfg.RowStyles.Selected.Background = resolveBackground(theme.Main.RowSelected)
	tableCfg.RowStyles.Selected.Bold = true

	ApplyTableTheme(theme)
	ApplySafeFallback(theme)
}

func flattenThemeBackground(theme *config.Theme, ref config.ColorRef) string {
	value := theme.ResolveColor(ref)
	if value == string(config.FallbackColorTransparent) {
		return ""
	}
	parsed, ok := utils.ParseColor(value)
	if !ok {
		return value
	}
	overlay, ok := color.NRGBAModel.Convert(parsed).(color.NRGBA)
	if !ok || overlay.A == 0xff {
		return value
	}
	if overlay.A == 0 {
		return ""
	}
	baseParsed, ok := utils.ParseColor(string(theme.Palette.Background))
	if !ok {
		return value
	}
	base, ok := color.NRGBAModel.Convert(baseParsed).(color.NRGBA)
	if !ok {
		return value
	}
	blend := func(bottom, top uint8) uint8 {
		alpha := uint32(overlay.A)
		return uint8((uint32(top)*alpha + uint32(bottom)*(255-alpha) + 127) / 255)
	}
	return fmt.Sprintf("#%02x%02x%02x", blend(base.R, overlay.R), blend(base.G, overlay.G), blend(base.B, overlay.B))
}

func ApplyTableTheme(theme *config.Theme) {
	if theme == nil {
		return
	}
	resolve := theme.ResolveColor
	tableCfg.RowStyles.Marked.Background = flattenThemeBackground(theme, theme.Table.MarkedBackground)
	tableCfg.RowStyles.Marked.Color = resolve(theme.Table.NameForeground)
	tableCfg.ColumnStyles = defaultColumnStyles(resolve(theme.Table.ColumnForeground), resolve(theme.Table.NameForeground))
}

type SafeFallbackStyles struct {
	Normal lipgloss.Style
	Bold   lipgloss.Style
	Dim    lipgloss.Style
	Accent lipgloss.Style
	Error  lipgloss.Style
}

var safeFallbackRef = SafeFallbackStyles{}

func ApplySafeFallback(theme *config.Theme) {
	if theme == nil {
		return
	}
	safeFallbackRef = SafeFallbackStyles{
		Normal: lipgloss.NewStyle().Foreground(lipgloss.Color(theme.ResolveColor(theme.SafeFallback.Normal))),
		Bold:   lipgloss.NewStyle().Foreground(lipgloss.Color(theme.ResolveColor(theme.SafeFallback.Bold))).Bold(true),
		Dim:    lipgloss.NewStyle().Foreground(lipgloss.Color(theme.ResolveColor(theme.SafeFallback.Dim))).Faint(true),
		Accent: lipgloss.NewStyle().Foreground(lipgloss.Color(theme.ResolveColor(theme.SafeFallback.Accent))),
		Error:  lipgloss.NewStyle().Foreground(lipgloss.Color(theme.ResolveColor(theme.SafeFallback.Error))),
	}
}

// RawStylesForTest exposes the package-level style cache to tests in other packages.
func RawStylesForTest() globalStyleRefs { return rawStyles }

// SetRawStylesForTest restores the package-level style cache from tests.
func SetRawStylesForTest(refs globalStyleRefs) { rawStyles = refs }

// SafeFallbackForTest exposes the package-level safe fallback cache to tests.
func SafeFallbackForTest() SafeFallbackStyles { return safeFallbackRef }

// SetSafeFallbackForTest restores the package-level safe fallback cache from tests.
func SetSafeFallbackForTest(refs SafeFallbackStyles) { safeFallbackRef = refs }

func RenderBackgroundLayer(content string, background color.Color) string {
	if content == "" || background == nil {
		return content
	}
	r, g, b, _ := background.RGBA()
	backgroundANSI := fmt.Sprintf("\033[48;2;%d;%d;%dm", r/257, g/257, b/257)
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		line = strings.ReplaceAll(line, "\033[0m", "\033[0m"+backgroundANSI)
		lines[i] = backgroundANSI + line + "\033[0m"
	}
	return strings.Join(lines, "\n")
}

func GetStyle(name string) lipgloss.Style {
	if ref, ok := rawStyles.lookup(name); ok {
		return buildStyle(ref)
	}
	if c := style.Color(name); c != nil {
		return lipgloss.NewStyle().Foreground(c)
	}
	return safeFallbackRef.Normal
}

func (s globalStyleRefs) lookup(name string) (styleRef, bool) {
	switch name {
	case "searchBar":
		return s.SearchBar, true
	case "searchCursor":
		return s.SearchCursor, true
	case "searchHint":
		return s.SearchHint, true
	case "commandPrefix":
		return s.CommandPrefix, true
	case "breadcrumb":
		return s.Breadcrumb, true
	case "breadcrumbActive":
		return s.BreadcrumbActive, true
	case "toastSuccess":
		return s.ToastSuccess, true
	case "toastError":
		return s.ToastError, true
	case "toastInfo":
		return s.ToastInfo, true
	case "toastWarning":
		return s.ToastWarning, true
	case "hintKey":
		return s.HintKey, true
	case "hintDesc":
		return s.HintDescription, true
	case "hintSep":
		return s.HintSeparator, true
	case "dialogTitle":
		return s.DialogTitle, true
	case "dialogBody":
		return s.DialogBody, true
	case "dialogBodyBackground":
		return s.DialogBodyBackground, true
	case "dialogOption":
		return s.DialogOption, true
	case "dialogOptionDisabled":
		return s.DialogOptionDisabled, true
	case "panel":
		return s.Panel, true
	case "panelTitle":
		return s.PanelTitle, true
	case "dim":
		return s.Dim, true
	case "helpKey":
		return s.HelpKey, true
	case "helpDesc":
		return s.HelpDescription, true
	case "detailSection":
		return s.DetailSection, true
	case "detailLabel":
		return s.DetailLabel, true
	case "detailValue":
		return s.DetailValue, true
	case "detailDim":
		return s.DetailDim, true
	case "detailSelection":
		return s.DetailSelection, true
	case "logTimestamp":
		return s.LogTimestamp, true
	case "logText":
		return s.LogText, true
	case "logStderr":
		return s.LogStderr, true
	case "logHighlightBg":
		return s.LogHighlight, true
	case "header":
		return s.Header, true
	case "headerBar":
		return s.HeaderBar, true
	case "headerLabel":
		return s.HeaderLabel, true
	case "keyBadge":
		return s.KeyBadge, true
	case "keyLast":
		return s.KeyLast, true
	case "footer":
		return s.Footer, true
	case "shortcutBar":
		return s.ShortcutBar, true
	case "actionBar":
		return s.ActionBar, true
	case "messageRail":
		return s.MessageRail, true
	case "queryBar":
		return s.QueryBar, true
	case "selectedRow":
		return s.SelectedRow, true
	case "dialogConfirm":
		return s.DialogConfirm, true
	case "dialogError":
		return s.DialogError, true
	case "dialogWarning":
		return s.DialogWarning, true
	default:
		return styleRef{}, false
	}
}

// buildStyle 将 styleRef 转换为 lipgloss.Style，通过当前调色板解析颜色名。
// 由 getStyleChain 在第 2 层和第 1 层调用。
func buildStyle(ref styleRef) lipgloss.Style {
	s := lipgloss.NewStyle()
	if ref.Color != "" {
		if c, ok := utils.ParseColor(ref.Color); ok {
			s = s.Foreground(c)
		}
	}
	if ref.Background != "" {
		if c, ok := utils.ParseColor(ref.Background); ok {
			s = s.Background(c)
		}
	}
	if ref.Bold {
		s = s.Bold(true)
	}
	if ref.Faint {
		s = s.Faint(true)
	}
	return s
}
