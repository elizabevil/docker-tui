package component

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/elizabevil/docker-tui/internal/data/config"
	"github.com/elizabevil/docker-tui/internal/utils"
)

// StyleName 标识已注册的组件样式,是 GetStyle 的入参类型。
// 所有取值由本组常量定义,调用方只能传入编译期已知的名称,
// 拼写错误无法通过编译;运行时未注册值一律回退到 safe fallback。
type StyleName string

// 已注册的组件样式名,与 globalStyleRefs 的 JSON 键一一对应。
const (
	StyleSearchBar            StyleName = "searchBar"
	StyleSearchCursor         StyleName = "searchCursor"
	StyleSearchHint           StyleName = "searchHint"
	StyleCommandPrefix        StyleName = "commandPrefix"
	StyleBreadcrumb           StyleName = "breadcrumb"
	StyleBreadcrumbActive     StyleName = "breadcrumbActive"
	StyleToastSuccess         StyleName = "toastSuccess"
	StyleToastError           StyleName = "toastError"
	StyleToastInfo            StyleName = "toastInfo"
	StyleToastWarning         StyleName = "toastWarning"
	StyleHintKey              StyleName = "hintKey"
	StyleHintDescription      StyleName = "hintDesc"
	StyleHintSeparator        StyleName = "hintSep"
	StyleDialogTitle          StyleName = "dialogTitle"
	StyleDialogBody           StyleName = "dialogBody"
	StyleDialogOption         StyleName = "dialogOption"
	StyleDialogOptionDisabled StyleName = "dialogOptionDisabled"
	StyleDialogBodyBackground StyleName = "dialogBodyBackground"
	StylePanel                StyleName = "panel"
	StylePanelTitle           StyleName = "panelTitle"
	StyleDim                  StyleName = "dim"
	StyleHelpKey              StyleName = "helpKey"
	StyleHelpDescription      StyleName = "helpDesc"
	StyleDetailSection        StyleName = "detailSection"
	StyleDetailLabel          StyleName = "detailLabel"
	StyleDetailValue          StyleName = "detailValue"
	StyleDetailDim            StyleName = "detailDim"
	StyleDetailSelection      StyleName = "detailSelection"
	StyleLogTimestamp         StyleName = "logTimestamp"
	StyleLogText              StyleName = "logText"
	StyleLogStderr            StyleName = "logStderr"
	StyleLogHighlight         StyleName = "logHighlightBg"
	StyleHeader               StyleName = "header"
	StyleHeaderBar            StyleName = "headerBar"
	StyleHeaderLabel          StyleName = "headerLabel"
	StyleKeyBadge             StyleName = "keyBadge"
	StyleKeyLast              StyleName = "keyLast"
	StyleFooter               StyleName = "footer"
	StyleShortcutBar          StyleName = "shortcutBar"
	StyleActionBar            StyleName = "actionBar"
	StyleActionBarBorder      StyleName = "actionBarBorder"
	StyleFormInput            StyleName = "formInput"
	StyleMessageRail          StyleName = "messageRail"
	StyleQueryBar             StyleName = "queryBar"
	StyleSelectedRow          StyleName = "selectedRow"
	StyleDialogConfirm        StyleName = "dialogConfirm"
	StyleDialogError          StyleName = "dialogError"
	StyleDialogWarning        StyleName = "dialogWarning"
	StyleActionContainerWindow       StyleName = "actionContainerWindow"
	StyleActionContainerWindowBorder StyleName = "actionContainerWindowBorder"
	StyleActionContainerFormInput    StyleName = "actionContainerFormInput"
	StyleActionContainerConfirm      StyleName = "actionContainerConfirm"
	StyleActionContainerCancel       StyleName = "actionContainerCancel"
	StyleActionImageWindow           StyleName = "actionImageWindow"
	StyleActionImageWindowBorder     StyleName = "actionImageWindowBorder"
	StyleActionImageFormInput        StyleName = "actionImageFormInput"
	StyleActionImageConfirm          StyleName = "actionImageConfirm"
	StyleActionImageCancel           StyleName = "actionImageCancel"
)

