package config

import "time"

// AppPatch represents one typed application-configuration layer. Pointer
// fields preserve the distinction between an omitted property and an explicit
// zero value.
type AppPatch struct {
	General  *GeneralPatch `json:"general,omitempty" yaml:"general,omitempty"`
	UI       *UIPatch      `json:"ui,omitempty" yaml:"ui,omitempty"`
	Docker   *DockerPatch  `json:"docker,omitempty" yaml:"docker,omitempty"`
	Runtime  *RuntimePatch `json:"runtime,omitempty" yaml:"runtime,omitempty"`
	Keymap   *KeymapPatch  `json:"keymap,omitempty" yaml:"keymap,omitempty"`
	Logs     *LogsPatch    `json:"logs,omitempty" yaml:"logs,omitempty"`
	Layout   *LayoutPatch  `json:"layout,omitempty" yaml:"layout,omitempty"`
	Commands *CommandPatch `json:"commands,omitempty" yaml:"commands,omitempty"`
}

type GeneralPatch struct {
	ScrollHeight *int           `json:"scrollHeight,omitempty" yaml:"scrollHeight,omitempty"`
	Reporting    *ReportingMode `json:"reporting,omitempty" yaml:"reporting,omitempty"`
	Lang         *Language      `json:"lang,omitempty" yaml:"lang,omitempty"`
	SizeFormat   *SizeFormat    `json:"sizeFormat,omitempty" yaml:"sizeFormat,omitempty"`
}

type UIPatch struct {
	WrapMainPanel     *bool              `json:"wrapMainPanel,omitempty" yaml:"wrapMainPanel,omitempty"`
	ReturnImmediately *bool              `json:"returnImmediately,omitempty" yaml:"returnImmediately,omitempty"`
	ShowHelp          *bool              `json:"showHelp,omitempty" yaml:"showHelp,omitempty"`
	HintTimeout       *int               `json:"hintTimeout,omitempty" yaml:"hintTimeout,omitempty"`
	EnableMouse       *bool              `json:"enableMouse,omitempty" yaml:"enableMouse,omitempty"`
	Dialog            *DialogLayoutPatch `json:"dialog,omitempty" yaml:"dialog,omitempty"`
	Window            *WindowLayoutPatch `json:"window,omitempty" yaml:"window,omitempty"`
	Header            *HeaderLayoutPatch `json:"header,omitempty" yaml:"header,omitempty"`
	Table             *TableLayoutPatch  `json:"table,omitempty" yaml:"table,omitempty"`
}

type WindowLayoutPatch struct {
	MarginTopPercent    *int `json:"marginTopPercent,omitempty" yaml:"marginTopPercent,omitempty"`
	MarginBottomPercent *int `json:"marginBottomPercent,omitempty" yaml:"marginBottomPercent,omitempty"`
	ContentWidthPercent *int `json:"contentWidthPercent,omitempty" yaml:"contentWidthPercent,omitempty"`
}

type HeaderLayoutPatch struct {
	Columns               *HeaderColumnWeightsPatch `json:"columns,omitempty" yaml:"columns,omitempty"`
	KeystrokeContentRatio *int                      `json:"keystrokeContentRatio,omitempty" yaml:"keystrokeContentRatio,omitempty"`
}

type HeaderColumnWeightsPatch struct {
	Host       *int `json:"host,omitempty" yaml:"host,omitempty"`
	Connection *int `json:"connection,omitempty" yaml:"connection,omitempty"`
	Keystroke  *int `json:"keystroke,omitempty" yaml:"keystroke,omitempty"`
	Logo       *int `json:"logo,omitempty" yaml:"logo,omitempty"`
}

type TableLayoutPatch struct {
	ColumnSpacing     *int                `json:"columnSpacing,omitempty" yaml:"columnSpacing,omitempty"`
	RowPrefix         *string             `json:"rowPrefix,omitempty" yaml:"rowPrefix,omitempty"`
	RowPrefixSelected *string             `json:"rowPrefixSelected,omitempty" yaml:"rowPrefixSelected,omitempty"`
	RowSpacing        *int                `json:"rowSpacing,omitempty" yaml:"rowSpacing,omitempty"`
	SelectionInfo     *SelectionInfoPatch `json:"selectionInfo,omitempty" yaml:"selectionInfo,omitempty"`
}

