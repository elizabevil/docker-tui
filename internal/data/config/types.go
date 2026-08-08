package config

import "time"

// RuntimeTLSConfig defines TLS client authentication for one runtime endpoint.
type RuntimeTLSConfig struct {
	Enabled            bool   `json:"enabled" yaml:"enabled"`
	Verify             bool   `json:"verify" yaml:"verify"`
	InsecureSkipVerify bool   `json:"insecureSkipVerify" yaml:"insecureSkipVerify"`
	CAFile             string `json:"caFile" yaml:"caFile"`
	CertFile           string `json:"certFile" yaml:"certFile"`
	KeyFile            string `json:"keyFile" yaml:"keyFile"`
	ServerName         string `json:"serverName" yaml:"serverName"`
}

// RuntimeConnection defines a named container runtime connection.
type RuntimeConnection struct {
	Name       string           `json:"name" yaml:"name"`
	Driver     RuntimeDriver    `json:"driver" yaml:"driver"`
	Endpoint   string           `json:"endpoint" yaml:"endpoint"`
	APIVersion string           `json:"apiVersion" yaml:"apiVersion"`
	TLS        RuntimeTLSConfig `json:"tls" yaml:"tls"`
}

type RuntimeDiscoveryConfig struct {
	LocalDocker bool `json:"localDocker" yaml:"localDocker"`
	LocalPodman bool `json:"localPodman" yaml:"localPodman"`
}

type RuntimeHealthConfig struct {
	IntervalSec      int `json:"intervalSec" yaml:"intervalSec"`
	TimeoutSec       int `json:"timeoutSec" yaml:"timeoutSec"`
	FailureThreshold int `json:"failureThreshold" yaml:"failureThreshold"`
}

// Interval is safe to use directly because loaded configurations are validated.
func (h RuntimeHealthConfig) Interval() time.Duration {
	return time.Duration(h.IntervalSec) * time.Second
}

// Timeout is safe to use directly because loaded configurations are validated.
func (h RuntimeHealthConfig) Timeout() time.Duration {
	return time.Duration(h.TimeoutSec) * time.Second
}

// RuntimeConfig configures container runtime connections.
type RuntimeConfig struct {
	Default     string                 `json:"default" yaml:"default"`
	Discovery   RuntimeDiscoveryConfig `json:"discovery" yaml:"discovery"`
	Health      RuntimeHealthConfig    `json:"health" yaml:"health"`
	Connections []RuntimeConnection    `json:"connections" yaml:"connections"`
}

// AppConfig is the complete, validated application behavior configuration.
// Visual properties are intentionally absent; they belong to Theme.
type AppConfig struct {
	Version  int           `json:"version" yaml:"version"`
	General  GeneralConfig `json:"general" yaml:"general"`
	UI       UIConfig      `json:"ui" yaml:"ui"`
	Docker   DockerConfig  `json:"docker" yaml:"docker"`
	Runtime  RuntimeConfig `json:"runtime" yaml:"runtime"`
	Keymap   KeymapConfig  `json:"keymap" yaml:"keymap"`
	Logs     LogsConfig    `json:"logs" yaml:"logs"`
	Layout   LayoutConfig  `json:"layout" yaml:"layout"`
	Commands CommandConfig `json:"commands" yaml:"commands"`
}

// GeneralConfig holds general application settings.
type GeneralConfig struct {
	ScrollHeight int           `json:"scrollHeight" yaml:"scrollHeight"`
	Reporting    ReportingMode `json:"reporting" yaml:"reporting"`
	Lang         Language      `json:"lang" yaml:"lang"`
	SizeFormat   SizeFormat    `json:"sizeFormat" yaml:"sizeFormat"`
}

// UIConfig holds UI-related settings.
type UIConfig struct {
	WrapMainPanel     bool               `json:"wrapMainPanel" yaml:"wrapMainPanel"`
	ReturnImmediately bool               `json:"returnImmediately" yaml:"returnImmediately"`
	ShowHelp          bool               `json:"showHelp" yaml:"showHelp"`
	HintTimeout       int                `json:"hintTimeout" yaml:"hintTimeout"`
	Dialog            DialogLayoutConfig `json:"dialog" yaml:"dialog"`
	Window            WindowLayoutConfig `json:"window" yaml:"window"`
	Header            HeaderLayoutConfig `json:"header" yaml:"header"`
	Table             TableLayoutConfig  `json:"table" yaml:"table"`
	// EnableMouse toggles the bubbletea MouseMode; mouse clicks move
	// the cursor and scroll, keyboard input remains the primary path.
	EnableMouse bool `json:"enableMouse" yaml:"enableMouse"`
}

