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

// ColorRef is a colour reference that can be either a palette token
// (any string without a "#" prefix) or a literal colour value (any
// string with a leading "#", e.g. "#3875d7"). The discriminator is
// purely lexical: a "#" prefix means value, anything else means token.
//
// This is a flat string rather than the {"token":..., "value":...}
// object wrapper the previous revision used because the wrapper was
// 99% boilerplate around a single field. Flat strings cut every theme
// JSONC file roughly in half and remove the only place where a JSON
// authoring mistake (both token and value set) was possible.
type ColorRef string

// TokenRef returns a ColorRef that resolves through the palette.
func TokenRef(token ColorToken) ColorRef {
	return ColorRef(token)
}

// ValueRef returns a ColorRef that holds a literal colour value.
func ValueRef(value Color) ColorRef {
	return ColorRef(value)
}

// IsToken reports whether the reference is a known palette token
// (true → resolve through Theme.Palette) or a literal colour value
// (false → pass through to ParseColor). The check is closed: only
// strings in the known palette token universe are treated as tokens,
// everything else is a literal. This means a literal like
// "rgb(250,240,230)" round-trips without needing a "#" prefix.
func (r ColorRef) IsToken() bool {
	switch r.Token() {
	case ColorTokenPrimary, ColorTokenSuccess, ColorTokenWarning, ColorTokenDanger,
		ColorTokenInfo, ColorTokenAccent, ColorTokenAccentSecondary,
		ColorTokenForeground, ColorTokenForegroundMuted,
		ColorTokenBackground, ColorTokenBackgroundSubtle, ColorTokenBackgroundDeep,
		ColorTokenTransparent:
		return true
	}
	return false
}

// Token returns the ColorToken the reference resolves through.
// Panics if the reference is a literal value rather than a token; call
// IsToken first when the input is user-supplied.
func (r ColorRef) Token() ColorToken {
	return ColorToken(r)
}

// Value returns the literal colour value. Panics if the reference is
// a token rather than a literal; call IsToken first when the input is
// user-supplied.
func (r ColorRef) Value() Color {
	return Color(r)
}

const (
	BorderRounded BorderKind = "rounded"
	BorderSingle  BorderKind = "single"
	BorderDouble  BorderKind = "double"
	BorderThick   BorderKind = "thick"
	BorderHidden  BorderKind = "hidden"
)

// ── Palette ───────────────────────────────────────────────────────
// Palette is the canonical named colour set. Every other theme field
// references one of these names via ColorRef.Token, so a single
// palette change cascades across chrome / surfaces / text / etc.

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

// ── Chrome (frame elements) ──────────────────────────────────────
// Chrome groups every "non-content" frame element: borders, titles,
// separators, dialog window borders. Theme authors editing the look
// of the app's framework touch only this section.

type ChromeStyles struct {
	BorderKind          BorderKind `json:"borderKind" yaml:"borderKind"`
	PanelBorderActive   ColorRef   `json:"panelBorderActive" yaml:"panelBorderActive"`
	PanelBorderInactive ColorRef   `json:"panelBorderInactive" yaml:"panelBorderInactive"`
	PanelBorderFocused  ColorRef   `json:"panelBorderFocused" yaml:"panelBorderFocused"`
	PanelTitle          ColorRef   `json:"panelTitle" yaml:"panelTitle"`
	PanelFooter         ColorRef   `json:"panelFooter" yaml:"panelFooter"`
	PanelRowSelected    ColorRef   `json:"panelRowSelected" yaml:"panelRowSelected"`
	PanelRowText        ColorRef   `json:"panelRowText" yaml:"panelRowText"`
	PanelTableHeader    ColorRef   `json:"panelTableHeader" yaml:"panelTableHeader"`
	DialogBorder        ColorRef   `json:"dialogBorder" yaml:"dialogBorder"`
	DialogTitle         ColorRef   `json:"dialogTitle" yaml:"dialogTitle"`
	DialogOptionActive  ColorRef   `json:"dialogOptionActive" yaml:"dialogOptionActive"`
	DialogOptionInactive ColorRef `json:"dialogOptionInactive" yaml:"dialogOptionInactive"`
	DialogOverlay       ColorRef   `json:"dialogOverlay" yaml:"dialogOverlay"`
	DialogOverlayOpacity uint8    `json:"dialogOverlayOpacity" yaml:"dialogOverlayOpacity"`
	HeaderLabel         ColorRef   `json:"headerLabel" yaml:"headerLabel"`
	HeaderValue         ColorRef   `json:"headerValue" yaml:"headerValue"`
	HeaderLogo          ColorRef   `json:"headerLogo" yaml:"headerLogo"`
	FooterKey           ColorRef   `json:"footerKey" yaml:"footerKey"`
	FooterDescription   ColorRef   `json:"footerDescription" yaml:"footerDescription"`
	FooterSeparator     ColorRef   `json:"footerSeparator" yaml:"footerSeparator"`
}