type SelectionInfoPatch struct {
	Enabled  *bool `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	PadLines *int  `json:"padLines,omitempty" yaml:"padLines,omitempty"`
}

type DialogLayoutPatch struct {
	Width     *ResponsiveSizePatch `json:"width,omitempty" yaml:"width,omitempty"`
	Height    *ResponsiveSizePatch `json:"height,omitempty" yaml:"height,omitempty"`
	Position  *DialogPositionPatch `json:"position,omitempty" yaml:"position,omitempty"`
	PanelSize *PanelSizePatch      `json:"panelSize,omitempty" yaml:"panelSize,omitempty"`
}

type ResponsiveSizePatch struct {
	Percent *int `json:"percent,omitempty" yaml:"percent,omitempty"`
	Min     *int `json:"min,omitempty" yaml:"min,omitempty"`
	Max     *int `json:"max,omitempty" yaml:"max,omitempty"`
}

type PanelSizePatch struct {
	WidthPercent  *int `json:"widthPercent,omitempty" yaml:"widthPercent,omitempty"`
	HeightPercent *int `json:"heightPercent,omitempty" yaml:"heightPercent,omitempty"`
	MaxWidth      *int `json:"maxWidth,omitempty" yaml:"maxWidth,omitempty"`
	MaxHeight     *int `json:"maxHeight,omitempty" yaml:"maxHeight,omitempty"`
}

type DialogPositionPatch struct {
	Horizontal *HorizontalAlignment `json:"horizontal,omitempty" yaml:"horizontal,omitempty"`
	Vertical   *VerticalAlignment   `json:"vertical,omitempty" yaml:"vertical,omitempty"`
	OffsetX    *int                 `json:"offsetX,omitempty" yaml:"offsetX,omitempty"`
	OffsetY    *int                 `json:"offsetY,omitempty" yaml:"offsetY,omitempty"`
}

type DockerPatch struct {
	Timeout      *time.Duration `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	StatsPollSec *int           `json:"statsPollSec,omitempty" yaml:"statsPollSec,omitempty"`
}

type RuntimePatch struct {
	Default     *string                `json:"default,omitempty" yaml:"default,omitempty"`
	Discovery   *RuntimeDiscoveryPatch `json:"discovery,omitempty" yaml:"discovery,omitempty"`
	Health      *RuntimeHealthPatch    `json:"health,omitempty" yaml:"health,omitempty"`
	Connections *[]RuntimeConnection   `json:"connections,omitempty" yaml:"connections,omitempty"`
}

type RuntimeDiscoveryPatch struct {
	LocalDocker *bool `json:"localDocker,omitempty" yaml:"localDocker,omitempty"`
	LocalPodman *bool `json:"localPodman,omitempty" yaml:"localPodman,omitempty"`
}

type RuntimeHealthPatch struct {
	IntervalSec      *int `json:"intervalSec,omitempty" yaml:"intervalSec,omitempty"`
	TimeoutSec       *int `json:"timeoutSec,omitempty" yaml:"timeoutSec,omitempty"`
	FailureThreshold *int `json:"failureThreshold,omitempty" yaml:"failureThreshold,omitempty"`
}

type LogsPatch struct {
	Since      *string `json:"since,omitempty" yaml:"since,omitempty"`
	Tail       *string `json:"tail,omitempty" yaml:"tail,omitempty"`
	Timestamps *bool   `json:"timestamps,omitempty" yaml:"timestamps,omitempty"`
}

type CommandPatch struct {
	DockerCompose *DockerComposeCommandPatch `json:"dockerCompose,omitempty" yaml:"dockerCompose,omitempty"`
}

type DockerComposeCommandPatch struct {
	Executable *string `json:"executable,omitempty" yaml:"executable,omitempty"`
	Subcommand *string `json:"subcommand,omitempty" yaml:"subcommand,omitempty"`
}