type WindowLayoutConfig struct {
	MarginTopPercent    int `json:"marginTopPercent" yaml:"marginTopPercent"`
	MarginBottomPercent int `json:"marginBottomPercent" yaml:"marginBottomPercent"`
	ContentWidthPercent int `json:"contentWidthPercent" yaml:"contentWidthPercent"`
}

type HeaderColumnWeights struct {
	Host       int `json:"host" yaml:"host"`
	Connection int `json:"connection" yaml:"connection"`
	Keystroke  int `json:"keystroke" yaml:"keystroke"`
	Logo       int `json:"logo" yaml:"logo"`
}

type HeaderLayoutConfig struct {
	Columns               HeaderColumnWeights `json:"columns" yaml:"columns"`
	KeystrokeContentRatio int                 `json:"keystrokeContentRatio" yaml:"keystrokeContentRatio"`
}

type TableLayoutConfig struct {
	ColumnSpacing     int                 `json:"columnSpacing" yaml:"columnSpacing"`
	RowPrefix         string              `json:"rowPrefix" yaml:"rowPrefix"`
	RowPrefixSelected string              `json:"rowPrefixSelected" yaml:"rowPrefixSelected"`
	RowSpacing        int                 `json:"rowSpacing" yaml:"rowSpacing"`
	SelectionInfo     SelectionInfoConfig `json:"selectionInfo" yaml:"selectionInfo"`
}

type SelectionInfoConfig struct {
	Enabled  bool `json:"enabled" yaml:"enabled"`
	PadLines int  `json:"padLines" yaml:"padLines"`
}

type ResponsiveSize struct {
	Percent int `json:"percent" yaml:"percent"`
	Min     int `json:"min" yaml:"min"`
	Max     int `json:"max" yaml:"max"`
}

type DialogPosition struct {
	Horizontal HorizontalAlignment `json:"horizontal" yaml:"horizontal"`
	Vertical   VerticalAlignment   `json:"vertical" yaml:"vertical"`
	OffsetX    int                 `json:"offsetX" yaml:"offsetX"`
	OffsetY    int                 `json:"offsetY" yaml:"offsetY"`
}

type DialogLayoutConfig struct {
	Width    ResponsiveSize `json:"width" yaml:"width"`
	Height   ResponsiveSize `json:"height" yaml:"height"`
	Position DialogPosition `json:"position" yaml:"position"`
	// PanelSize governs dialog sizing when a panel body is available
	// (BR-043 §3.2). WidthPercent is multiplied by panel body width;
	// HeightPercent by panel body height. The width is upper-bounded by
	// MaxWidth so wide content never exceeds a sensible dialog footprint
	// (defaults: 75 % × full height, capped at 120 cells / 40 rows).
	PanelSize PanelSize `json:"panelSize" yaml:"panelSize"`
}

// PanelSize sizes a dialog against the active panel body (BR-043 §3.2).
// WidthPercent is multiplied by panel body width; HeightPercent by
// panel body height. The width is upper-bounded by MaxWidth; the height
// is upper-bounded by MaxHeight. A zero value means "fall back to
// cfg.Width / cfg.Height" in the panel sizing helpers.
type PanelSize struct {
	WidthPercent  int `json:"widthPercent" yaml:"widthPercent"`
	HeightPercent int `json:"heightPercent" yaml:"heightPercent"`
	MaxWidth      int `json:"maxWidth" yaml:"maxWidth"`
	MaxHeight     int `json:"maxHeight" yaml:"maxHeight"`
}

// DockerConfig holds Docker connection settings.
type DockerConfig struct {
	Timeout      time.Duration `json:"timeout" yaml:"timeout"`
	StatsPollSec int           `json:"statsPollSec" yaml:"statsPollSec"`
}

type Key string