// allStyleNames 列出全部已注册样式名,供测试断言 lookup 无遗漏。
var allStyleNames = []StyleName{
	StyleSearchBar, StyleSearchCursor, StyleSearchHint, StyleCommandPrefix,
	StyleBreadcrumb, StyleBreadcrumbActive, StyleToastSuccess, StyleToastError,
	StyleToastInfo, StyleToastWarning, StyleHintKey, StyleHintDescription,
	StyleHintSeparator, StyleDialogTitle, StyleDialogBody, StyleDialogOption,
	StyleDialogOptionDisabled, StyleDialogBodyBackground, StylePanel, StylePanelTitle,
	StyleDim, StyleHelpKey, StyleHelpDescription, StyleDetailSection, StyleDetailLabel,
	StyleDetailValue, StyleDetailDim, StyleDetailSelection, StyleLogTimestamp,
	StyleLogText, StyleLogStderr, StyleLogHighlight, StyleHeader, StyleHeaderBar,
	StyleHeaderLabel, StyleKeyBadge, StyleKeyLast, StyleFooter, StyleShortcutBar,
	StyleActionBar, StyleActionBarBorder, StyleFormInput, StyleMessageRail, StyleQueryBar, StyleSelectedRow,
	StyleDialogConfirm, StyleDialogError, StyleDialogWarning,
	StyleActionContainerWindow, StyleActionContainerWindowBorder, StyleActionContainerFormInput, StyleActionContainerConfirm, StyleActionContainerCancel,
	StyleActionImageWindow, StyleActionImageWindowBorder, StyleActionImageFormInput, StyleActionImageConfirm, StyleActionImageCancel,
}

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
	ActionBarBorder      styleRef `json:"actionBarBorder"`
	FormInput            styleRef `json:"formInput"`
	MessageRail          styleRef `json:"messageRail"`
	QueryBar             styleRef `json:"queryBar"`
	SelectedRow          styleRef `json:"selectedRow"`
	DialogConfirm        styleRef `json:"dialogConfirm"`
	DialogError          styleRef `json:"dialogError"`
	DialogWarning        styleRef `json:"dialogWarning"`
	ActionContainerWindow       styleRef `json:"actionContainerWindow"`
	ActionContainerWindowBorder styleRef `json:"actionContainerWindowBorder"`
	ActionContainerFormInput    styleRef `json:"actionContainerFormInput"`
	ActionContainerConfirm      styleRef `json:"actionContainerConfirm"`
	ActionContainerCancel       styleRef `json:"actionContainerCancel"`
	ActionImageWindow           styleRef `json:"actionImageWindow"`
	ActionImageWindowBorder     styleRef `json:"actionImageWindowBorder"`
	ActionImageFormInput        styleRef `json:"actionImageFormInput"`
	ActionImageConfirm          styleRef `json:"actionImageConfirm"`
	ActionImageCancel           styleRef `json:"actionImageCancel"`
}

