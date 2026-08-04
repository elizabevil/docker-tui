package component

import (
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
	SelectedRow          styleRef `json:"selectedRow"`
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
	rawStyles.SearchBar.Color = resolve(theme.Text.HelpDescription)
	rawStyles.SearchCursor.Color = resolve(theme.Text.HelpDescription)
	rawStyles.SearchHint.Color = resolve(theme.Text.Dim)
	rawStyles.CommandPrefix.Color = resolve(theme.Text.Success)
	rawStyles.Breadcrumb.Color = resolve(theme.Text.Dim)
	rawStyles.BreadcrumbActive.Color = resolve(theme.Text.Info)
	rawStyles.ToastSuccess = styleRef{Color: resolve(theme.Toast.Success), Background: resolve(theme.Toast.Background)}
	rawStyles.ToastError = styleRef{Color: resolve(theme.Toast.Error), Background: resolve(theme.Toast.Background)}
	rawStyles.ToastInfo = styleRef{Color: resolve(theme.Text.Info), Background: resolve(theme.Toast.Background)}
	rawStyles.ToastWarning = styleRef{Color: resolve(theme.Text.Warning), Background: resolve(theme.Toast.Background)}
	rawStyles.HintKey.Color = resolve(theme.Footer.Key)
	rawStyles.HintDescription.Color = resolve(theme.Footer.Description)
	rawStyles.HintSeparator.Color = resolve(theme.Footer.Separator)
	rawStyles.DialogTitle.Color = resolve(theme.Dialog.Title)
	rawStyles.DialogBody.Color = resolve(theme.Dialog.Body)
	rawStyles.DialogOption.Color = resolve(theme.Dialog.OptionActive)
	rawStyles.DialogOptionDisabled.Color = resolve(theme.Dialog.OptionInactive)
	rawStyles.PanelTitle.Color = resolve(theme.Main.Title)
	rawStyles.Dim.Color = resolve(theme.Text.Dim)
	rawStyles.HelpKey.Color = resolve(theme.Text.HelpKey)
	rawStyles.HelpDescription.Color = resolve(theme.Text.HelpDescription)
	rawStyles.DetailLabel.Color = resolve(theme.Text.Info)
	rawStyles.DetailSection.Color = resolve(theme.Text.HelpDescription)
	rawStyles.DetailValue.Color = resolve(theme.Text.HelpDescription)
	rawStyles.DetailDim.Color = resolve(theme.Text.Dim)
	rawStyles.DetailSelection = styleRef{Color: resolve(theme.Main.RowText), Background: resolve(theme.Main.RowSelected)}
	rawStyles.LogTimestamp.Color = resolve(theme.Text.Info)
	rawStyles.LogText.Color = resolve(theme.Text.HelpDescription)
	rawStyles.LogStderr.Color = resolve(theme.Text.Error)
	rawStyles.LogHighlight = styleRef{Color: resolve(theme.Header.Background), Background: resolve(theme.Text.Warning)}
	rawStyles.Header = styleRef{Color: resolve(theme.Main.TableHeader), Bold: true}
	rawStyles.HeaderBar = styleRef{Color: resolve(theme.Header.Value), Background: resolve(theme.Header.Background)}
	rawStyles.HeaderLabel = styleRef{Color: resolve(theme.Header.Label)}
	rawStyles.KeyBadge = styleRef{Color: resolve(theme.Header.Value), Bold: true}
	rawStyles.KeyLast = styleRef{Color: resolve(theme.Text.Dim)}
	rawStyles.Footer = styleRef{Color: resolve(theme.Main.Footer)}
	rawStyles.SelectedRow = styleRef{Color: resolve(theme.Main.RowText), Background: resolve(theme.Main.RowSelected), Bold: true}

	tableCfg.RowStyles.Selected.Color = resolve(theme.Main.RowText)
	tableCfg.RowStyles.Selected.Background = resolve(theme.Main.RowSelected)
	tableCfg.RowStyles.Selected.Bold = true

	ApplyTableTheme(theme)
	ApplySafeFallback(theme)
}

func ApplyTableTheme(theme *config.Theme) {
	if theme == nil {
		return
	}
	resolve := theme.ResolveColor
	tableCfg.RowStyles.Marked.Background = resolve(theme.Table.MarkedBackground)
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
	case "dialogOption":
		return s.DialogOption, true
	case "dialogOptionDisabled":
		return s.DialogOptionDisabled, true
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
	case "selectedRow":
		return s.SelectedRow, true
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