// ── Surfaces (background fills) ───────────────────────────────────
// Surfaces groups every fill colour used as a backdrop: panel
// background, action bar, message rail, query bar, dialog body,
// header, footer, toast. Together with chrome's rowSelected, every
// "what colour is this rectangle" question is answered here.

type SurfacesStyles struct {
	Panel       ColorRef `json:"panel" yaml:"panel"`
	ActionBar   ColorRef `json:"actionBar" yaml:"actionBar"`
	MessageRail ColorRef `json:"messageRail" yaml:"messageRail"`
	QueryBar    ColorRef `json:"queryBar" yaml:"queryBar"`
	DialogBody  ColorRef `json:"dialogBody" yaml:"dialogBody"`
	Header      ColorRef `json:"header" yaml:"header"`
	Footer      ColorRef `json:"footer" yaml:"footer"`
	Toast       ColorRef `json:"toast" yaml:"toast"`
	RowSelected ColorRef `json:"rowSelected" yaml:"rowSelected"`
}

// ── Text (foreground typography) ──────────────────────────────────
// Text groups every foreground colour used for content text:
// semantic (info/success/warning/error/dim), help, dialog body.
// Table header / row text live in chrome since they are part of the
// table widget's chrome.

type TextStyles struct {
	Info            ColorRef `json:"info" yaml:"info"`
	Success         ColorRef `json:"success" yaml:"success"`
	Warning         ColorRef `json:"warning" yaml:"warning"`
	Error           ColorRef `json:"error" yaml:"error"`
	Dim             ColorRef `json:"dim" yaml:"dim"`
	HelpKey         ColorRef `json:"helpKey" yaml:"helpKey"`
	HelpDescription ColorRef `json:"helpDescription" yaml:"helpDescription"`
	DialogBody      ColorRef `json:"dialogBody" yaml:"dialogBody"`
}

// ── Feedback (transient + fallback) ───────────────────────────────
// Feedback groups the transient notification palette (toasts) and
// the safe-fallback palette (used when a component style has not
// been resolved yet). Both are "what does the user see when the
// normal flow is broken / needs attention" — colocating them makes
// that obvious.

type FeedbackStyles struct {
	ToastSuccess ColorRef `json:"toastSuccess" yaml:"toastSuccess"`
	ToastError   ColorRef `json:"toastError" yaml:"toastError"`
	ToastInfo    ColorRef `json:"toastInfo" yaml:"toastInfo"`
	ToastWarning ColorRef `json:"toastWarning" yaml:"toastWarning"`
	SafeNormal   ColorRef `json:"safeNormal" yaml:"safeNormal"`
	SafeBold     ColorRef `json:"safeBold" yaml:"safeBold"`
	SafeDim      ColorRef `json:"safeDim" yaml:"safeDim"`
	SafeAccent   ColorRef `json:"safeAccent" yaml:"safeAccent"`
	SafeError    ColorRef `json:"safeError" yaml:"safeError"`
}

// ── Data (table content) ──────────────────────────────────────────
// Data groups the table-content palette: marked row background,
// column foreground, name foreground. The structural table chrome
// (row selected, table header, row text) lives in chrome + surfaces
// because those are frame-related; the literal palette tokens for
// the data inside the table live here.

type DataStyles struct {
	MarkedBackground ColorRef `json:"markedBackground" yaml:"markedBackground"`
	ColumnForeground ColorRef `json:"columnForeground" yaml:"columnForeground"`
	NameForeground   ColorRef `json:"nameForeground" yaml:"nameForeground"`
}

// ── Theme (composition) ──────────────────────────────────────────

type Theme struct {
	Palette Palette          `json:"palette" yaml:"palette"`
	Chrome  ChromeStyles     `json:"chrome" yaml:"chrome"`
	Surfaces SurfacesStyles  `json:"surfaces" yaml:"surfaces"`
	Text    TextStyles       `json:"text" yaml:"text"`
	Feedback FeedbackStyles  `json:"feedback" yaml:"feedback"`
	Data    DataStyles       `json:"data" yaml:"data"`
	Action  ActionStyles     `json:"action" yaml:"action"`
}