// styleRef 是 utils.StyleRef 的本地别名。组件层沿用短名以保留
// 字段标注的紧凑布局；类型与 BuildStyle 方法来自 utils，以便其它
// 包跨包复用相同的颜色解析和 α=0 守卫。
type styleRef = utils.StyleRef

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
	resolveForeground := func(ref config.ColorRef) string {
		return flattenThemeForeground(theme, ref)
	}
	resolveBackground := func(ref config.ColorRef) string {
		return flattenThemeBackground(theme, ref)
	}
	rawStyles.SearchBar.Color = resolveForeground(theme.Text.HelpDescription)
	rawStyles.SearchCursor.Color = resolveForeground(theme.Text.HelpDescription)
	rawStyles.SearchHint.Color = resolveForeground(theme.Text.Dim)
	rawStyles.CommandPrefix.Color = resolveForeground(theme.Text.Success)
	rawStyles.Breadcrumb.Color = resolveForeground(theme.Text.Dim)
	rawStyles.BreadcrumbActive.Color = resolveForeground(theme.Text.Info)
	rawStyles.ToastSuccess = styleRef{Color: resolveForeground(theme.Toast.Success), Background: resolveBackground(theme.Toast.Background)}
	rawStyles.ToastError = styleRef{Color: resolveForeground(theme.Toast.Error), Background: resolveBackground(theme.Toast.Background)}
	rawStyles.ToastInfo = styleRef{Color: resolveForeground(theme.Text.Info), Background: resolveBackground(theme.Toast.Background)}
	rawStyles.ToastWarning = styleRef{Color: resolveForeground(theme.Text.Warning), Background: resolveBackground(theme.Toast.Background)}
	rawStyles.HintKey.Color = resolveForeground(theme.Footer.Key)
	rawStyles.HintDescription.Color = resolveForeground(theme.Footer.Description)
	rawStyles.HintSeparator.Color = resolveForeground(theme.Footer.Separator)
	rawStyles.DialogTitle.Color = resolveForeground(theme.Dialog.Title)
	rawStyles.DialogBody.Color = resolveForeground(theme.Dialog.Body)
	rawStyles.DialogBodyBackground = styleRef{Background: resolveBackground(theme.Dialog.BodyBackground)}
	rawStyles.DialogOption.Color = resolveForeground(theme.Dialog.OptionActive)
	rawStyles.DialogOptionDisabled.Color = resolveForeground(theme.Dialog.OptionInactive)
	rawStyles.Panel = styleRef{Background: resolveBackground(theme.Main.PanelBackground)}
	rawStyles.PanelTitle.Color = resolveForeground(theme.Main.Title)
	rawStyles.Dim.Color = resolveForeground(theme.Text.Dim)
	rawStyles.HelpKey.Color = resolveForeground(theme.Text.HelpKey)
	rawStyles.HelpDescription.Color = resolveForeground(theme.Text.HelpDescription)
	rawStyles.DetailLabel.Color = resolveForeground(theme.Text.Info)
	rawStyles.DetailSection.Color = resolveForeground(theme.Text.HelpDescription)
	rawStyles.DetailValue.Color = resolveForeground(theme.Text.HelpDescription)
	rawStyles.DetailDim.Color = resolveForeground(theme.Text.Dim)
	rawStyles.DetailSelection = styleRef{Color: resolveForeground(theme.Main.RowText), Background: resolveBackground(theme.Main.RowSelected), Bold: true}
	rawStyles.LogTimestamp.Color = resolveForeground(theme.Text.Info)
	rawStyles.LogText.Color = resolveForeground(theme.Text.HelpDescription)
	rawStyles.LogStderr.Color = resolveForeground(theme.Text.Error)
	rawStyles.LogHighlight = styleRef{Color: resolveBackground(theme.Header.Background), Background: resolveBackground(theme.Text.Warning)}
	rawStyles.Header = styleRef{Color: resolveForeground(theme.Main.TableHeader), Bold: true}
	rawStyles.HeaderBar = styleRef{Color: resolveForeground(theme.Header.Value), Background: resolveBackground(theme.Header.Background)}
	rawStyles.HeaderLabel = styleRef{Color: resolveForeground(theme.Header.Label)}
	rawStyles.KeyBadge = styleRef{Color: resolveForeground(theme.Header.Value), Background: resolveBackground(theme.Header.Background), Bold: true}
	rawStyles.Footer = styleRef{Color: resolveForeground(theme.Main.Footer)}
	rawStyles.ShortcutBar = styleRef{Background: resolveBackground(theme.Footer.ShortcutBackground)}
	rawStyles.ActionBar = styleRef{Color: resolveForeground(theme.Main.RowText), Background: resolveBackground(theme.Main.ActionBarBackground)}
	rawStyles.ActionBarBorder = styleRef{Color: resolveForeground(theme.Main.BorderActive)}
	rawStyles.FormInput = styleRef{Background: string(theme.Palette.Background)}
	rawStyles.MessageRail = styleRef{Background: resolveBackground(theme.Main.MessageRailBackground)}
	rawStyles.QueryBar = styleRef{Background: resolveBackground(theme.Main.QueryBarBackground)}
	rawStyles.SelectedRow = styleRef{Color: resolveForeground(theme.Main.RowText), Background: resolveBackground(theme.Main.RowSelected)}
	rawStyles.DialogConfirm = styleRef{Color: resolveForeground(theme.Text.Success), Background: resolveBackground(theme.Toast.Background), Bold: true}
	rawStyles.DialogError = styleRef{Color: resolveForeground(theme.Text.Error), Background: resolveBackground(theme.Toast.Background)}
	rawStyles.DialogWarning = styleRef{Color: resolveForeground(theme.Text.Warning), Background: resolveBackground(theme.Toast.Background)}
	rawStyles.ActionContainerWindow = styleRef{Background: resolveBackground(theme.Action.Container.Window.Background)}
	rawStyles.ActionContainerWindowBorder = styleRef{Color: resolveForeground(theme.Action.Container.Window.Border)}
	rawStyles.ActionContainerFormInput = styleRef{Color: resolveForeground(theme.Action.Container.FormInput.Foreground), Background: resolveBackground(theme.Action.Container.FormInput.Background)}
	rawStyles.ActionContainerConfirm = styleRef{Color: resolveForeground(theme.Action.Container.Confirm.Foreground), Background: resolveBackground(theme.Action.Container.Confirm.Background), Bold: true}
	rawStyles.ActionContainerCancel = styleRef{Color: resolveForeground(theme.Action.Container.Cancel.Foreground), Background: resolveBackground(theme.Action.Container.Cancel.Background)}
	rawStyles.ActionImageWindow = styleRef{Background: resolveBackground(theme.Action.Image.Window.Background)}
	rawStyles.ActionImageWindowBorder = styleRef{Color: resolveForeground(theme.Action.Image.Window.Border)}
	rawStyles.ActionImageFormInput = styleRef{Color: resolveForeground(theme.Action.Image.FormInput.Foreground), Background: resolveBackground(theme.Action.Image.FormInput.Background)}
	rawStyles.ActionImageConfirm = styleRef{Color: resolveForeground(theme.Action.Image.Confirm.Foreground), Background: resolveBackground(theme.Action.Image.Confirm.Background), Bold: true}
	rawStyles.ActionImageCancel = styleRef{Color: resolveForeground(theme.Action.Image.Cancel.Foreground), Background: resolveBackground(theme.Action.Image.Cancel.Background)}

	tableCfg.RowStyles.Selected.Color = resolveForeground(theme.Main.RowText)
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