type KeyBinding struct {
	Primary   Key `json:"primary" yaml:"primary"`
	Secondary Key `json:"secondary,omitempty" yaml:"secondary,omitempty"`
}

func (b KeyBinding) Values() []string {
	values := make([]string, 0, 2)
	if b.Primary != KeyEmpty {
		values = append(values, string(b.Primary))
	}
	if b.Secondary != KeyEmpty {
		values = append(values, string(b.Secondary))
	}
	return values
}

type GlobalKeymap struct {
	Quit      KeyBinding `json:"quit" yaml:"quit"`
	ActionBar KeyBinding `json:"actionBar" yaml:"actionBar"`
	Help      KeyBinding `json:"help" yaml:"help"`
	Filter    KeyBinding `json:"filter" yaml:"filter"`
	Refresh   KeyBinding `json:"refresh" yaml:"refresh"`
}

type ContainerKeymap struct {
	Start   KeyBinding `json:"start" yaml:"start"`
	Stop    KeyBinding `json:"stop" yaml:"stop"`
	Restart KeyBinding `json:"restart" yaml:"restart"`
	Kill    KeyBinding `json:"kill" yaml:"kill"`
	Remove  KeyBinding `json:"remove" yaml:"remove"`
	Logs    KeyBinding `json:"logs" yaml:"logs"`
	Exec    KeyBinding `json:"exec" yaml:"exec"`
	Inspect KeyBinding `json:"inspect" yaml:"inspect"`
	Stats   KeyBinding `json:"stats" yaml:"stats"`
	Pause   KeyBinding `json:"pause" yaml:"pause"`
	Update  KeyBinding `json:"update" yaml:"update"`
	Diff    KeyBinding `json:"diff" yaml:"diff"`
	Export  KeyBinding `json:"export" yaml:"export"`
	Commit  KeyBinding `json:"commit" yaml:"commit"`
	Wait    KeyBinding `json:"wait" yaml:"wait"`
	Copy    KeyBinding `json:"copy" yaml:"copy"`
}

type ImageKeymap struct {
	Pull    KeyBinding `json:"pull" yaml:"pull"`
	Remove  KeyBinding `json:"remove" yaml:"remove"`
	Prune   KeyBinding `json:"prune" yaml:"prune"`
	Tag     KeyBinding `json:"tag" yaml:"tag"`
	Push    KeyBinding `json:"push" yaml:"push"`
	Save    KeyBinding `json:"save" yaml:"save"`
	Load    KeyBinding `json:"load" yaml:"load"`
	History KeyBinding `json:"history" yaml:"history"`
}

type ResourceKeymap struct {
	Create KeyBinding `json:"create" yaml:"create"`
	Prune  KeyBinding `json:"prune" yaml:"prune"`
	Remove KeyBinding `json:"remove" yaml:"remove"`
}

type NavigationKeymap struct {
	TabNext KeyBinding `json:"tabNext" yaml:"tabNext"`
	TabPrev KeyBinding `json:"tabPrev" yaml:"tabPrev"`
	Up      KeyBinding `json:"up" yaml:"up"`
	Down    KeyBinding `json:"down" yaml:"down"`
	Enter   KeyBinding `json:"enter" yaml:"enter"`
	Back    KeyBinding `json:"back" yaml:"back"`
	Delete  KeyBinding `json:"delete" yaml:"delete"`
}

type DialogKeymap struct {
	Confirm KeyBinding `json:"confirm" yaml:"confirm"`
	Cancel  KeyBinding `json:"cancel" yaml:"cancel"`
}

