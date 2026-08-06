package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

type Color string
type BorderKind string
type ThemeName string

type ColorRef struct {
	Token ColorToken `json:"token,omitempty" yaml:"token,omitempty"`
	Value Color      `json:"value,omitempty" yaml:"value,omitempty"`
}

func TokenRef(token ColorToken) ColorRef {
	return ColorRef{Token: token}
}

func ValueRef(value Color) ColorRef {
	return ColorRef{Value: value}
}

const (
	BorderRounded BorderKind = "rounded"
	BorderSingle  BorderKind = "single"
	BorderDouble  BorderKind = "double"
	BorderThick   BorderKind = "thick"
	BorderHidden  BorderKind = "hidden"
)

type Palette struct {
	Primary          Color `json:"primary" yaml:"primary"`
	Success          Color `json:"success" yaml:"success"`
	Warning          Color `json:"warning" yaml:"warning"`
	Danger           Color `json:"danger" yaml:"danger"`
	Info             Color `json:"info" yaml:"info"`
	Accent           Color `json:"accent" yaml:"accent"`
	AccentSecondary  Color `json:"accentSecondary" yaml:"accentSecondary"`
	Foreground       Color `json:"foreground" yaml:"foreground"`
	ForegroundMuted  Color `json:"foregroundMuted" yaml:"foregroundMuted"`
	Background       Color `json:"background" yaml:"background"`
	BackgroundSubtle Color `json:"backgroundSubtle" yaml:"backgroundSubtle"`
	BackgroundDeep   Color `json:"backgroundDeep" yaml:"backgroundDeep"`
}

type BorderStyles struct {
	Active   ColorRef   `json:"active" yaml:"active"`
	Inactive ColorRef   `json:"inactive" yaml:"inactive"`
	Focused  ColorRef   `json:"focused" yaml:"focused"`
	Kind     BorderKind `json:"kind" yaml:"kind"`
}

type HeaderStyles struct {
	Background ColorRef `json:"background" yaml:"background"`
	Label      ColorRef `json:"label" yaml:"label"`
	Value      ColorRef `json:"value" yaml:"value"`
	Logo       ColorRef `json:"logo" yaml:"logo"`
}

type MainStyles struct {
	BorderActive          ColorRef `json:"borderActive" yaml:"borderActive"`
	BorderInactive        ColorRef `json:"borderInactive" yaml:"borderInactive"`
	Title                 ColorRef `json:"title" yaml:"title"`
	TableHeader           ColorRef `json:"tableHeader" yaml:"tableHeader"`
	RowSelected           ColorRef `json:"rowSelected" yaml:"rowSelected"`
	RowText               ColorRef `json:"rowText" yaml:"rowText"`
	Footer                ColorRef `json:"footer" yaml:"footer"`
	PanelBackground       ColorRef `json:"panelBackground" yaml:"panelBackground"`
	ActionBarBackground   ColorRef `json:"actionBarBackground" yaml:"actionBarBackground"`
	MessageRailBackground ColorRef `json:"messageRailBackground" yaml:"messageRailBackground"`
	QueryBarBackground    ColorRef `json:"queryBarBackground" yaml:"queryBarBackground"`
}

type FooterStyles struct {
	StatusBackground   ColorRef `json:"statusBackground" yaml:"statusBackground"`
	ShortcutBackground ColorRef `json:"shortcutBackground" yaml:"shortcutBackground"`
	Key                ColorRef `json:"key" yaml:"key"`
	Description        ColorRef `json:"description" yaml:"description"`
	Separator          ColorRef `json:"separator" yaml:"separator"`
}

type DialogStyles struct {
	Border         ColorRef `json:"border" yaml:"border"`
	Title          ColorRef `json:"title" yaml:"title"`
	Body           ColorRef `json:"body" yaml:"body"`
	BodyBackground ColorRef `json:"bodyBackground" yaml:"bodyBackground"`
	OptionActive   ColorRef `json:"optionActive" yaml:"optionActive"`
	OptionInactive ColorRef `json:"optionInactive" yaml:"optionInactive"`
	Overlay        ColorRef `json:"overlay" yaml:"overlay"`
	OverlayOpacity uint8    `json:"overlayOpacity" yaml:"overlayOpacity"`
}