// ── Patches (for ThemePatch.Apply) ─────────────────────────────────

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

type ChromeStylesPatch struct {
	BorderKind           *BorderKind `json:"borderKind,omitempty" yaml:"borderKind,omitempty"`
	PanelBorderActive    *ColorRef   `json:"panelBorderActive,omitempty" yaml:"panelBorderActive,omitempty"`
	PanelBorderInactive  *ColorRef   `json:"panelBorderInactive,omitempty" yaml:"panelBorderInactive,omitempty"`
	PanelBorderFocused   *ColorRef   `json:"panelBorderFocused,omitempty" yaml:"panelBorderFocused,omitempty"`
	PanelTitle           *ColorRef   `json:"panelTitle,omitempty" yaml:"panelTitle,omitempty"`
	PanelFooter          *ColorRef   `json:"panelFooter,omitempty" yaml:"panelFooter,omitempty"`
	PanelRowSelected     *ColorRef   `json:"panelRowSelected,omitempty" yaml:"panelRowSelected,omitempty"`
	PanelRowText         *ColorRef   `json:"panelRowText,omitempty" yaml:"panelRowText,omitempty"`
	PanelTableHeader     *ColorRef   `json:"panelTableHeader,omitempty" yaml:"panelTableHeader,omitempty"`
	DialogBorder         *ColorRef   `json:"dialogBorder,omitempty" yaml:"dialogBorder,omitempty"`
	DialogTitle          *ColorRef   `json:"dialogTitle,omitempty" yaml:"dialogTitle,omitempty"`
	DialogOptionActive   *ColorRef   `json:"dialogOptionActive,omitempty" yaml:"dialogOptionActive,omitempty"`
	DialogOptionInactive *ColorRef   `json:"dialogOptionInactive,omitempty" yaml:"dialogOptionInactive,omitempty"`
	DialogOverlay        *ColorRef   `json:"dialogOverlay,omitempty" yaml:"dialogOverlay,omitempty"`
	DialogOverlayOpacity *uint8      `json:"dialogOverlayOpacity,omitempty" yaml:"dialogOverlayOpacity,omitempty"`
	HeaderLabel          *ColorRef   `json:"headerLabel,omitempty" yaml:"headerLabel,omitempty"`
	HeaderValue          *ColorRef   `json:"headerValue,omitempty" yaml:"headerValue,omitempty"`
	HeaderLogo           *ColorRef   `json:"headerLogo,omitempty" yaml:"headerLogo,omitempty"`
	FooterKey            *ColorRef   `json:"footerKey,omitempty" yaml:"footerKey,omitempty"`
	FooterDescription    *ColorRef   `json:"footerDescription,omitempty" yaml:"footerDescription,omitempty"`
	FooterSeparator      *ColorRef   `json:"footerSeparator,omitempty" yaml:"footerSeparator,omitempty"`
}

type SurfacesStylesPatch struct {
	Panel       *ColorRef `json:"panel,omitempty" yaml:"panel,omitempty"`
	ActionBar   *ColorRef `json:"actionBar,omitempty" yaml:"actionBar,omitempty"`
	MessageRail *ColorRef `json:"messageRail,omitempty" yaml:"messageRail,omitempty"`
	QueryBar    *ColorRef `json:"queryBar,omitempty" yaml:"queryBar,omitempty"`
	DialogBody  *ColorRef `json:"dialogBody,omitempty" yaml:"dialogBody,omitempty"`
	Header      *ColorRef `json:"header,omitempty" yaml:"header,omitempty"`
	Footer      *ColorRef `json:"footer,omitempty" yaml:"footer,omitempty"`
	Toast       *ColorRef `json:"toast,omitempty" yaml:"toast,omitempty"`
	RowSelected *ColorRef `json:"rowSelected,omitempty" yaml:"rowSelected,omitempty"`
}