type ComposeKeymap struct {
	Start       KeyBinding `json:"start" yaml:"start"`
	Stop        KeyBinding `json:"stop" yaml:"stop"`
	Restart     KeyBinding `json:"restart" yaml:"restart"`
	Down        KeyBinding `json:"down" yaml:"down"`
	Logs        KeyBinding `json:"logs" yaml:"logs"`
	Top         KeyBinding `json:"top" yaml:"top"`
	Port        KeyBinding `json:"port" yaml:"port"`
	Stats       KeyBinding `json:"stats" yaml:"stats"`
	Build       KeyBinding `json:"build" yaml:"build"`
	Pull        KeyBinding `json:"pull" yaml:"pull"`
	Push        KeyBinding `json:"push" yaml:"push"`
	Scale       KeyBinding `json:"scale" yaml:"scale"`
	Pause       KeyBinding `json:"pause" yaml:"pause"`
	Unpause     KeyBinding `json:"unpause" yaml:"unpause"`
	Kill        KeyBinding `json:"kill" yaml:"kill"`
	Rm          KeyBinding `json:"rm" yaml:"rm"`
	Prune       KeyBinding `json:"prune" yaml:"prune"`
	Events      KeyBinding `json:"events" yaml:"events"`
	Run         KeyBinding `json:"run" yaml:"run"`
	Exec        KeyBinding `json:"exec" yaml:"exec"`
	ServiceLogs KeyBinding `json:"serviceLogs" yaml:"service_logs"`
	Detail      KeyBinding `json:"detail" yaml:"detail"`
}

type KeymapConfig struct {
	Global     GlobalKeymap     `json:"global" yaml:"global"`
	Container  ContainerKeymap  `json:"container" yaml:"container"`
	Image      ImageKeymap      `json:"image" yaml:"image"`
	Volume     ResourceKeymap   `json:"volume" yaml:"volume"`
	Network    ResourceKeymap   `json:"network" yaml:"network"`
	Navigation NavigationKeymap `json:"navigation" yaml:"navigation"`
	Dialog     DialogKeymap     `json:"dialog" yaml:"dialog"`
	Compose    ComposeKeymap    `json:"compose" yaml:"compose"`
}

type DockerComposeCommand struct {
	Executable string `json:"executable" yaml:"executable"`
	Subcommand string `json:"subcommand" yaml:"subcommand"`
}

type CommandConfig struct {
	DockerCompose DockerComposeCommand `json:"dockerCompose" yaml:"dockerCompose"`
}

// BackgroundImage defines image background settings (CSS-inspired).
type BackgroundImage struct {
	// Image path/URL. CSS background-image: url(...).
	Src string `json:"src" yaml:"src"`

	// SampleRate controls horizontal sampling quality as % of image width.
	// 1 (default) = single pixel column at center (fastest).
	// 50 = average 50% of image width horizontally per row.
	// 100 = average the full image width per row (richest colors).
	SampleRate int `json:"sampleRate" yaml:"sampleRate"`

	// ImageSizing: "cover" (default, fill & crop) or "contain" (fit & letterbox).
	Sizing BackgroundSizing `json:"sizing" yaml:"sizing"`

	// ImagePosition: vertical alignment for sampling. "center", "top", "bottom".
	Position BackgroundPosition `json:"position" yaml:"position"`

	// Opacity: 0-100. 100 = fully opaque (default). Lower = more transparent,
	// letting terminal background show through for better text readability.
	Opacity int `json:"opacity" yaml:"opacity"`
}

// SectionBackground defines background rendering for one section (or global).
// CSS-inspired: solid color, gradient, or image with optional overlay.
type SectionBackground struct {
	Type  BackgroundType `json:"type" yaml:"type"`   // "solid" (default), "gradient", "image"
	Color string         `json:"color" yaml:"color"` // hex color

	// Gradient
	StartColor string `json:"startColor" yaml:"startColor"`
	EndColor   string `json:"endColor" yaml:"endColor"`

	// Image settings (used when Type = "image")
	Image BackgroundImage `json:"image" yaml:"image"`

	// Overlay: semi-transparent color on top of background for text readability.
	// CSS-like: overlay creates a colored layer between bg and text.
	OverlayColor string `json:"overlayColor" yaml:"overlayColor"`
	// OverlayOpacity: 0-100. Higher = darker overlay, clearer text.
	OverlayOpacity int `json:"overlayOpacity" yaml:"overlayOpacity"`
}

// SectionBackgroundOverrides defines optional per-section background overrides.
// Each section inherits from the global BackgroundConfig if not overridden.
type SectionBackgroundOverrides struct {
	Top    *SectionBackground `json:"top,omitempty" yaml:"top,omitempty"`
	Middle *SectionBackground `json:"middle,omitempty" yaml:"middle,omitempty"`
	Bottom *SectionBackground `json:"bottom,omitempty" yaml:"bottom,omitempty"`
}