type KeymapPatch struct {
	Global     *GlobalKeymapPatch     `json:"global,omitempty" yaml:"global,omitempty"`
	Container  *ContainerKeymapPatch  `json:"container,omitempty" yaml:"container,omitempty"`
	Image      *ImageKeymapPatch      `json:"image,omitempty" yaml:"image,omitempty"`
	Volume     *ResourceKeymapPatch   `json:"volume,omitempty" yaml:"volume,omitempty"`
	Network    *ResourceKeymapPatch   `json:"network,omitempty" yaml:"network,omitempty"`
	Navigation *NavigationKeymapPatch `json:"navigation,omitempty" yaml:"navigation,omitempty"`
	Dialog     *DialogKeymapPatch     `json:"dialog,omitempty" yaml:"dialog,omitempty"`
	Compose    *ComposeKeymapPatch    `json:"compose,omitempty" yaml:"compose,omitempty"`
}

type GlobalKeymapPatch struct {
	Quit      *KeyBinding `json:"quit,omitempty" yaml:"quit,omitempty"`
	ActionBar *KeyBinding `json:"actionBar,omitempty" yaml:"actionBar,omitempty"`
	Help      *KeyBinding `json:"help,omitempty" yaml:"help,omitempty"`
	Filter    *KeyBinding `json:"filter,omitempty" yaml:"filter,omitempty"`
	Refresh   *KeyBinding `json:"refresh,omitempty" yaml:"refresh,omitempty"`
}

type ContainerKeymapPatch struct {
	Start   *KeyBinding `json:"start,omitempty" yaml:"start,omitempty"`
	Stop    *KeyBinding `json:"stop,omitempty" yaml:"stop,omitempty"`
	Restart *KeyBinding `json:"restart,omitempty" yaml:"restart,omitempty"`
	Kill    *KeyBinding `json:"kill,omitempty" yaml:"kill,omitempty"`
	Remove  *KeyBinding `json:"remove,omitempty" yaml:"remove,omitempty"`
	Logs    *KeyBinding `json:"logs,omitempty" yaml:"logs,omitempty"`
	Exec    *KeyBinding `json:"exec,omitempty" yaml:"exec,omitempty"`
	Inspect *KeyBinding `json:"inspect,omitempty" yaml:"inspect,omitempty"`
	Stats   *KeyBinding `json:"stats,omitempty" yaml:"stats,omitempty"`
	Pause   *KeyBinding `json:"pause,omitempty" yaml:"pause,omitempty"`
	Update  *KeyBinding `json:"update,omitempty" yaml:"update,omitempty"`
	Diff    *KeyBinding `json:"diff,omitempty" yaml:"diff,omitempty"`
	Export  *KeyBinding `json:"export,omitempty" yaml:"export,omitempty"`
	Commit  *KeyBinding `json:"commit,omitempty" yaml:"commit,omitempty"`
	Wait    *KeyBinding `json:"wait,omitempty" yaml:"wait,omitempty"`
	Copy    *KeyBinding `json:"copy,omitempty" yaml:"copy,omitempty"`
}

type ImageKeymapPatch struct {
	Pull    *KeyBinding `json:"pull,omitempty" yaml:"pull,omitempty"`
	Remove  *KeyBinding `json:"remove,omitempty" yaml:"remove,omitempty"`
	Prune   *KeyBinding `json:"prune,omitempty" yaml:"prune,omitempty"`
	Tag     *KeyBinding `json:"tag,omitempty" yaml:"tag,omitempty"`
	Push    *KeyBinding `json:"push,omitempty" yaml:"push,omitempty"`
	Save    *KeyBinding `json:"save,omitempty" yaml:"save,omitempty"`
	Load    *KeyBinding `json:"load,omitempty" yaml:"load,omitempty"`
	History *KeyBinding `json:"history,omitempty" yaml:"history,omitempty"`
}

type ResourceKeymapPatch struct {
	Create *KeyBinding `json:"create,omitempty" yaml:"create,omitempty"`
	Prune  *KeyBinding `json:"prune,omitempty" yaml:"prune,omitempty"`
	Remove *KeyBinding `json:"remove,omitempty" yaml:"remove,omitempty"`
}

type NavigationKeymapPatch struct {
	TabNext *KeyBinding `json:"tabNext,omitempty" yaml:"tabNext,omitempty"`
	TabPrev *KeyBinding `json:"tabPrev,omitempty" yaml:"tabPrev,omitempty"`
	Up      *KeyBinding `json:"up,omitempty" yaml:"up,omitempty"`
	Down    *KeyBinding `json:"down,omitempty" yaml:"down,omitempty"`
	Enter   *KeyBinding `json:"enter,omitempty" yaml:"enter,omitempty"`
	Back    *KeyBinding `json:"back,omitempty" yaml:"back,omitempty"`
	Delete  *KeyBinding `json:"delete,omitempty" yaml:"delete,omitempty"`
}