func flattenThemeForeground(theme *config.Theme, ref config.ColorRef) string {
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
	baseParsed, ok := utils.ParseColor(string(theme.Palette.Foreground))
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
	tableCfg.RowStyles.Marked.Background = flattenThemeBackground(theme, theme.Table.MarkedBackground)
	tableCfg.RowStyles.Marked.Color = flattenThemeForeground(theme, theme.Table.NameForeground)
	tableCfg.ColumnStyles = defaultColumnStyles(flattenThemeForeground(theme, theme.Table.ColumnForeground), flattenThemeForeground(theme, theme.Table.NameForeground))
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
	// lipgloss.Style.GetBackground() returns lipgloss.NoColor{} — a value
	// type, not nil — when the background slot is unset. NoColor.RGBA()
	// returns (0, 0, 0, 0xFFFF), so without this guard the function below
	// would emit an opaque-black SGR over content the caller wanted to
	// leave transparent. The type assertion is the only reliable signal.
	if _, isNoColor := background.(lipgloss.NoColor); isNoColor {
		return content
	}
	r, g, b, a := background.RGBA()
	if a == 0 {
		return content
	}
	backgroundANSI := fmt.Sprintf("\033[48;2;%d;%d;%dm", r/257, g/257, b/257)
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		line = strings.ReplaceAll(line, "\033[0m", "\033[0m"+backgroundANSI)
		lines[i] = backgroundANSI + line + "\033[0m"
	}
	return strings.Join(lines, "\n")
}

func GetStyle(name StyleName) lipgloss.Style {
	if ref, ok := rawStyles.lookup(name); ok {
		return ref.BuildStyle()
	}
	return safeFallbackRef.Normal
}

