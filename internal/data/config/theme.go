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
	Green      Color `json:"green" yaml:"green"`
	Cyan       Color `json:"cyan" yaml:"cyan"`
	Blue       Color `json:"blue" yaml:"blue"`
	Red        Color `json:"red" yaml:"red"`
	Yellow     Color `json:"yellow" yaml:"yellow"`
	Orange     Color `json:"orange" yaml:"orange"`
	Purple     Color `json:"purple" yaml:"purple"`
	White      Color `json:"white" yaml:"white"`
	Gray       Color `json:"gray" yaml:"gray"`
	Dark       Color `json:"dark" yaml:"dark"`
	Surface    Color `json:"surface" yaml:"surface"`
	Background Color `json:"background" yaml:"background"`
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
	BorderActive   ColorRef `json:"borderActive" yaml:"borderActive"`
	BorderInactive ColorRef `json:"borderInactive" yaml:"borderInactive"`
	Title          ColorRef `json:"title" yaml:"title"`
	TableHeader    ColorRef `json:"tableHeader" yaml:"tableHeader"`
	RowSelected    ColorRef `json:"rowSelected" yaml:"rowSelected"`
	RowText        ColorRef `json:"rowText" yaml:"rowText"`
	Footer         ColorRef `json:"footer" yaml:"footer"`
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
	Green      *Color `json:"green,omitempty" yaml:"green,omitempty"`
	Cyan       *Color `json:"cyan,omitempty" yaml:"cyan,omitempty"`
	Blue       *Color `json:"blue,omitempty" yaml:"blue,omitempty"`
	Red        *Color `json:"red,omitempty" yaml:"red,omitempty"`
	Yellow     *Color `json:"yellow,omitempty" yaml:"yellow,omitempty"`
	Orange     *Color `json:"orange,omitempty" yaml:"orange,omitempty"`
	Purple     *Color `json:"purple,omitempty" yaml:"purple,omitempty"`
	White      *Color `json:"white,omitempty" yaml:"white,omitempty"`
	Gray       *Color `json:"gray,omitempty" yaml:"gray,omitempty"`
	Dark       *Color `json:"dark,omitempty" yaml:"dark,omitempty"`
	Surface    *Color `json:"surface,omitempty" yaml:"surface,omitempty"`
	Background *Color `json:"background,omitempty" yaml:"background,omitempty"`
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
	BorderActive   *ColorRef `json:"borderActive,omitempty" yaml:"borderActive,omitempty"`
	BorderInactive *ColorRef `json:"borderInactive,omitempty" yaml:"borderInactive,omitempty"`
	Title          *ColorRef `json:"title,omitempty" yaml:"title,omitempty"`
	TableHeader    *ColorRef `json:"tableHeader,omitempty" yaml:"tableHeader,omitempty"`
	RowSelected    *ColorRef `json:"rowSelected,omitempty" yaml:"rowSelected,omitempty"`
	RowText        *ColorRef `json:"rowText,omitempty" yaml:"rowText,omitempty"`
	Footer         *ColorRef `json:"footer,omitempty" yaml:"footer,omitempty"`
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
}