type DialogKeymapPatch struct {
	Confirm *KeyBinding `json:"confirm,omitempty" yaml:"confirm,omitempty"`
	Cancel  *KeyBinding `json:"cancel,omitempty" yaml:"cancel,omitempty"`
}

type ComposeKeymapPatch struct {
	Start       *KeyBinding `json:"start,omitempty" yaml:"start,omitempty"`
	Stop        *KeyBinding `json:"stop,omitempty" yaml:"stop,omitempty"`
	Restart     *KeyBinding `json:"restart,omitempty" yaml:"restart,omitempty"`
	Down        *KeyBinding `json:"down,omitempty" yaml:"down,omitempty"`
	Logs        *KeyBinding `json:"logs,omitempty" yaml:"logs,omitempty"`
	Top         *KeyBinding `json:"top,omitempty" yaml:"top,omitempty"`
	Port        *KeyBinding `json:"port,omitempty" yaml:"port,omitempty"`
	Stats       *KeyBinding `json:"stats,omitempty" yaml:"stats,omitempty"`
	Build       *KeyBinding `json:"build,omitempty" yaml:"build,omitempty"`
	Pull        *KeyBinding `json:"pull,omitempty" yaml:"pull,omitempty"`
	Push        *KeyBinding `json:"push,omitempty" yaml:"push,omitempty"`
	Scale       *KeyBinding `json:"scale,omitempty" yaml:"scale,omitempty"`
	Pause       *KeyBinding `json:"pause,omitempty" yaml:"pause,omitempty"`
	Unpause     *KeyBinding `json:"unpause,omitempty" yaml:"unpause,omitempty"`
	Kill        *KeyBinding `json:"kill,omitempty" yaml:"kill,omitempty"`
	Rm          *KeyBinding `json:"rm,omitempty" yaml:"rm,omitempty"`
	Prune       *KeyBinding `json:"prune,omitempty" yaml:"prune,omitempty"`
	Events      *KeyBinding `json:"events,omitempty" yaml:"events,omitempty"`
	Run         *KeyBinding `json:"run,omitempty" yaml:"run,omitempty"`
	Exec        *KeyBinding `json:"exec,omitempty" yaml:"exec,omitempty"`
	ServiceLogs *KeyBinding `json:"serviceLogs,omitempty" yaml:"service_logs,omitempty"`
	Detail      *KeyBinding `json:"detail,omitempty" yaml:"detail,omitempty"`
}

type LayoutPatch struct {
	SectionWeights *SectionWeightsPatch `json:"sectionWeights,omitempty" yaml:"sectionWeights,omitempty"`
	Background     *BackgroundPatch     `json:"background,omitempty" yaml:"background,omitempty"`
}

type SectionWeightsPatch struct {
	Top     *int `json:"top,omitempty" yaml:"top,omitempty"`
	Content *int `json:"content,omitempty" yaml:"content,omitempty"`
	Bottom  *int `json:"bottom,omitempty" yaml:"bottom,omitempty"`
}

type BackgroundPatch struct {
	Enable      *bool                    `json:"enable,omitempty" yaml:"enable,omitempty"`
	Fallthrough *bool                    `json:"fallthrough,omitempty" yaml:"fallthrough,omitempty"`
	Type        *BackgroundType          `json:"type,omitempty" yaml:"type,omitempty"`
	Color       *string                  `json:"color,omitempty" yaml:"color,omitempty"`
	StartColor  *string                  `json:"startColor,omitempty" yaml:"startColor,omitempty"`
	EndColor    *string                  `json:"endColor,omitempty" yaml:"endColor,omitempty"`
	Image       *BackgroundImagePatch    `json:"image,omitempty" yaml:"image,omitempty"`
	Overlay     *BackgroundOverlayPatch  `json:"overlay,omitempty" yaml:"overlay,omitempty"`
	Sections    *BackgroundSectionsPatch `json:"sections,omitempty" yaml:"sections,omitempty"`
}