// BackgroundOverlay defines a semi-transparent overlay layer (Layer 2).
type BackgroundOverlay struct {
	Color   string `json:"color" yaml:"color"`     // hex overlay color
	Opacity int    `json:"opacity" yaml:"opacity"` // 0-100
}

// BackgroundConfig is the top-level background configuration.
// Three-layer architecture (bottom to top):
//
//	Layer 1: Image/Color/Gradient → pixelated full-terminal background
//	Layer 2: Overlay → semi-transparent tint for readability
//	Layer 3: Content → component styles rendered on top
type BackgroundConfig struct {
	// Enable toggles the custom background on/off.
	Enable bool `json:"enable" yaml:"enable"`

	// Fallthrough controls whether component background/foreground colors are
	// overridden by the image. When true, component styles are hidden and only
	// the image+overlay layers show through.
	Fallthrough bool `json:"fallthrough" yaml:"fallthrough"`

	// Layer 1: Background type and colors
	Type  BackgroundType `json:"type" yaml:"type"`
	Color string         `json:"color" yaml:"color"`

	// Gradient
	StartColor string `json:"startColor" yaml:"startColor"`
	EndColor   string `json:"endColor" yaml:"endColor"`

	// Image
	Image BackgroundImage `json:"image" yaml:"image"`

	// Layer 2: Semi-transparent overlay on top of Layer 1
	Overlay BackgroundOverlay `json:"overlay" yaml:"overlay"`

	// Per-section overrides (optional — inherit global when nil)
	Sections SectionBackgroundOverrides `json:"sections" yaml:"sections"`
}

// SectionWeights defines weight-based proportional sizing for the three layout sections.
// Weights are relative: total = top + content + bottom, each section gets weight/total.
type SectionWeights struct {
	Top     int `json:"top" yaml:"top"`
	Content int `json:"content" yaml:"content"`
	Bottom  int `json:"bottom" yaml:"bottom"`
}

// LayoutConfig controls the three-section layout (top / content / bottom).
// Section sizing uses weights (sibling-level); margins use percentages (scaling/nested).
type LayoutConfig struct {
	SectionWeights SectionWeights   `json:"sectionWeights" yaml:"sectionWeights"`
	Background     BackgroundConfig `json:"background" yaml:"background"`
}

// LogsConfig holds log streaming settings.
type LogsConfig struct {
	Since      string `json:"since" yaml:"since"`
	Tail       string `json:"tail" yaml:"tail"`
	Timestamps bool   `json:"timestamps" yaml:"timestamps"`
}

// DefaultAppConfig returns the compiled fallback. Embedded defaults are
// applied by LoadResolved and intentionally remain a higher-priority layer.
func DefaultAppConfig() *AppConfig {
	return &AppConfig{
		Version: CurrentConfigVersion,
		General: GeneralConfig{
			ScrollHeight: 2, Reporting: ReportingOff, Lang: LanguageEnglish, SizeFormat: SizeFormatBinary,
		},
		UI: UIConfig{
			ShowHelp: true, HintTimeout: 3, EnableMouse: true,
			Dialog: DialogLayoutConfig{
				// Dialogs (Copy / Update / Export / Commit / Action Bar /
				// Image Save / Image Load) use 75% of the active panel body
				// so the form fields have room to render, clamped to a
				// reasonable Min/Max so the dialog never collapses to a
				// sliver on a wide terminal or explodes past the viewport.
				Width:    ResponsiveSize{Percent: 75, Min: 60, Max: 120},
				Height:   ResponsiveSize{Percent: 75, Min: 16, Max: 40},
				Position: DialogPosition{Horizontal: HorizontalCenter, Vertical: VerticalCenter},
				// PanelSize: panel-body driven sizing (BR-043 §3.2).
				// Width 75% of body, full height (100%), capped at 120 cells / 40 rows.
				PanelSize: PanelSize{
					WidthPercent: 75, HeightPercent: 100,
					MaxWidth: 120, MaxHeight: 40,
				},
			},
			Window: WindowLayoutConfig{MarginTopPercent: 2, MarginBottomPercent: 2, ContentWidthPercent: 99},
			Header: HeaderLayoutConfig{
				// Header columns split the available width using a 2:2:3:3
				// ratio (Host / Connection / Keystroke / Logo). Keystroke and
				// Logo together take half the header so the per-key badge row
				// stays readable at typical 80-120 column TTI widths.
				Columns:               HeaderColumnWeights{Host: 2, Connection: 2, Keystroke: 3, Logo: 3},
				KeystrokeContentRatio: 80,
			},
			Table: TableLayoutConfig{
				ColumnSpacing: 2, RowPrefix: DefaultTableRowPrefix, RowPrefixSelected: DefaultTableSelectedPrefix,
				SelectionInfo: SelectionInfoConfig{Enabled: true, PadLines: 1},
			},
		},
		Docker: DockerConfig{Timeout: DefaultDockerTimeout, StatsPollSec: 3},
		Runtime: RuntimeConfig{
			Default:   DefaultConnectionLocalDocker,
			Discovery: RuntimeDiscoveryConfig{LocalDocker: true, LocalPodman: true},
			Health:    RuntimeHealthConfig{IntervalSec: 3, TimeoutSec: 2, FailureThreshold: 2},
		},
		Logs: LogsConfig{Since: DefaultLogsSince, Tail: DefaultLogsTail},
		Layout: LayoutConfig{
			SectionWeights: SectionWeights{Top: 2, Content: 7, Bottom: 1},
			Background:     LayoutBackgroundFallback(),
		},
		Commands: CommandConfig{DockerCompose: DockerComposeCommand{Executable: DefaultDockerExecutable, Subcommand: DefaultDockerSubcommand}},
		Keymap:   defaultKeymap(),
	}
}

