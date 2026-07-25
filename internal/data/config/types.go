package config

import (
	"time"

	"github.com/elizabevil/docker-tui/internal/utils"
)

const CurrentConfigVersion = 1

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

// RuntimeConn defines a named container runtime connection.
type RuntimeConn struct {
	Name       string           `json:"name" yaml:"name"`
	Driver     string           `json:"driver" yaml:"driver"`
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
	Connections []RuntimeConn          `json:"connections" yaml:"connections"`
}

// Config represents the application configuration.
type Config struct {
	ConfigVersion    int               `json:"configVersion" yaml:"configVersion"`
	General          GeneralConfig     `json:"general" yaml:"general"`
	UI               UIConfig          `json:"ui" yaml:"ui"`
	Docker           DockerConfig      `json:"docker" yaml:"docker"`
	Runtime          RuntimeConfig     `json:"runtime" yaml:"runtime"`
	Keymap           KeymapConfig      `json:"keymap" yaml:"keymap"`
	Logs             LogsConfig        `json:"logs" yaml:"logs"`
	Layout           LayoutConfig      `json:"layout" yaml:"layout"`
	CommandTemplates map[string]string `json:"commandTemplates" yaml:"commandTemplates"`
	Theme            string            `json:"theme" yaml:"theme"`
}

// GeneralConfig holds general application settings.
type GeneralConfig struct {
	ScrollHeight int    `json:"scrollHeight" yaml:"scrollHeight"`
	Reporting    string `json:"reporting" yaml:"reporting"`
	Lang         string `json:"lang" yaml:"lang"`
	SizeFormat   string `json:"sizeFormat" yaml:"sizeFormat"` // "binary" (1024-base, default) or "si" (1000-base)
}

// UIConfig holds UI-related settings.
type UIConfig struct {
	Theme              ThemeConfig `json:"theme" yaml:"theme"`
	WrapMainPanel      bool        `json:"wrapMainPanel" yaml:"wrapMainPanel"`
	ReturnImmediately  bool        `json:"returnImmediately" yaml:"returnImmediately"`
	BorderStyle        string      `json:"borderStyle" yaml:"borderStyle"`
	ShowHelp           bool        `json:"showHelp" yaml:"showHelp"`
	HintTimeout        int         `json:"hintTimeout" yaml:"hintTimeout"`               // seconds, default 3
	DialogOverlayColor string      `json:"dialogOverlayColor" yaml:"dialogOverlayColor"` // hex, e.g. "#0d1117"
	// EnableMouse toggles the bubbletea MouseMode; mouse clicks move
	// the cursor and scroll, keyboard input remains the primary path.
	EnableMouse bool `json:"enableMouse" yaml:"enableMouse"`
}

// ThemeConfig defines the color theme.
type ThemeConfig struct {
	ActiveBorderColor   []string `json:"activeBorderColor" yaml:"activeBorderColor"`
	InactiveBorderColor []string `json:"inactiveBorderColor" yaml:"inactiveBorderColor"`
	TableHeaderColor    []string `json:"tableHeaderColor" yaml:"tableHeaderColor"`
	SelectedRowColor    []string `json:"selectedRowColor" yaml:"selectedRowColor"`
	StatusBarColor      []string `json:"statusBarColor" yaml:"statusBarColor"`
	InfoColor           []string `json:"infoColor" yaml:"infoColor"`
	ErrorColor          []string `json:"errorColor" yaml:"errorColor"`
}

// DockerConfig holds Docker connection settings.
type DockerConfig struct {
	Timeout      time.Duration `json:"timeout" yaml:"timeout"`
	StatsPollSec int           `json:"statsPollSec" yaml:"statsPollSec"`
}