type ToastStyles struct {
	Success    ColorRef `json:"success" yaml:"success"`
	Error      ColorRef `json:"error" yaml:"error"`
	Background ColorRef `json:"background" yaml:"background"`
}

type TextStyles struct {
	Info            ColorRef `json:"info" yaml:"info"`
	Error           ColorRef `json:"error" yaml:"error"`
	Success         ColorRef `json:"success" yaml:"success"`
	Warning         ColorRef `json:"warning" yaml:"warning"`
	Dim             ColorRef `json:"dim" yaml:"dim"`
	HelpKey         ColorRef `json:"helpKey" yaml:"helpKey"`
	HelpDescription ColorRef `json:"helpDescription" yaml:"helpDescription"`
}

type TableStyles struct {
	MarkedBackground ColorRef `json:"markedBackground" yaml:"markedBackground"`
	ColumnForeground ColorRef `json:"columnForeground" yaml:"columnForeground"`
	NameForeground   ColorRef `json:"nameForeground" yaml:"nameForeground"`
}

type SafeFallbackStyles struct {
	Normal ColorRef `json:"normal" yaml:"normal"`
	Bold   ColorRef `json:"bold" yaml:"bold"`
	Dim    ColorRef `json:"dim" yaml:"dim"`
	Accent ColorRef `json:"accent" yaml:"accent"`
	Error  ColorRef `json:"error" yaml:"error"`
}

// ActionWindow groups the background + border colours of the window chrome
// used by any Operation whose body fills a window (Form / Page / Async).
type ActionWindow struct {
	Background ColorRef `json:"background" yaml:"background"`
	Border     ColorRef `json:"border" yaml:"border"`
}

// ActionFormInput groups the foreground + background colours of the
// editable form input fields inside the window. Only consumed when the
// Operation Mode is Form.
type ActionFormInput struct {
	Foreground ColorRef `json:"foreground" yaml:"foreground"`
	Background ColorRef `json:"background" yaml:"background"`
}

// ActionButton groups the foreground + background colours of an action
// button (Confirm or Cancel) inside the window. Only consumed when the
// Operation Mode is Form or Async.
type ActionButton struct {
	Foreground ColorRef `json:"foreground" yaml:"foreground"`
	Background ColorRef `json:"background" yaml:"background"`
}

// ActionScopeStyles is the per-scope visual contract: one entry per
// resource scope (Container / Image / ...). The four slots are reused
// across all scopes so each Operation Mode (Form / Page / Async) can
// pull from the same shape regardless of scope.
type ActionScopeStyles struct {
	Window    ActionWindow    `json:"window" yaml:"window"`
	FormInput ActionFormInput `json:"formInput" yaml:"formInput"`
	Confirm   ActionButton    `json:"confirm" yaml:"confirm"`
	Cancel    ActionButton    `json:"cancel" yaml:"cancel"`
}

// ActionStyles is the visual contract for any Operation, grouped by
// resource Scope. It supersedes the earlier single-scope
// ContainerOperationStyles; the Skin-vs-Skeleton separation is enforced
// by JSON schema (themes only define colours, operations/*.jsonc — once
// introduced — will only define behaviour).
type ActionStyles struct {
	Container ActionScopeStyles `json:"container" yaml:"container"`
	Image     ActionScopeStyles `json:"image" yaml:"image"`
}

