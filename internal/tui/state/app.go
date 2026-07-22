package state

import (
	"github.com/elizabevil/docker-tui/internal/data/config"
	dockerclient "github.com/elizabevil/docker-tui/internal/data/docker"
	"github.com/elizabevil/docker-tui/internal/data/i18n"
)

// PanelType identifies which panel is active.
type PanelType int

const (
	PanelContainers PanelType = iota
	PanelImages
	PanelVolumes
	PanelNetworks
	PanelCompose
	PanelLogs
	PanelDetail
	PanelHelp
)

// ResourceType identifies the type of resource in the detail view.
type ResourceType string

const (
	ResourceContainer ResourceType = "container"
	ResourceNetwork   ResourceType = "network"
	ResourceVolume    ResourceType = "volume"
	ResourceImage     ResourceType = "image"
)

// AppMode represents the current UI mode.
type AppMode int

const (
	ModeNormal AppMode = iota
	ModeFilter
	ModeSearch
	ModeImagePull
	ModeLogView
	ModeDetail
	ModeHelp
	ModeConfirm
	ModeExport
	ModeDebug
	ModeExec
	ModeMark
	ModeCommand
	ModeExecShell
	ModeExecPassthrough
	ModeRuntimeSelect
	ModeRename
	ModeTop
	ModeResourceCreate
)

// ExecDialogFocus identifies focusable regions in the exec dialog.
const (
	ExecFocusShell1 = iota
	ExecFocusShell2
	ExecFocusShell3
	ExecFocusInput
	ExecFocusConfirm
	ExecFocusCancel
	ExecFocusCount
)

// AppModel is the top-level application model.
// Fields are grouped by concern:
//
//	Config — app configuration, theme, Docker client
//	Nav    — panel, mode, filter, messages
//	Resources — container/image/volume/network lists
//	Terminal  — width, height
//	State     — UI transient state (toast, dialog, detail, compose, etc.)
type AppModel struct {
	Dependencies Dependencies
	Connection   ConnectionState
	Navigation   NavigationState
	Dialog       DialogState
	Feedback     FeedbackState
	Log          LogState
	Detail       DetailState
	Exec         ExecState
	Compose      ComposeState
	Confirm      ConfirmState
	Selection    SelectionState
	Resources    ResourceState
	Metrics      MetricsState
	Viewport     ViewportState
	Processes    ProcessState
}

// NewAppModel creates a new application model with default state.
func NewAppModel(cfg *config.Config, client *dockerclient.Client, appVersion string) *AppModel {
	return &AppModel{
		Dependencies: Dependencies{Config: cfg, AppVersion: appVersion},
		Connection:   NewConnectionState(client),
		Navigation:   NewNavigationState(),
		Feedback:     NewFeedbackState(),
		Resources:    NewResourceState(),
	}
}

// PanelLabel returns a display label for a panel type.
func PanelLabel(p PanelType) string {
	switch p {
	case PanelContainers:
		return i18n.T("panel.containers")
	case PanelImages:
		return i18n.T("panel.images")
	case PanelVolumes:
		return i18n.T("panel.volumes")
	case PanelNetworks:
		return i18n.T("panel.networks")
	case PanelCompose:
		return i18n.T("panel.compose")
	case PanelLogs:
		return i18n.T("panel.logs")
	case PanelDetail:
		return i18n.T("panel.detail")
	case PanelHelp:
		return i18n.T("panel.help")
	default:
		return "Unknown"
	}
}

// PanelList returns all navigable resource panels in order.
func PanelList() []PanelType {
	return []PanelType{
		PanelContainers,
		PanelImages,
		PanelVolumes,
		PanelNetworks,
		PanelCompose,
	}
}