type TextStylesPatch struct {
	Info            *ColorRef `json:"info,omitempty" yaml:"info,omitempty"`
	Success         *ColorRef `json:"success,omitempty" yaml:"success,omitempty"`
	Warning         *ColorRef `json:"warning,omitempty" yaml:"warning,omitempty"`
	Error           *ColorRef `json:"error,omitempty" yaml:"error,omitempty"`
	Dim             *ColorRef `json:"dim,omitempty" yaml:"dim,omitempty"`
	HelpKey         *ColorRef `json:"helpKey,omitempty" yaml:"helpKey,omitempty"`
	HelpDescription *ColorRef `json:"helpDescription,omitempty" yaml:"helpDescription,omitempty"`
	DialogBody      *ColorRef `json:"dialogBody,omitempty" yaml:"dialogBody,omitempty"`
}

type FeedbackStylesPatch struct {
	ToastSuccess *ColorRef `json:"toastSuccess,omitempty" yaml:"toastSuccess,omitempty"`
	ToastError   *ColorRef `json:"toastError,omitempty" yaml:"toastError,omitempty"`
	ToastInfo    *ColorRef `json:"toastInfo,omitempty" yaml:"toastInfo,omitempty"`
	ToastWarning *ColorRef `json:"toastWarning,omitempty" yaml:"toastWarning,omitempty"`
	SafeNormal   *ColorRef `json:"safeNormal,omitempty" yaml:"safeNormal,omitempty"`
	SafeBold     *ColorRef `json:"safeBold,omitempty" yaml:"safeBold,omitempty"`
	SafeDim      *ColorRef `json:"safeDim,omitempty" yaml:"safeDim,omitempty"`
	SafeAccent   *ColorRef `json:"safeAccent,omitempty" yaml:"safeAccent,omitempty"`
	SafeError    *ColorRef `json:"safeError,omitempty" yaml:"safeError,omitempty"`
}

type DataStylesPatch struct {
	MarkedBackground *ColorRef `json:"markedBackground,omitempty" yaml:"markedBackground,omitempty"`
	ColumnForeground *ColorRef `json:"columnForeground,omitempty" yaml:"columnForeground,omitempty"`
	NameForeground   *ColorRef `json:"nameForeground,omitempty" yaml:"nameForeground,omitempty"`
}

// ThemePatch is the user-override shape; each section is independently
// patchable so user themes can change only what they care about.
type ThemePatch struct {
	Palette  *PalettePatch  `json:"palette,omitempty" yaml:"palette,omitempty"`
	Chrome   *ChromeStylesPatch `json:"chrome,omitempty" yaml:"chrome,omitempty"`
	Surfaces *SurfacesStylesPatch `json:"surfaces,omitempty" yaml:"surfaces,omitempty"`
	Text     *TextStylesPatch `json:"text,omitempty" yaml:"text,omitempty"`
	Feedback *FeedbackStylesPatch `json:"feedback,omitempty" yaml:"feedback,omitempty"`
	Data     *DataStylesPatch `json:"data,omitempty" yaml:"data,omitempty"`
	Action   *ActionStylesPatch `json:"action,omitempty" yaml:"action,omitempty"`
}

func DefaultTheme() *Theme {
	return &Theme{
		Palette: Palette{
			Primary:          FallbackColorPrimary,
			Success:          FallbackColorSuccess,
			Warning:          FallbackColorWarning,
			Danger:           FallbackColorDanger,
			Info:             FallbackColorInfo,
			Accent:           FallbackColorAccent,
			AccentSecondary:  FallbackColorAccentSecondary,
			Foreground:       FallbackColorForeground,
			ForegroundMuted:  FallbackColorForegroundMuted,
			Background:       FallbackColorBackground,
			BackgroundSubtle: FallbackColorBackgroundSubtle,
			BackgroundDeep:   FallbackColorBackgroundDeep,
		},
		Chrome:   defaultChromeStyles(),
		Surfaces: defaultSurfacesStyles(),
		Text:     defaultTextStyles(),
		Feedback: defaultFeedbackStyles(),
		Data:     defaultDataStyles(),
		Action: ActionStyles{
			Container: defaultActionScopeStyles(),
			Image:     defaultActionScopeStyles(),
			Volume:    defaultActionScopeStyles(),
			Network:   defaultActionScopeStyles(),
		},
	}
}

