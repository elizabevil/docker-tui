package config

import (
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/elizabevil/docker-tui/internal/utils"
)

// ── Palette: raw color values ──────────────────────────────────

type Palette struct {
	Green      string `json:"green"`
	Cyan       string `json:"cyan"`
	Blue       string `json:"blue"`
	Red        string `json:"red"`
	Yellow     string `json:"yellow"`
	Orange     string `json:"orange"`
	Purple     string `json:"purple"`
	White      string `json:"white"`
	Gray       string `json:"gray"`
	Dark       string `json:"dark"`
	Surface    string `json:"surface"`
	Background string `json:"background"`
}

func defaultPalette() Palette {
	return Palette{
		Green: "#2ecc71", Cyan: "#00bcd4", Blue: "#42a5f5",
		Red: "#ef5350", Yellow: "#ffca28", Orange: "#ff9800",
		Purple: "#ab47bc", White: "#ffffff", Gray: "#546e7a",
		Dark: "#1a1a2e", Surface: "#16213e", Background: "#0d1117",
	}
}

// ── Global style groups ───────────────────────────────────────

type BorderStyles struct {
	Active   string `json:"active"`
	Inactive string `json:"inactive"`
	Focused  string `json:"focused"`
	Style    string `json:"style"`
}

type HeaderStyles struct {
	BarBG string `json:"bar_bg"`
	Label string `json:"label"`
	Value string `json:"value"`
	Logo  string `json:"logo"`
}

type MainStyles struct {
	BorderActive   string `json:"border_active"`
	BorderInactive string `json:"border_inactive"`
	Title          string `json:"title"`
	HeaderFG       string `json:"header_fg"`
	RowSelected    string `json:"row_selected"`
	RowText        string `json:"row_text"`
	Footer         string `json:"footer"`
}

type FooterStyles struct {
	StatusBarBG    string `json:"status_bar_bg"`
	ShortcutsBarBG string `json:"shortcuts_bar_bg"`
	Key            string `json:"key"`
	Desc           string `json:"desc"`
	Separator      string `json:"separator"`
}

type DialogStyles struct {
	Border    string `json:"border"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	OptionOn  string `json:"option_on"`
	OptionOff string `json:"option_off"`
}

type ToastStyles struct {
	Success    string `json:"success"`
	Error      string `json:"error"`
	Background string `json:"background"`
}

type TextStyles struct {
	Info     string `json:"info"`
	Error    string `json:"error"`
	Success  string `json:"success"`
	Warning  string `json:"warning"`
	Dim      string `json:"dim"`
	HelpKey  string `json:"help_key"`
	HelpDesc string `json:"help_desc"`
}

// ── Theme: palette + global style groups + component styles ────

type Theme struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Colors      Palette      `json:"colors"`
	Border      BorderStyles `json:"border"`
	Header      HeaderStyles `json:"header"`
	Main        MainStyles   `json:"main"`
	Footer      FooterStyles `json:"footer"`
	Dialog      DialogStyles `json:"dialog"`
	Toast       ToastStyles  `json:"toast"`
	Text        TextStyles   `json:"text"`
}

func DefaultTheme() *Theme {
	return &Theme{
		Name: "Default Dark", Description: "k9s-inspired dark theme",
		Colors: defaultPalette(),
		Border: BorderStyles{Active: "cyan", Inactive: "gray", Focused: "green", Style: "rounded"},
		Header: HeaderStyles{BarBG: "#1a1a2e", Label: "#546e7a", Value: "#ffffff", Logo: "#00bcd4"},
		Main:   MainStyles{BorderActive: "cyan", BorderInactive: "gray", Title: "cyan", HeaderFG: "blue", RowSelected: "blue", RowText: "white", Footer: "gray"},
		Footer: FooterStyles{StatusBarBG: "#16213e", ShortcutsBarBG: "#1a1a2e", Key: "white", Desc: "gray", Separator: "#16213e"},
		Dialog: DialogStyles{Border: "red", Title: "red", Body: "white", OptionOn: "cyan", OptionOff: "gray"},
		Toast:  ToastStyles{Success: "green", Error: "red", Background: "#1a1a2e"},
		Text:   TextStyles{Info: "cyan", Error: "red", Success: "green", Warning: "yellow", Dim: "gray", HelpKey: "cyan", HelpDesc: "white"},
	}
}

func LoadTheme(name string) (*Theme, error) {
	if name == "" {
		return DefaultTheme(), nil
	}
	themeFS := ThemeFS()
	filename := name
	if !strings.HasSuffix(filename, ".json") && !strings.HasSuffix(filename, ".jsonc") {
		filename = name + ".jsonc"
	}
	data, err := fs.ReadFile(themeFS, filename)
	if err == nil {
		return parseTheme(data)
	}
	// Fallback: .jsonc → .json
	if strings.HasSuffix(filename, ".jsonc") {
		alt := filename[:len(filename)-1] // foo.jsonc → foo.json
		data, err = fs.ReadFile(themeFS, alt)
		if err == nil {
			return parseTheme(data)
		}
	}
	data, err = os.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("theme %q not found: %w", name, err)
	}
	return parseTheme(data)
}

func ListThemes() []string {
	themeFS := ThemeFS()
	entries, err := fs.ReadDir(themeFS, ".")
	if err != nil {
		return nil
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), ".jsonc") {
			names = append(names, e.Name()[:len(e.Name())-6])
		} else if strings.HasSuffix(e.Name(), ".json") {
			names = append(names, e.Name()[:len(e.Name())-5])
		}
	}
	return names
}

func parseTheme(data []byte) (*Theme, error) {
	var theme Theme
	if err := utils.UnmarshalJSONCSonic(data, &theme); err != nil {
		return nil, fmt.Errorf("parse theme: %w", err)
	}
	if theme.Name == "" {
		theme.Name = "Custom"
	}
	return &theme, nil
}
