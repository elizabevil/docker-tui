package state

import (
	"net"

	"github.com/elizabevil/docker-tui/internal/data/audit"
	"github.com/elizabevil/docker-tui/internal/data/config"
	dockerclient "github.com/elizabevil/docker-tui/internal/data/docker"
	"github.com/elizabevil/docker-tui/internal/data/i18n"
	"github.com/elizabevil/docker-tui/internal/tui/term"
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
	// ── Config ──
	Config     *config.Config
	Theme      *config.Theme
	AppVersion string
	Audit      *audit.Service
	ConnectionState
	NavigationState
	DialogState
	FeedbackState
	LogState
	DetailState

	// ── Nav ──

	// ── Resources ──
	Containers *ContainerListModel
	Images     *ImageListModel
	Volumes    *VolumeListModel
	Networks   *NetworkListModel

	// ── Terminal ──
	Width  int
	Height int

	// ── Stats ──
	StatsActive  bool
	HostCPU      float64
	HostMem      float64
	HostDisk     string
	HostCPUCores int
	HostMemUsed  uint64
	HostMemTotal uint64

	// ── Confirm ──
	ConfirmAction  string
	ConfirmTarget  string
	ConfirmMessage string
	ConfirmAudit   audit.Trace

	// ── Select ──
	MarkedIDs             map[string]bool
	PendingImagePull      string
	PendingImagePullAudit audit.Trace

	// ── Exec Passthrough ──
	ExecConn   net.Conn      // hijacked connection for exec stdin
	ExecID     string        // exec session ID (for resize)
	ExecCh     chan string   // output channel for exec data
	ExecDone   chan struct{} // closed when exec finishes
	ExecBuf    *term.Buffer  // terminal output buffer with scrollback
	ExecScroll int           // scroll offset (0 = bottom, >0 = scrolled up)
	ExecShell  string        // shell to use for exec
	ExecAudit  audit.Trace

	// ── Header ──
	HeaderVisible bool
	EscPending    bool

	// ── Compose ──
	ComposeCursor          int // 左栏：项目列表光标
	ComposeServiceCursor   int // 右栏：服务列表光标
	ComposeProjectFilter   string
	ComposeServiceFilter   string
	ComposeDetailProject   string
	ComposeFocus           int    // 0=左栏(项目), 1=右栏(服务)
	ComposeContainerViewID string // 非空时显示容器子视图
	ComposeContainerCursor int    // 容器子视图光标
}

// NewAppModel creates a new application model with default state.
func NewAppModel(cfg *config.Config, client *dockerclient.Client, appVersion string) *AppModel {
	return &AppModel{
		Config:          cfg,
		ConnectionState: NewConnectionState(client),
		NavigationState: NewNavigationState(),
		FeedbackState:   NewFeedbackState(),
		Containers:      NewContainerListModel(),
		Images:          NewImageListModel(),
		Volumes:         NewVolumeListModel(),
		Networks:        NewNetworkListModel(),
		AppVersion:      appVersion,
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