func defaultChromeStyles() ChromeStyles {
	return ChromeStyles{
		BorderKind:           BorderRounded,
		PanelBorderActive:    TokenRef(ColorTokenPrimary),
		PanelBorderInactive:  TokenRef(ColorTokenForegroundMuted),
		PanelBorderFocused:   TokenRef(ColorTokenSuccess),
		PanelTitle:           TokenRef(ColorTokenPrimary),
		PanelFooter:          TokenRef(ColorTokenForegroundMuted),
		PanelRowSelected:     TokenRef(ColorTokenInfo),
		PanelRowText:         TokenRef(ColorTokenForeground),
		PanelTableHeader:     TokenRef(ColorTokenInfo),
		DialogBorder:         TokenRef(ColorTokenDanger),
		DialogTitle:          TokenRef(ColorTokenDanger),
		DialogOptionActive:   TokenRef(ColorTokenPrimary),
		DialogOptionInactive: TokenRef(ColorTokenForegroundMuted),
		DialogOverlay:        TokenRef(ColorTokenBackground),
		DialogOverlayOpacity: 80,
		HeaderLabel:          TokenRef(ColorTokenForegroundMuted),
		HeaderValue:          TokenRef(ColorTokenForeground),
		HeaderLogo:           TokenRef(ColorTokenInfo),
		FooterKey:            TokenRef(ColorTokenForeground),
		FooterDescription:    TokenRef(ColorTokenForegroundMuted),
		FooterSeparator:      TokenRef(ColorTokenForegroundMuted),
	}
}

func defaultSurfacesStyles() SurfacesStyles {
	return SurfacesStyles{
		Panel:       TokenRef(ColorTokenBackgroundDeep),
		ActionBar:   TokenRef(ColorTokenTransparent),
		MessageRail: TokenRef(ColorTokenTransparent),
		QueryBar:    TokenRef(ColorTokenTransparent),
		DialogBody:  TokenRef(ColorTokenBackground),
		Header:      TokenRef(ColorTokenTransparent),
		Footer:      TokenRef(ColorTokenTransparent),
		Toast:       TokenRef(ColorTokenTransparent),
		RowSelected: TokenRef(ColorTokenInfo),
	}
}

func defaultTextStyles() TextStyles {
	return TextStyles{
		Info:            TokenRef(ColorTokenPrimary),
		Success:         TokenRef(ColorTokenSuccess),
		Warning:         TokenRef(ColorTokenWarning),
		Error:           TokenRef(ColorTokenDanger),
		Dim:             TokenRef(ColorTokenForegroundMuted),
		HelpKey:         TokenRef(ColorTokenPrimary),
		HelpDescription: TokenRef(ColorTokenForeground),
		DialogBody:      TokenRef(ColorTokenForeground),
	}
}

func defaultFeedbackStyles() FeedbackStyles {
	return FeedbackStyles{
		ToastSuccess: TokenRef(ColorTokenSuccess),
		ToastError:   TokenRef(ColorTokenDanger),
		ToastInfo:    TokenRef(ColorTokenInfo),
		ToastWarning: TokenRef(ColorTokenWarning),
		SafeNormal:   TokenRef(ColorTokenForeground),
		SafeBold:     TokenRef(ColorTokenForeground),
		SafeDim:      TokenRef(ColorTokenForegroundMuted),
		SafeAccent:   TokenRef(ColorTokenPrimary),
		SafeError:    TokenRef(ColorTokenDanger),
	}
}

func defaultDataStyles() DataStyles {
	return DataStyles{
		MarkedBackground: ValueRef(FallbackColorTableMarkedBackground),
		ColumnForeground: ValueRef(FallbackColorTableColumnForeground),
		NameForeground:   ValueRef(FallbackColorTableNameForeground),
	}
}