type BackgroundImagePatch struct {
	Src        *string             `json:"src,omitempty" yaml:"src,omitempty"`
	SampleRate *int                `json:"sampleRate,omitempty" yaml:"sampleRate,omitempty"`
	Sizing     *BackgroundSizing   `json:"sizing,omitempty" yaml:"sizing,omitempty"`
	Position   *BackgroundPosition `json:"position,omitempty" yaml:"position,omitempty"`
	Opacity    *int                `json:"opacity,omitempty" yaml:"opacity,omitempty"`
}

type BackgroundOverlayPatch struct {
	Color   *string `json:"color,omitempty" yaml:"color,omitempty"`
	Opacity *int    `json:"opacity,omitempty" yaml:"opacity,omitempty"`
}

type BackgroundSectionsPatch struct {
	Top    *SectionBackground `json:"top,omitempty" yaml:"top,omitempty"`
	Middle *SectionBackground `json:"middle,omitempty" yaml:"middle,omitempty"`
	Bottom *SectionBackground `json:"bottom,omitempty" yaml:"bottom,omitempty"`
}

func assign[T any](target *T, value *T) {
	if value != nil {
		*target = *value
	}
}

func (p AppPatch) Apply(target *AppConfig) {
	if target == nil {
		return
	}
	if p.General != nil {
		p.General.apply(&target.General)
	}
	if p.UI != nil {
		p.UI.apply(&target.UI)
	}
	if p.Docker != nil {
		assign(&target.Docker.Timeout, p.Docker.Timeout)
		assign(&target.Docker.StatsPollSec, p.Docker.StatsPollSec)
	}
	if p.Runtime != nil {
		p.Runtime.apply(&target.Runtime)
	}
	if p.Keymap != nil {
		p.Keymap.apply(&target.Keymap)
	}
	if p.Logs != nil {
		assign(&target.Logs.Since, p.Logs.Since)
		assign(&target.Logs.Tail, p.Logs.Tail)
		assign(&target.Logs.Timestamps, p.Logs.Timestamps)
	}
	if p.Layout != nil {
		p.Layout.apply(&target.Layout)
	}
	if p.Commands != nil && p.Commands.DockerCompose != nil {
		assign(&target.Commands.DockerCompose.Executable, p.Commands.DockerCompose.Executable)
		assign(&target.Commands.DockerCompose.Subcommand, p.Commands.DockerCompose.Subcommand)
	}
}

func (p GeneralPatch) apply(target *GeneralConfig) {
	assign(&target.ScrollHeight, p.ScrollHeight)
	assign(&target.Reporting, p.Reporting)
	assign(&target.Lang, p.Lang)
	assign(&target.SizeFormat, p.SizeFormat)
}

func (p UIPatch) apply(target *UIConfig) {
	assign(&target.WrapMainPanel, p.WrapMainPanel)
	assign(&target.ReturnImmediately, p.ReturnImmediately)
	assign(&target.ShowHelp, p.ShowHelp)
	assign(&target.HintTimeout, p.HintTimeout)
	assign(&target.EnableMouse, p.EnableMouse)
	if p.Dialog != nil {
		p.Dialog.apply(&target.Dialog)
	}
	if p.Window != nil {
		assign(&target.Window.MarginTopPercent, p.Window.MarginTopPercent)
		assign(&target.Window.MarginBottomPercent, p.Window.MarginBottomPercent)
		assign(&target.Window.ContentWidthPercent, p.Window.ContentWidthPercent)
	}
	if p.Header != nil {
		assign(&target.Header.KeystrokeContentRatio, p.Header.KeystrokeContentRatio)
		if p.Header.Columns != nil {
			assign(&target.Header.Columns.Host, p.Header.Columns.Host)
			assign(&target.Header.Columns.Connection, p.Header.Columns.Connection)
			assign(&target.Header.Columns.Keystroke, p.Header.Columns.Keystroke)
			assign(&target.Header.Columns.Logo, p.Header.Columns.Logo)
		}
	}
	if p.Table != nil {
		assign(&target.Table.ColumnSpacing, p.Table.ColumnSpacing)
		assign(&target.Table.RowPrefix, p.Table.RowPrefix)
		assign(&target.Table.RowPrefixSelected, p.Table.RowPrefixSelected)
		assign(&target.Table.RowSpacing, p.Table.RowSpacing)
		if p.Table.SelectionInfo != nil {
			assign(&target.Table.SelectionInfo.Enabled, p.Table.SelectionInfo.Enabled)
			assign(&target.Table.SelectionInfo.PadLines, p.Table.SelectionInfo.PadLines)
		}
	}
}