func LayoutBackgroundFallback() BackgroundConfig {
	return BackgroundConfig{
		Type: BackgroundSolid,
		Image: BackgroundImage{
			Sizing:   BackgroundSizingCover,
			Position: BackgroundPositionCenter,
			Opacity:  100,
		},
	}
}

func defaultKeymap() KeymapConfig {
	b := func(primary Key, secondary ...Key) KeyBinding {
		binding := KeyBinding{Primary: primary}
		if len(secondary) > 0 {
			binding.Secondary = secondary[0]
		}
		return binding
	}
	return KeymapConfig{
		Global:     GlobalKeymap{Quit: b(KeyCtrlC), ActionBar: b(KeySemicolon), Help: b(KeyQuestion, KeyF1), Filter: b(KeySlash), Refresh: b(KeyR)},
		Container:  ContainerKeymap{Start: b(KeyS), Stop: b(KeyCtrlS), Restart: b(KeyCtrlR), Kill: b(KeyCtrlK), Remove: b(KeyCtrlD), Logs: b(KeyL), Exec: b(KeyE), Inspect: b(KeyI), Stats: b(KeyM), Pause: b(KeyP)},
		Image:      ImageKeymap{Pull: b(KeyCtrlP), Remove: b(KeyCtrlD), Prune: b(KeyP), Tag: b(KeyCtrlT), Push: b(KeyCtrlU), Save: b(KeyCtrlE), Load: b(KeyCtrlL), History: b(KeyH)},
		Volume:     ResourceKeymap{Create: b(KeyC), Prune: b(KeyP), Remove: b(KeyCtrlD)},
		Network:    ResourceKeymap{Create: b(KeyC), Prune: b(KeyP), Remove: b(KeyCtrlD)},
		Navigation: NavigationKeymap{TabNext: b(KeyTab, KeyCloseBracket), TabPrev: b(KeyShiftTab, KeyOpenBracket), Up: b(KeyUp, KeyK), Down: b(KeyDown, KeyJ), Enter: b(KeyEnter), Back: b(KeyEscape), Delete: b(KeyCtrlD)},
		Dialog:     DialogKeymap{Confirm: b(KeyEnter), Cancel: b(KeyEscape)},
		Compose: ComposeKeymap{
			Start:   b(KeyS),
			Stop:    b(KeyCtrlS),
			Restart: b(KeyCtrlR),
			Down:    b(KeyCtrlD),
			Logs:    b(KeyL),
			Top:     b(KeyCtrlT),
			Port:    b(KeyComma),
			Kill:    b(KeyCtrlK),
			Events:  b(KeyCtrlF3),
			Exec:    b(KeyE),
			Detail:  b(KeyD),
		},
	}
}