func DefaultTheme() *Theme {
	return &Theme{
		Palette:      Palette{Green: FallbackColorGreen, Cyan: FallbackColorCyan, Blue: FallbackColorBlue, Red: FallbackColorRed, Yellow: FallbackColorYellow, Orange: FallbackColorOrange, Purple: FallbackColorPurple, White: FallbackColorWhite, Gray: FallbackColorGray, Dark: FallbackColorDark, Surface: FallbackColorSurface, Background: FallbackColorBackground},
		Border:       BorderStyles{Active: TokenRef(ColorTokenCyan), Inactive: TokenRef(ColorTokenGray), Focused: TokenRef(ColorTokenGreen), Kind: BorderRounded},
		Header:       HeaderStyles{Background: TokenRef(ColorTokenDark), Label: TokenRef(ColorTokenGray), Value: TokenRef(ColorTokenWhite), Logo: TokenRef(ColorTokenCyan)},
		Main:         MainStyles{BorderActive: TokenRef(ColorTokenCyan), BorderInactive: TokenRef(ColorTokenGray), Title: TokenRef(ColorTokenCyan), TableHeader: TokenRef(ColorTokenBlue), RowSelected: TokenRef(ColorTokenBlue), RowText: TokenRef(ColorTokenWhite), Footer: TokenRef(ColorTokenGray)},
		Footer:       FooterStyles{StatusBackground: TokenRef(ColorTokenSurface), ShortcutBackground: TokenRef(ColorTokenDark), Key: TokenRef(ColorTokenWhite), Description: TokenRef(ColorTokenGray), Separator: TokenRef(ColorTokenSurface)},
		Dialog:       DialogStyles{Border: TokenRef(ColorTokenRed), Title: TokenRef(ColorTokenRed), Body: TokenRef(ColorTokenWhite), OptionActive: TokenRef(ColorTokenCyan), OptionInactive: TokenRef(ColorTokenGray), Overlay: TokenRef(ColorTokenBackground), OverlayOpacity: 80},
		Toast:        ToastStyles{Success: TokenRef(ColorTokenGreen), Error: TokenRef(ColorTokenRed), Background: TokenRef(ColorTokenDark)},
		Text:         TextStyles{Info: TokenRef(ColorTokenCyan), Error: TokenRef(ColorTokenRed), Success: TokenRef(ColorTokenGreen), Warning: TokenRef(ColorTokenYellow), Dim: TokenRef(ColorTokenGray), HelpKey: TokenRef(ColorTokenCyan), HelpDescription: TokenRef(ColorTokenWhite)},
		Table:        TableStyles{MarkedBackground: ValueRef(FallbackColorTableMarkedBackground), ColumnForeground: ValueRef(FallbackColorTableColumnForeground), NameForeground: ValueRef(FallbackColorTableNameForeground)},
		SafeFallback: SafeFallbackStyles{Normal: TokenRef(ColorTokenWhite), Bold: TokenRef(ColorTokenWhite), Dim: TokenRef(ColorTokenGray), Accent: TokenRef(ColorTokenCyan), Error: TokenRef(ColorTokenRed)},
	}
}

func (t *Theme) ResolveColor(ref ColorRef) string {
	if t == nil {
		return string(ref.Value)
	}
	switch ref.Token {
	case ColorTokenGreen:
		return string(t.Palette.Green)
	case ColorTokenCyan:
		return string(t.Palette.Cyan)
	case ColorTokenBlue:
		return string(t.Palette.Blue)
	case ColorTokenRed:
		return string(t.Palette.Red)
	case ColorTokenYellow:
		return string(t.Palette.Yellow)
	case ColorTokenOrange:
		return string(t.Palette.Orange)
	case ColorTokenPurple:
		return string(t.Palette.Purple)
	case ColorTokenWhite:
		return string(t.Palette.White)
	case ColorTokenGray:
		return string(t.Palette.Gray)
	case ColorTokenDark:
		return string(t.Palette.Dark)
	case ColorTokenSurface:
		return string(t.Palette.Surface)
	case ColorTokenBackground:
		return string(t.Palette.Background)
	default:
		return string(ref.Value)
	}
}

func (p ThemePatch) Apply(target *Theme) {
	if target == nil {
		return
	}
	if p.Palette != nil {
		assign(&target.Palette.Green, p.Palette.Green)
		assign(&target.Palette.Cyan, p.Palette.Cyan)
		assign(&target.Palette.Blue, p.Palette.Blue)
		assign(&target.Palette.Red, p.Palette.Red)
		assign(&target.Palette.Yellow, p.Palette.Yellow)
		assign(&target.Palette.Orange, p.Palette.Orange)
		assign(&target.Palette.Purple, p.Palette.Purple)
		assign(&target.Palette.White, p.Palette.White)
		assign(&target.Palette.Gray, p.Palette.Gray)
		assign(&target.Palette.Dark, p.Palette.Dark)
		assign(&target.Palette.Surface, p.Palette.Surface)
		assign(&target.Palette.Background, p.Palette.Background)
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
	if err := validateThemeName(name); err != nil {
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

func validateThemeName(name ThemeName) error {
	if name == ThemeName(emptyValue) {
		return errors.New(errThemeNameRequired)
	}
	for _, char := range string(name) {
		if unicode.IsLetter(char) || unicode.IsDigit(char) || char == themeNameDash || char == themeNameUnderscore {
			continue
		}
		return fmt.Errorf(errThemeNameInvalidFormat, name)
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