func (p DialogLayoutPatch) apply(target *DialogLayoutConfig) {
	if p.Width != nil {
		p.Width.apply(&target.Width)
	}
	if p.Height != nil {
		p.Height.apply(&target.Height)
	}
	if p.Position != nil {
		assign(&target.Position.Horizontal, p.Position.Horizontal)
		assign(&target.Position.Vertical, p.Position.Vertical)
		assign(&target.Position.OffsetX, p.Position.OffsetX)
		assign(&target.Position.OffsetY, p.Position.OffsetY)
	}
	if p.PanelSize != nil {
		p.PanelSize.apply(&target.PanelSize)
	}
}

func (p PanelSizePatch) apply(target *PanelSize) {
	assign(&target.WidthPercent, p.WidthPercent)
	assign(&target.HeightPercent, p.HeightPercent)
	assign(&target.MaxWidth, p.MaxWidth)
	assign(&target.MaxHeight, p.MaxHeight)
}

func (p ResponsiveSizePatch) apply(target *ResponsiveSize) {
	assign(&target.Percent, p.Percent)
	assign(&target.Min, p.Min)
	assign(&target.Max, p.Max)
}

func (p RuntimePatch) apply(target *RuntimeConfig) {
	assign(&target.Default, p.Default)
	if p.Discovery != nil {
		assign(&target.Discovery.LocalDocker, p.Discovery.LocalDocker)
		assign(&target.Discovery.LocalPodman, p.Discovery.LocalPodman)
	}
	if p.Health != nil {
		assign(&target.Health.IntervalSec, p.Health.IntervalSec)
		assign(&target.Health.TimeoutSec, p.Health.TimeoutSec)
		assign(&target.Health.FailureThreshold, p.Health.FailureThreshold)
	}
	if p.Connections != nil {
		target.Connections = append([]RuntimeConnection(nil), (*p.Connections)...)
	}
}