// KeymapConfig defines keyboard shortcuts.
type KeymapConfig struct {
	Quit    []string `json:"quit" yaml:"quit"`
	Help    []string `json:"help" yaml:"help"`
	Filter  []string `json:"filter" yaml:"filter"`
	Refresh []string `json:"refresh" yaml:"refresh"`

	ContainerStart   []string `json:"containerStart" yaml:"containerStart"`
	ContainerStop    []string `json:"containerStop" yaml:"containerStop"`
	ContainerRestart []string `json:"containerRestart" yaml:"containerRestart"`
	ContainerKill    []string `json:"containerKill" yaml:"containerKill"`
	ContainerRemove  []string `json:"containerRemove" yaml:"containerRemove"`
	ContainerLogs    []string `json:"containerLogs" yaml:"containerLogs"`
	ContainerExec    []string `json:"containerExec" yaml:"containerExec"`
	ContainerInspect []string `json:"containerInspect" yaml:"containerInspect"`
	ContainerStats   []string `json:"containerStats" yaml:"containerStats"`
	ContainerPause   []string `json:"containerPause" yaml:"containerPause"`
	// TASK-019: advanced container actions.
	ContainerUpdate []string `json:"containerUpdate" yaml:"containerUpdate"`
	ContainerDiff   []string `json:"containerDiff" yaml:"containerDiff"`
	ContainerExport []string `json:"containerExport" yaml:"containerExport"`
	ContainerCommit []string `json:"containerCommit" yaml:"containerCommit"`
	ContainerWait   []string `json:"containerWait" yaml:"containerWait"`
	ContainerCopy   []string `json:"containerCopy" yaml:"containerCopy"`

	ImagePull   []string `json:"imagePull" yaml:"imagePull"`
	ImageRemove []string `json:"imageRemove" yaml:"imageRemove"`
	ImagePrune  []string `json:"imagePrune" yaml:"imagePrune"`
	ImageTag    []string `json:"imageTag" yaml:"imageTag"`
	ImagePush   []string `json:"imagePush" yaml:"imagePush"`
	ImageSave   []string `json:"imageSave" yaml:"imageSave"`
	ImageLoad   []string `json:"imageLoad" yaml:"imageLoad"`

	VolumeCreate  []string `json:"volumeCreate" yaml:"volumeCreate"`
	VolumePrune   []string `json:"volumePrune" yaml:"volumePrune"`
	VolumeRemove  []string `json:"volumeRemove" yaml:"volumeRemove"`
	NetworkCreate []string `json:"networkCreate" yaml:"networkCreate"`
	NetworkPrune  []string `json:"networkPrune" yaml:"networkPrune"`
	NetworkRemove []string `json:"networkRemove" yaml:"networkRemove"`

	TabNext []string `json:"tabNext" yaml:"tabNext"`
	TabPrev []string `json:"tabPrev" yaml:"tabPrev"`

	Up     []string `json:"up" yaml:"up"`
	Down   []string `json:"down" yaml:"down"`
	Enter  []string `json:"enter" yaml:"enter"`
	Back   []string `json:"back" yaml:"back"`
	Delete []string `json:"delete" yaml:"delete"`
}

// BackgroundType defines how a background is rendered.
type BackgroundType string

const (
	BackgroundSolid    BackgroundType = "solid"
	BackgroundGradient BackgroundType = "gradient"
)

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
	Sizing string `json:"sizing" yaml:"sizing"`

	// ImagePosition: vertical alignment for sampling. "center", "top", "bottom".
	Position string `json:"position" yaml:"position"`

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

// DefaultConfig returns a configuration with sensible defaults,
// loaded from the embedded default.jsonc and merged with Go-level defaults.
func DefaultConfig() *Config {
	var cfg Config
	if err := utils.UnmarshalJSONCSonic([]byte(defaultConfigJSON), &cfg); err != nil {
		// Embedded JSONC should always parse; fallback if something goes wrong.
		return fallbackConfig()
	}
	// Ensure UI theme colors have defaults (too complex for JSONC).
	if len(cfg.UI.Theme.ActiveBorderColor) == 0 {
		cfg.UI.Theme = ThemeConfig{
			ActiveBorderColor:   []string{"green", "bold"},
			InactiveBorderColor: []string{"default"},
			TableHeaderColor:    []string{"cyan", "bold"},
			SelectedRowColor:    []string{"blue", "bold"},
			StatusBarColor:      []string{"green", "bold"},
			InfoColor:           []string{"cyan"},
			ErrorColor:          []string{"red", "bold"},
		}
	}
	return &cfg
}

// fallbackConfig returns a minimal hardcoded config if the embedded JSONC fails to parse.
func fallbackConfig() *Config {
	return &Config{
		ConfigVersion: CurrentConfigVersion,
		General:       GeneralConfig{ScrollHeight: 2, Reporting: "off", Lang: "en"},
		Docker:        DockerConfig{Timeout: 30 * time.Second, StatsPollSec: 3},
		Runtime: RuntimeConfig{
			Default:   "local-docker",
			Discovery: RuntimeDiscoveryConfig{LocalDocker: true, LocalPodman: true},
			Health:    RuntimeHealthConfig{IntervalSec: 3, TimeoutSec: 2, FailureThreshold: 2},
		},
		Layout: LayoutConfig{SectionWeights: SectionWeights{Top: 2, Content: 7, Bottom: 1}},
		Keymap: KeymapConfig{Quit: []string{"q"}, Up: []string{"up", "k"}, Down: []string{"down", "j"}},
	}
}