func (t *Theme) ResolveColor(ref ColorRef) string {
	if t == nil {
		return string(ref)
	}
	if !ref.IsToken() {
		return string(ref)
	}
	switch ref.Token() {
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
		return string(ref)
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
	if p.Chrome != nil {
		assign(&target.Chrome.BorderKind, p.Chrome.BorderKind)
		assign(&target.Chrome.PanelBorderActive, p.Chrome.PanelBorderActive)
		assign(&target.Chrome.PanelBorderInactive, p.Chrome.PanelBorderInactive)
		assign(&target.Chrome.PanelBorderFocused, p.Chrome.PanelBorderFocused)
		assign(&target.Chrome.PanelTitle, p.Chrome.PanelTitle)
		assign(&target.Chrome.PanelFooter, p.Chrome.PanelFooter)
		assign(&target.Chrome.PanelRowSelected, p.Chrome.PanelRowSelected)
		assign(&target.Chrome.PanelRowText, p.Chrome.PanelRowText)
		assign(&target.Chrome.PanelTableHeader, p.Chrome.PanelTableHeader)
		assign(&target.Chrome.DialogBorder, p.Chrome.DialogBorder)
		assign(&target.Chrome.DialogTitle, p.Chrome.DialogTitle)
		assign(&target.Chrome.DialogOptionActive, p.Chrome.DialogOptionActive)
		assign(&target.Chrome.DialogOptionInactive, p.Chrome.DialogOptionInactive)
		assign(&target.Chrome.DialogOverlay, p.Chrome.DialogOverlay)
		assign(&target.Chrome.DialogOverlayOpacity, p.Chrome.DialogOverlayOpacity)
		assign(&target.Chrome.HeaderLabel, p.Chrome.HeaderLabel)
		assign(&target.Chrome.HeaderValue, p.Chrome.HeaderValue)
		assign(&target.Chrome.HeaderLogo, p.Chrome.HeaderLogo)
		assign(&target.Chrome.FooterKey, p.Chrome.FooterKey)
		assign(&target.Chrome.FooterDescription, p.Chrome.FooterDescription)
		assign(&target.Chrome.FooterSeparator, p.Chrome.FooterSeparator)
	}
	if p.Surfaces != nil {
		assign(&target.Surfaces.Panel, p.Surfaces.Panel)
		assign(&target.Surfaces.ActionBar, p.Surfaces.ActionBar)
		assign(&target.Surfaces.MessageRail, p.Surfaces.MessageRail)
		assign(&target.Surfaces.QueryBar, p.Surfaces.QueryBar)
		assign(&target.Surfaces.DialogBody, p.Surfaces.DialogBody)
		assign(&target.Surfaces.Header, p.Surfaces.Header)
		assign(&target.Surfaces.Footer, p.Surfaces.Footer)
		assign(&target.Surfaces.Toast, p.Surfaces.Toast)
		assign(&target.Surfaces.RowSelected, p.Surfaces.RowSelected)
	}
	if p.Text != nil {
		assign(&target.Text.Info, p.Text.Info)
		assign(&target.Text.Success, p.Text.Success)
		assign(&target.Text.Warning, p.Text.Warning)
		assign(&target.Text.Error, p.Text.Error)
		assign(&target.Text.Dim, p.Text.Dim)
		assign(&target.Text.HelpKey, p.Text.HelpKey)
		assign(&target.Text.HelpDescription, p.Text.HelpDescription)
		assign(&target.Text.DialogBody, p.Text.DialogBody)
	}
	if p.Feedback != nil {
		assign(&target.Feedback.ToastSuccess, p.Feedback.ToastSuccess)
		assign(&target.Feedback.ToastError, p.Feedback.ToastError)
		assign(&target.Feedback.ToastInfo, p.Feedback.ToastInfo)
		assign(&target.Feedback.ToastWarning, p.Feedback.ToastWarning)
		assign(&target.Feedback.SafeNormal, p.Feedback.SafeNormal)
		assign(&target.Feedback.SafeBold, p.Feedback.SafeBold)
		assign(&target.Feedback.SafeDim, p.Feedback.SafeDim)
		assign(&target.Feedback.SafeAccent, p.Feedback.SafeAccent)
		assign(&target.Feedback.SafeError, p.Feedback.SafeError)
	}
	if p.Data != nil {
		assign(&target.Data.MarkedBackground, p.Data.MarkedBackground)
		assign(&target.Data.ColumnForeground, p.Data.ColumnForeground)
		assign(&target.Data.NameForeground, p.Data.NameForeground)
	}
	if p.Action != nil {
		if p.Action.Container != nil {
			applyActionScopePatch(&target.Action.Container, p.Action.Container)
		}
		if p.Action.Image != nil {
			applyActionScopePatch(&target.Action.Image, p.Action.Image)
		}
		if p.Action.Volume != nil {
			applyActionScopePatch(&target.Action.Volume, p.Action.Volume)
		}
		if p.Action.Network != nil {
			applyActionScopePatch(&target.Action.Network, p.Action.Network)
		}
	}
}

// loadThemeDocument / loadNamedTheme / ThemeName.Validate / ListThemes
// — unchanged from the previous round; they live below as plumbing.

type ThemeMetadata struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"description"`
}

type ThemeDocument struct {
	Meta  ThemeMetadata `json:"meta" yaml:"meta"`
	Theme ThemePatch    `json:"theme" yaml:"theme"`
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