func (s globalStyleRefs) lookup(name StyleName) (styleRef, bool) {
	switch name {
	case StyleSearchBar:
		return s.SearchBar, true
	case StyleSearchCursor:
		return s.SearchCursor, true
	case StyleSearchHint:
		return s.SearchHint, true
	case StyleCommandPrefix:
		return s.CommandPrefix, true
	case StyleBreadcrumb:
		return s.Breadcrumb, true
	case StyleBreadcrumbActive:
		return s.BreadcrumbActive, true
	case StyleToastSuccess:
		return s.ToastSuccess, true
	case StyleToastError:
		return s.ToastError, true
	case StyleToastInfo:
		return s.ToastInfo, true
	case StyleToastWarning:
		return s.ToastWarning, true
	case StyleHintKey:
		return s.HintKey, true
	case StyleHintDescription:
		return s.HintDescription, true
	case StyleHintSeparator:
		return s.HintSeparator, true
	case StyleDialogTitle:
		return s.DialogTitle, true
	case StyleDialogBody:
		return s.DialogBody, true
	case StyleDialogBodyBackground:
		return s.DialogBodyBackground, true
	case StyleDialogOption:
		return s.DialogOption, true
	case StyleDialogOptionDisabled:
		return s.DialogOptionDisabled, true
	case StylePanel:
		return s.Panel, true
	case StylePanelTitle:
		return s.PanelTitle, true
	case StyleDim:
		return s.Dim, true
	case StyleHelpKey:
		return s.HelpKey, true
	case StyleHelpDescription:
		return s.HelpDescription, true
	case StyleDetailSection:
		return s.DetailSection, true
	case StyleDetailLabel:
		return s.DetailLabel, true
	case StyleDetailValue:
		return s.DetailValue, true
	case StyleDetailDim:
		return s.DetailDim, true
	case StyleDetailSelection:
		return s.DetailSelection, true
	case StyleLogTimestamp:
		return s.LogTimestamp, true
	case StyleLogText:
		return s.LogText, true
	case StyleLogStderr:
		return s.LogStderr, true
	case StyleLogHighlight:
		return s.LogHighlight, true
	case StyleHeader:
		return s.Header, true
	case StyleHeaderBar:
		return s.HeaderBar, true
	case StyleHeaderLabel:
		return s.HeaderLabel, true
	case StyleKeyBadge:
		return s.KeyBadge, true
	case StyleKeyLast:
		return s.KeyLast, true
	case StyleFooter:
		return s.Footer, true
	case StyleShortcutBar:
		return s.ShortcutBar, true
	case StyleActionBar:
		return s.ActionBar, true
	case StyleActionBarBorder:
		return s.ActionBarBorder, true
	case StyleFormInput:
		return s.FormInput, true
	case StyleMessageRail:
		return s.MessageRail, true
	case StyleQueryBar:
		return s.QueryBar, true
	case StyleSelectedRow:
		return s.SelectedRow, true
	case StyleDialogConfirm:
		return s.DialogConfirm, true
	case StyleDialogError:
		return s.DialogError, true
	case StyleDialogWarning:
		return s.DialogWarning, true
	case StyleActionContainerWindow:
		return s.ActionContainerWindow, true
	case StyleActionContainerWindowBorder:
		return s.ActionContainerWindowBorder, true
	case StyleActionContainerFormInput:
		return s.ActionContainerFormInput, true
	case StyleActionContainerConfirm:
		return s.ActionContainerConfirm, true
	case StyleActionContainerCancel:
		return s.ActionContainerCancel, true
	case StyleActionImageWindow:
		return s.ActionImageWindow, true
	case StyleActionImageWindowBorder:
		return s.ActionImageWindowBorder, true
	case StyleActionImageFormInput:
		return s.ActionImageFormInput, true
	case StyleActionImageConfirm:
		return s.ActionImageConfirm, true
	case StyleActionImageCancel:
		return s.ActionImageCancel, true
	default:
		return styleRef{}, false
	}
}