// Theme is a complete visual configuration with fixed component scopes.
type Theme struct {
	Palette      Palette            `json:"palette" yaml:"palette"`
	Border       BorderStyles       `json:"border" yaml:"border"`
	Header       HeaderStyles       `json:"header" yaml:"header"`
	Main         MainStyles         `json:"main" yaml:"main"`
	Footer       FooterStyles       `json:"footer" yaml:"footer"`
	Dialog       DialogStyles       `json:"dialog" yaml:"dialog"`
	Toast        ToastStyles        `json:"toast" yaml:"toast"`
	Text         TextStyles         `json:"text" yaml:"text"`
	Table        TableStyles        `json:"table" yaml:"table"`
	SafeFallback SafeFallbackStyles `json:"safeFallback" yaml:"safeFallback"`
	Action       ActionStyles       `json:"action" yaml:"action"`
}

type ThemeMetadata struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"description"`
}

type ThemeDocument struct {
	Meta  ThemeMetadata `json:"meta" yaml:"meta"`
	Theme ThemePatch    `json:"theme" yaml:"theme"`
}

type PalettePatch struct {
	Primary          *Color `json:"primary,omitempty" yaml:"primary,omitempty"`
	Success          *Color `json:"success,omitempty" yaml:"success,omitempty"`
	Warning          *Color `json:"warning,omitempty" yaml:"warning,omitempty"`
	Danger           *Color `json:"danger,omitempty" yaml:"danger,omitempty"`
	Info             *Color `json:"info,omitempty" yaml:"info,omitempty"`
	Accent           *Color `json:"accent,omitempty" yaml:"accent,omitempty"`
	AccentSecondary  *Color `json:"accentSecondary,omitempty" yaml:"accentSecondary,omitempty"`
	Foreground       *Color `json:"foreground,omitempty" yaml:"foreground,omitempty"`
	ForegroundMuted  *Color `json:"foregroundMuted,omitempty" yaml:"foregroundMuted,omitempty"`
	Background       *Color `json:"background,omitempty" yaml:"background,omitempty"`
	BackgroundSubtle *Color `json:"backgroundSubtle,omitempty" yaml:"backgroundSubtle,omitempty"`
	BackgroundDeep   *Color `json:"backgroundDeep,omitempty" yaml:"backgroundDeep,omitempty"`
}

type BorderStylesPatch struct {
	Active   *ColorRef   `json:"active,omitempty" yaml:"active,omitempty"`
	Inactive *ColorRef   `json:"inactive,omitempty" yaml:"inactive,omitempty"`
	Focused  *ColorRef   `json:"focused,omitempty" yaml:"focused,omitempty"`
	Kind     *BorderKind `json:"kind,omitempty" yaml:"kind,omitempty"`
}
type HeaderStylesPatch struct {
	Background *ColorRef `json:"background,omitempty" yaml:"background,omitempty"`
	Label      *ColorRef `json:"label,omitempty" yaml:"label,omitempty"`
	Value      *ColorRef `json:"value,omitempty" yaml:"value,omitempty"`
	Logo       *ColorRef `json:"logo,omitempty" yaml:"logo,omitempty"`
}
type MainStylesPatch struct {
	BorderActive          *ColorRef `json:"borderActive,omitempty" yaml:"borderActive,omitempty"`
	BorderInactive        *ColorRef `json:"borderInactive,omitempty" yaml:"borderInactive,omitempty"`
	Title                 *ColorRef `json:"title,omitempty" yaml:"title,omitempty"`
	TableHeader           *ColorRef `json:"tableHeader,omitempty" yaml:"tableHeader,omitempty"`
	RowSelected           *ColorRef `json:"rowSelected,omitempty" yaml:"rowSelected,omitempty"`
	RowText               *ColorRef `json:"rowText,omitempty" yaml:"rowText,omitempty"`
	Footer                *ColorRef `json:"footer,omitempty" yaml:"footer,omitempty"`
	PanelBackground       *ColorRef `json:"panelBackground,omitempty" yaml:"panelBackground,omitempty"`
	ActionBarBackground   *ColorRef `json:"actionBarBackground,omitempty" yaml:"actionBarBackground,omitempty"`
	MessageRailBackground *ColorRef `json:"messageRailBackground,omitempty" yaml:"messageRailBackground,omitempty"`
	QueryBarBackground    *ColorRef `json:"queryBarBackground,omitempty" yaml:"queryBarBackground,omitempty"`
}
type FooterStylesPatch struct {
	StatusBackground   *ColorRef `json:"statusBackground,omitempty" yaml:"statusBackground,omitempty"`
	ShortcutBackground *ColorRef `json:"shortcutBackground,omitempty" yaml:"shortcutBackground,omitempty"`
	Key                *ColorRef `json:"key,omitempty" yaml:"key,omitempty"`
	Description        *ColorRef `json:"description,omitempty" yaml:"description,omitempty"`
	Separator          *ColorRef `json:"separator,omitempty" yaml:"separator,omitempty"`
}
type DialogStylesPatch struct {
	Border         *ColorRef `json:"border,omitempty" yaml:"border,omitempty"`
	Title          *ColorRef `json:"title,omitempty" yaml:"title,omitempty"`
	Body           *ColorRef `json:"body,omitempty" yaml:"body,omitempty"`
	BodyBackground *ColorRef `json:"bodyBackground,omitempty" yaml:"bodyBackground,omitempty"`
	OptionActive   *ColorRef `json:"optionActive,omitempty" yaml:"optionActive,omitempty"`
	OptionInactive *ColorRef `json:"optionInactive,omitempty" yaml:"optionInactive,omitempty"`
	Overlay        *ColorRef `json:"overlay,omitempty" yaml:"overlay,omitempty"`
	OverlayOpacity *uint8    `json:"overlayOpacity,omitempty" yaml:"overlayOpacity,omitempty"`
}
type ToastStylesPatch struct {
	Success    *ColorRef `json:"success,omitempty" yaml:"success,omitempty"`
	Error      *ColorRef `json:"error,omitempty" yaml:"error,omitempty"`
	Background *ColorRef `json:"background,omitempty" yaml:"background,omitempty"`
}
type TextStylesPatch struct {
	Info            *ColorRef `json:"info,omitempty" yaml:"info,omitempty"`
	Error           *ColorRef `json:"error,omitempty" yaml:"error,omitempty"`
	Success         *ColorRef `json:"success,omitempty" yaml:"success,omitempty"`
	Warning         *ColorRef `json:"warning,omitempty" yaml:"warning,omitempty"`
	Dim             *ColorRef `json:"dim,omitempty" yaml:"dim,omitempty"`
	HelpKey         *ColorRef `json:"helpKey,omitempty" yaml:"helpKey,omitempty"`
	HelpDescription *ColorRef `json:"helpDescription,omitempty" yaml:"helpDescription,omitempty"`
}

type TableStylesPatch struct {
	MarkedBackground *ColorRef `json:"markedBackground,omitempty" yaml:"markedBackground,omitempty"`
	ColumnForeground *ColorRef `json:"columnForeground,omitempty" yaml:"columnForeground,omitempty"`
	NameForeground   *ColorRef `json:"nameForeground,omitempty" yaml:"nameForeground,omitempty"`
}

type SafeFallbackStylesPatch struct {
	Normal *ColorRef `json:"normal,omitempty" yaml:"normal,omitempty"`
	Bold   *ColorRef `json:"bold,omitempty" yaml:"bold,omitempty"`
	Dim    *ColorRef `json:"dim,omitempty" yaml:"dim,omitempty"`
	Accent *ColorRef `json:"accent,omitempty" yaml:"accent,omitempty"`
	Error  *ColorRef `json:"error,omitempty" yaml:"error,omitempty"`
}

type ActionWindowPatch struct {
	Background *ColorRef `json:"background,omitempty" yaml:"background,omitempty"`
	Border     *ColorRef `json:"border,omitempty" yaml:"border,omitempty"`
}

type ActionFormInputPatch struct {
	Foreground *ColorRef `json:"foreground,omitempty" yaml:"foreground,omitempty"`
	Background *ColorRef `json:"background,omitempty" yaml:"background,omitempty"`
}

type ActionButtonPatch struct {
	Foreground *ColorRef `json:"foreground,omitempty" yaml:"foreground,omitempty"`
	Background *ColorRef `json:"background,omitempty" yaml:"background,omitempty"`
}

type ActionScopeStylesPatch struct {
	Window    *ActionWindowPatch    `json:"window,omitempty" yaml:"window,omitempty"`
	FormInput *ActionFormInputPatch `json:"formInput,omitempty" yaml:"formInput,omitempty"`
	Confirm   *ActionButtonPatch    `json:"confirm,omitempty" yaml:"confirm,omitempty"`
	Cancel    *ActionButtonPatch    `json:"cancel,omitempty" yaml:"cancel,omitempty"`
}

type ActionStylesPatch struct {
	Container *ActionScopeStylesPatch `json:"container,omitempty" yaml:"container,omitempty"`
	Image     *ActionScopeStylesPatch `json:"image,omitempty" yaml:"image,omitempty"`
}

type ThemePatch struct {
	Palette      *PalettePatch            `json:"palette,omitempty" yaml:"palette,omitempty"`
	Border       *BorderStylesPatch       `json:"border,omitempty" yaml:"border,omitempty"`
	Header       *HeaderStylesPatch       `json:"header,omitempty" yaml:"header,omitempty"`
	Main         *MainStylesPatch         `json:"main,omitempty" yaml:"main,omitempty"`
	Footer       *FooterStylesPatch       `json:"footer,omitempty" yaml:"footer,omitempty"`
	Dialog       *DialogStylesPatch       `json:"dialog,omitempty" yaml:"dialog,omitempty"`
	Toast        *ToastStylesPatch        `json:"toast,omitempty" yaml:"toast,omitempty"`
	Text         *TextStylesPatch         `json:"text,omitempty" yaml:"text,omitempty"`
	Table        *TableStylesPatch        `json:"table,omitempty" yaml:"table,omitempty"`
	SafeFallback *SafeFallbackStylesPatch `json:"safeFallback,omitempty" yaml:"safeFallback,omitempty"`
	Action       *ActionStylesPatch       `json:"action,omitempty" yaml:"action,omitempty"`
}

func DefaultTheme() *Theme {
	return &Theme{
		Palette: Palette{Primary: FallbackColorPrimary, Success: FallbackColorSuccess, Warning: FallbackColorWarning, Danger: FallbackColorDanger, Info: FallbackColorInfo, Accent: FallbackColorAccent, AccentSecondary: FallbackColorAccentSecondary, Foreground: FallbackColorForeground, ForegroundMuted: FallbackColorForegroundMuted, Background: FallbackColorBackground, BackgroundSubtle: FallbackColorBackgroundSubtle, BackgroundDeep: FallbackColorBackgroundDeep},
		Border:  BorderStyles{Active: TokenRef(ColorTokenPrimary), Inactive: TokenRef(ColorTokenForegroundMuted), Focused: TokenRef(ColorTokenSuccess), Kind: BorderRounded},
		Header:  HeaderStyles{Background: TokenRef(ColorTokenTransparent), Label: TokenRef(ColorTokenForegroundMuted), Value: TokenRef(ColorTokenForeground), Logo: TokenRef(ColorTokenInfo)},
		Main: MainStyles{
			BorderActive:          TokenRef(ColorTokenPrimary),
			BorderInactive:        TokenRef(ColorTokenForegroundMuted),
			Title:                 TokenRef(ColorTokenPrimary),
			TableHeader:           TokenRef(ColorTokenInfo),
			RowSelected:           TokenRef(ColorTokenInfo),
			RowText:               TokenRef(ColorTokenForeground),
			Footer:                TokenRef(ColorTokenForegroundMuted),
			PanelBackground:       TokenRef(ColorTokenBackgroundDeep),
			ActionBarBackground:   TokenRef(ColorTokenTransparent),
			MessageRailBackground: TokenRef(ColorTokenTransparent),
			QueryBarBackground:    TokenRef(ColorTokenTransparent),
		},
		Footer:       FooterStyles{StatusBackground: TokenRef(ColorTokenTransparent), ShortcutBackground: TokenRef(ColorTokenTransparent), Key: TokenRef(ColorTokenForeground), Description: TokenRef(ColorTokenForegroundMuted), Separator: TokenRef(ColorTokenBackgroundDeep)},
		Dialog:       DialogStyles{Border: TokenRef(ColorTokenDanger), Title: TokenRef(ColorTokenDanger), Body: TokenRef(ColorTokenForeground), BodyBackground: TokenRef(ColorTokenBackground), OptionActive: TokenRef(ColorTokenPrimary), OptionInactive: TokenRef(ColorTokenForegroundMuted), Overlay: TokenRef(ColorTokenBackground), OverlayOpacity: 80},
		Toast:        ToastStyles{Success: TokenRef(ColorTokenSuccess), Error: TokenRef(ColorTokenDanger), Background: TokenRef(ColorTokenTransparent)},
		Text:         TextStyles{Info: TokenRef(ColorTokenInfo), Error: TokenRef(ColorTokenDanger), Success: TokenRef(ColorTokenSuccess), Warning: TokenRef(ColorTokenWarning), Dim: TokenRef(ColorTokenForegroundMuted), HelpKey: TokenRef(ColorTokenPrimary), HelpDescription: TokenRef(ColorTokenForeground)},
		Table:        TableStyles{MarkedBackground: ValueRef(FallbackColorTableMarkedBackground), ColumnForeground: ValueRef(FallbackColorTableColumnForeground), NameForeground: ValueRef(FallbackColorTableNameForeground)},
		SafeFallback: SafeFallbackStyles{Normal: TokenRef(ColorTokenForeground), Bold: TokenRef(ColorTokenForeground), Dim: TokenRef(ColorTokenForegroundMuted), Accent: TokenRef(ColorTokenPrimary), Error: TokenRef(ColorTokenDanger)},
		Action: ActionStyles{
			Container: defaultActionScopeStyles(),
			Image:     defaultActionScopeStyles(),
		},
	}
}

func defaultActionScopeStyles() ActionScopeStyles {
	return ActionScopeStyles{
		Window: ActionWindow{
			Background: TokenRef(ColorTokenBackgroundSubtle),
			Border:     TokenRef(ColorTokenPrimary),
		},
		FormInput: ActionFormInput{
			Foreground: TokenRef(ColorTokenForeground),
			Background: TokenRef(ColorTokenBackgroundDeep),
		},
		Confirm: ActionButton{
			Foreground: TokenRef(ColorTokenSuccess),
			Background: TokenRef(ColorTokenTransparent),
		},
		Cancel: ActionButton{
			Foreground: TokenRef(ColorTokenForegroundMuted),
			Background: TokenRef(ColorTokenTransparent),
		},
	}
}

func (t *Theme) ResolveColor(ref ColorRef) string {
	if t == nil {
		return string(ref.Value)
	}
	switch ref.Token {
	case ColorTokenPrimary:
		return string(t.Palette.Primary)
	case ColorTokenSuccess:
		return string(t.Palette.Success)
	case ColorTokenWarning:
		return string(t.Palette.Warning)
	case ColorTokenDanger:
		return string(t.Palette.Danger)
	case ColorTokenInfo:
		return string(t.Palette.Info)
	case ColorTokenAccent:
		return string(t.Palette.Accent)
	case ColorTokenAccentSecondary:
		return string(t.Palette.AccentSecondary)
	case ColorTokenForeground:
		return string(t.Palette.Foreground)
	case ColorTokenForegroundMuted:
		return string(t.Palette.ForegroundMuted)
	case ColorTokenBackground:
		return string(t.Palette.Background)
	case ColorTokenBackgroundSubtle:
		return string(t.Palette.BackgroundSubtle)
	case ColorTokenBackgroundDeep:
		return string(t.Palette.BackgroundDeep)
	case ColorTokenTransparent:
		return string(FallbackColorTransparent)
	default:
		return string(ref.Value)
	}
}

func (p ThemePatch) Apply(target *Theme) {
	if target == nil {
		return
	}
	if p.Palette != nil {
		assign(&target.Palette.Primary, p.Palette.Primary)
		assign(&target.Palette.Success, p.Palette.Success)
		assign(&target.Palette.Warning, p.Palette.Warning)
		assign(&target.Palette.Danger, p.Palette.Danger)
		assign(&target.Palette.Info, p.Palette.Info)
		assign(&target.Palette.Accent, p.Palette.Accent)
		assign(&target.Palette.AccentSecondary, p.Palette.AccentSecondary)
		assign(&target.Palette.Foreground, p.Palette.Foreground)
		assign(&target.Palette.ForegroundMuted, p.Palette.ForegroundMuted)
		assign(&target.Palette.Background, p.Palette.Background)
		assign(&target.Palette.BackgroundSubtle, p.Palette.BackgroundSubtle)
		assign(&target.Palette.BackgroundDeep, p.Palette.BackgroundDeep)
	}
	if p.Border != nil {
		assign(&target.Border.Active, p.Border.Active)
		assign(&target.Border.Inactive, p.Border.Inactive)
		assign(&target.Border.Focused, p.Border.Focused)
		assign(&target.Border.Kind, p.Border.Kind)
	}
	if p.Header != nil {
		assign(&target.Header.Background, p.Header.Background)
		assign(&target.Header.Label, p.Header.Label)
		assign(&target.Header.Value, p.Header.Value)
		assign(&target.Header.Logo, p.Header.Logo)
	}
	if p.Main != nil {
		assign(&target.Main.BorderActive, p.Main.BorderActive)
		assign(&target.Main.BorderInactive, p.Main.BorderInactive)
		assign(&target.Main.Title, p.Main.Title)
		assign(&target.Main.TableHeader, p.Main.TableHeader)
		assign(&target.Main.RowSelected, p.Main.RowSelected)
		assign(&target.Main.RowText, p.Main.RowText)
		assign(&target.Main.Footer, p.Main.Footer)
		assign(&target.Main.PanelBackground, p.Main.PanelBackground)
		assign(&target.Main.ActionBarBackground, p.Main.ActionBarBackground)
		assign(&target.Main.MessageRailBackground, p.Main.MessageRailBackground)
		assign(&target.Main.QueryBarBackground, p.Main.QueryBarBackground)
	}
	if p.Footer != nil {
		assign(&target.Footer.StatusBackground, p.Footer.StatusBackground)
		assign(&target.Footer.ShortcutBackground, p.Footer.ShortcutBackground)
		assign(&target.Footer.Key, p.Footer.Key)
		assign(&target.Footer.Description, p.Footer.Description)
		assign(&target.Footer.Separator, p.Footer.Separator)
	}
	if p.Dialog != nil {
		assign(&target.Dialog.Border, p.Dialog.Border)
		assign(&target.Dialog.Title, p.Dialog.Title)
		assign(&target.Dialog.Body, p.Dialog.Body)
		assign(&target.Dialog.BodyBackground, p.Dialog.BodyBackground)
		assign(&target.Dialog.OptionActive, p.Dialog.OptionActive)
		assign(&target.Dialog.OptionInactive, p.Dialog.OptionInactive)
		assign(&target.Dialog.Overlay, p.Dialog.Overlay)
		assign(&target.Dialog.OverlayOpacity, p.Dialog.OverlayOpacity)
	}
	if p.Toast != nil {
		assign(&target.Toast.Success, p.Toast.Success)
		assign(&target.Toast.Error, p.Toast.Error)
		assign(&target.Toast.Background, p.Toast.Background)
	}
	if p.Text != nil {
		assign(&target.Text.Info, p.Text.Info)
		assign(&target.Text.Error, p.Text.Error)
		assign(&target.Text.Success, p.Text.Success)
		assign(&target.Text.Warning, p.Text.Warning)
		assign(&target.Text.Dim, p.Text.Dim)
		assign(&target.Text.HelpKey, p.Text.HelpKey)
		assign(&target.Text.HelpDescription, p.Text.HelpDescription)
	}
	if p.Table != nil {
		assign(&target.Table.MarkedBackground, p.Table.MarkedBackground)
		assign(&target.Table.ColumnForeground, p.Table.ColumnForeground)
		assign(&target.Table.NameForeground, p.Table.NameForeground)
	}
	if p.SafeFallback != nil {
		assign(&target.SafeFallback.Normal, p.SafeFallback.Normal)
		assign(&target.SafeFallback.Bold, p.SafeFallback.Bold)
		assign(&target.SafeFallback.Dim, p.SafeFallback.Dim)
		assign(&target.SafeFallback.Accent, p.SafeFallback.Accent)
		assign(&target.SafeFallback.Error, p.SafeFallback.Error)
	}
	if p.Action != nil {
		if p.Action.Container != nil {
			applyActionScopePatch(&target.Action.Container, p.Action.Container)
		}
		if p.Action.Image != nil {
			applyActionScopePatch(&target.Action.Image, p.Action.Image)
		}
	}
}

func applyActionScopePatch(target *ActionScopeStyles, p *ActionScopeStylesPatch) {
	if p.Window != nil {
		assign(&target.Window.Background, p.Window.Background)
		assign(&target.Window.Border, p.Window.Border)
	}
	if p.FormInput != nil {
		assign(&target.FormInput.Foreground, p.FormInput.Foreground)
		assign(&target.FormInput.Background, p.FormInput.Background)
	}
	if p.Confirm != nil {
		assign(&target.Confirm.Foreground, p.Confirm.Foreground)
		assign(&target.Confirm.Background, p.Confirm.Background)
	}
	if p.Cancel != nil {
		assign(&target.Cancel.Foreground, p.Cancel.Foreground)
		assign(&target.Cancel.Background, p.Cancel.Background)
	}
}

type loadedTheme struct {
	Meta  ThemeMetadata
	Theme *Theme
}

func loadThemeDocument(data []byte, source string, base *Theme) (*loadedTheme, error) {
	var document ThemeDocument
	if err := decodeJSONCStrict(data, &document); err != nil {
		return nil, fmt.Errorf(errParseThemeFormat, source, err)
	}
	result := *base
	document.Theme.Apply(&result)
	if err := ValidateTheme(&result); err != nil {
		return nil, fmt.Errorf(errValidateThemeSourceFormat, source, err)
	}
	return &loadedTheme{Meta: document.Meta, Theme: &result}, nil
}

func loadNamedTheme(name ThemeName, userDir string, base *Theme) (*loadedTheme, error) {
	if err := name.Validate(); err != nil {
		return nil, err
	}
	filename := string(name) + jsoncExtension
	if userDir != emptyValue {
		path := filepath.Join(userDir, filename)
		if data, err := os.ReadFile(path); err == nil {
			return loadThemeDocument(data, path, base)
		} else if !os.IsNotExist(err) {
			return nil, err
		}
	}
	data, err := fs.ReadFile(ThemeFS(), filename)
	if err != nil {
		return nil, fmt.Errorf(errThemeNotFoundFormat, name, err)
	}
	return loadThemeDocument(data, themeSourcePrefix+filename, base)
}

func (n ThemeName) Validate() error {
	if n == ThemeName(emptyValue) {
		return errors.New(errThemeNameRequired)
	}
	for _, char := range string(n) {
		if unicode.IsLetter(char) || unicode.IsDigit(char) || char == themeNameDash || char == themeNameUnderscore {
			continue
		}
		return fmt.Errorf(errThemeNameInvalidFormat, n)
	}
	return nil
}

func ListThemes() []string {
	entries, err := fs.ReadDir(ThemeFS(), currentFSDir)
	if err != nil {
		return nil
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), jsoncExtension) {
			names = append(names, strings.TrimSuffix(entry.Name(), jsoncExtension))
		}
	}
	sort.Strings(names)
	return names
}