func (p KeymapPatch) apply(target *KeymapConfig) {
	if p.Global != nil {
		assign(&target.Global.Quit, p.Global.Quit)
		assign(&target.Global.ActionBar, p.Global.ActionBar)
		assign(&target.Global.Help, p.Global.Help)
		assign(&target.Global.Filter, p.Global.Filter)
		assign(&target.Global.Refresh, p.Global.Refresh)
	}
	if p.Container != nil {
		assign(&target.Container.Start, p.Container.Start)
		assign(&target.Container.Stop, p.Container.Stop)
		assign(&target.Container.Restart, p.Container.Restart)
		assign(&target.Container.Kill, p.Container.Kill)
		assign(&target.Container.Remove, p.Container.Remove)
		assign(&target.Container.Logs, p.Container.Logs)
		assign(&target.Container.Exec, p.Container.Exec)
		assign(&target.Container.Inspect, p.Container.Inspect)
		assign(&target.Container.Stats, p.Container.Stats)
		assign(&target.Container.Pause, p.Container.Pause)
		assign(&target.Container.Update, p.Container.Update)
		assign(&target.Container.Diff, p.Container.Diff)
		assign(&target.Container.Export, p.Container.Export)
		assign(&target.Container.Commit, p.Container.Commit)
		assign(&target.Container.Wait, p.Container.Wait)
		assign(&target.Container.Copy, p.Container.Copy)
	}
	if p.Image != nil {
		assign(&target.Image.Pull, p.Image.Pull)
		assign(&target.Image.Remove, p.Image.Remove)
		assign(&target.Image.Prune, p.Image.Prune)
		assign(&target.Image.Tag, p.Image.Tag)
		assign(&target.Image.Push, p.Image.Push)
		assign(&target.Image.Save, p.Image.Save)
		assign(&target.Image.Load, p.Image.Load)
		assign(&target.Image.History, p.Image.History)
	}
	if p.Volume != nil {
		p.Volume.apply(&target.Volume)
	}
	if p.Network != nil {
		p.Network.apply(&target.Network)
	}
	if p.Navigation != nil {
		assign(&target.Navigation.TabNext, p.Navigation.TabNext)
		assign(&target.Navigation.TabPrev, p.Navigation.TabPrev)
		assign(&target.Navigation.Up, p.Navigation.Up)
		assign(&target.Navigation.Down, p.Navigation.Down)
		assign(&target.Navigation.Enter, p.Navigation.Enter)
		assign(&target.Navigation.Back, p.Navigation.Back)
		assign(&target.Navigation.Delete, p.Navigation.Delete)
	}
	if p.Dialog != nil {
		assign(&target.Dialog.Confirm, p.Dialog.Confirm)
		assign(&target.Dialog.Cancel, p.Dialog.Cancel)
	}
	if p.Compose != nil {
		assign(&target.Compose.Start, p.Compose.Start)
		assign(&target.Compose.Stop, p.Compose.Stop)
		assign(&target.Compose.Restart, p.Compose.Restart)
		assign(&target.Compose.Down, p.Compose.Down)
		assign(&target.Compose.Logs, p.Compose.Logs)
		assign(&target.Compose.Top, p.Compose.Top)
		assign(&target.Compose.Port, p.Compose.Port)
		assign(&target.Compose.Stats, p.Compose.Stats)
		assign(&target.Compose.Build, p.Compose.Build)
		assign(&target.Compose.Pull, p.Compose.Pull)
		assign(&target.Compose.Push, p.Compose.Push)
		assign(&target.Compose.Scale, p.Compose.Scale)
		assign(&target.Compose.Pause, p.Compose.Pause)
		assign(&target.Compose.Unpause, p.Compose.Unpause)
		assign(&target.Compose.Kill, p.Compose.Kill)
		assign(&target.Compose.Rm, p.Compose.Rm)
		assign(&target.Compose.Prune, p.Compose.Prune)
		assign(&target.Compose.Events, p.Compose.Events)
		assign(&target.Compose.Run, p.Compose.Run)
		assign(&target.Compose.Exec, p.Compose.Exec)
		assign(&target.Compose.ServiceLogs, p.Compose.ServiceLogs)
		assign(&target.Compose.Detail, p.Compose.Detail)
	}
}

func (p ResourceKeymapPatch) apply(target *ResourceKeymap) {
	assign(&target.Create, p.Create)
	assign(&target.Prune, p.Prune)
	assign(&target.Remove, p.Remove)
}

func (p LayoutPatch) apply(target *LayoutConfig) {
	if p.SectionWeights != nil {
		assign(&target.SectionWeights.Top, p.SectionWeights.Top)
		assign(&target.SectionWeights.Content, p.SectionWeights.Content)
		assign(&target.SectionWeights.Bottom, p.SectionWeights.Bottom)
	}
	if p.Background != nil {
		p.Background.apply(&target.Background)
	}
}

func (p BackgroundPatch) apply(target *BackgroundConfig) {
	assign(&target.Enable, p.Enable)
	assign(&target.Fallthrough, p.Fallthrough)
	assign(&target.Type, p.Type)
	assign(&target.Color, p.Color)
	assign(&target.StartColor, p.StartColor)
	assign(&target.EndColor, p.EndColor)
	if p.Image != nil {
		assign(&target.Image.Src, p.Image.Src)
		assign(&target.Image.SampleRate, p.Image.SampleRate)
		assign(&target.Image.Sizing, p.Image.Sizing)
		assign(&target.Image.Position, p.Image.Position)
		assign(&target.Image.Opacity, p.Image.Opacity)
	}
	if p.Overlay != nil {
		assign(&target.Overlay.Color, p.Overlay.Color)
		assign(&target.Overlay.Opacity, p.Overlay.Opacity)
	}
	if p.Sections != nil {
		if p.Sections.Top != nil {
			value := *p.Sections.Top
			target.Sections.Top = &value
		}
		if p.Sections.Middle != nil {
			value := *p.Sections.Middle
			target.Sections.Middle = &value
		}
		if p.Sections.Bottom != nil {
			value := *p.Sections.Bottom
			target.Sections.Bottom = &value
		}
	}
}
