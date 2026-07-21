package state

import (
	"net"

	"github.com/elizabevil/docker-tui/internal/data/audit"
	"github.com/elizabevil/docker-tui/internal/data/config"
	"github.com/elizabevil/docker-tui/internal/data/i18n"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"

	dockerclient "github.com/elizabevil/docker-tui/internal/data/docker"
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

// QueryInputState holds editable text shared by Filter and Search UIs while
// keeping their business state independent.
type QueryInputState struct {
	Text   string
	Cursor int
}

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
	Docker     *dockerclient.Client
	Pool       *dockerclient.ConnectionPool
	AppVersion string
	Audit      *audit.Service

	// Connection state.
	Connecting bool

	// ── Nav ──
	ActivePanel           PanelType
	PrevPanel             PanelType // 进入覆盖层前的面板，返回时恢复
	Mode                  AppMode
	FilterText            string
	FilterCursor          int
	FilterInput           QueryInputState
	SearchInput           QueryInputState
	FilterExitPending     bool
	FilterExitToken       uint64
	ErrorMessage          string
	ErrorCount            int // 累计错误次数，用于状态栏显示
	InfoMessage           string
	AuditOperationMessage string

	// ── Resources ──
	Containers *ContainerListModel
	Images     *ImageListModel
	Volumes    *VolumeListModel
	Networks   *NetworkListModel

	// ── Terminal ──
	Width  int
	Height int

	// ── Connection ──
	Connected               bool
	ConnectionTarget        string
	ConnectionError         string
	RuntimeSelectorCursor   int
	RuntimeSelectorError    map[string]string
	RuntimeSelectorDisabled bool
	HealthFailures          int
	HealthDegraded          bool
	RuntimeType             string
	EngineVersion           string

	// ── Log ──
	LogContainerID string
	LogContent     []string
	LogViewOffset  int
	LogSearchText  string
	LogSearchIdx   int // index of current match in LogSearchMatches
	LogSearchMatch int // line number of current match
	LogWrapEnabled bool

	// ── Stats ──
	StatsActive  bool
	HostCPU      float64
	HostMem      float64
	HostDisk     string
	HostCPUCores int
	HostMemUsed  uint64
	HostMemTotal uint64

	// ── Key Hint ──
	KeyHint      string
	KeyHintTimer int // ticks remaining (1 tick = 100ms)

	// ── Key Stroke Display ──
	KeyStrokeBuffer     []KeyStrokeEvent // buffered key events (max 3)
	KeyStrokeTimer      int              // collection countdown (100ms/tick)
	KeyStrokeAnim       int              // animation frame (0 = done)
	LastKeyStroke       []KeyStrokeEvent // displayed after animation
	KeyStrokeDispTimer  int              // display countdown after animation (0 = destroyed)
	KeyStrokeDuration   int              // 显示持续时间（tick数，默认30=3s）
	KeyStrokeAnimFrames int              // 动画帧数（默认5=500ms）

	// ── Toast ──
	ToastMessage string
	ToastLevel   component.ToastLevel
	ToastTimer   int

	// ── Detail ──
	ImageDetailID      string
	ImageDetailContent string
	ImageDetailData    *dockerclient.ImageDetailData
	DetailTitle        string
	DetailHint         string
	DetailOffset       int
	DetailRawJSON      []byte // raw JSON from Docker inspect API
	DetailSourceType   string       // "section" (default), "yaml", "json"
	DetailResourceType ResourceType // Resource* constant

	// ── Dialog ──
	DialogTitle   string
	DialogBody    string
	DialogPreview string
	DialogAction  string
	DialogFocus   int // 0=confirm, 1=cancel; toggled by Tab
	DialogCursor  int // cursor position within dialog input text

	// ── Confirm ──
	ConfirmAction  string
	ConfirmTarget  string
	ConfirmMessage string
	ConfirmAudit   audit.Trace

	// ── Select ──
	MarkedIDs             map[string]bool
	PendingImagePull      string
	PendingImagePullAudit audit.Trace
	Spinner               *component.Spinner

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
	engineType := ""
	engineVersion := ""
	connected := false
	if client != nil {
		connected = true
		engineType = string(client.RuntimeType)
		engineVersion = client.EngineVersion
	}
	return &AppModel{
		Config:              cfg,
		Docker:              client,
		ActivePanel:         PanelContainers,
		Mode:                ModeNormal,
		Connected:           connected,
		RuntimeType:         engineType,
		EngineVersion:       engineVersion,
		Containers:          NewContainerListModel(),
		Images:              NewImageListModel(),
		Volumes:             NewVolumeListModel(),
		Networks:            NewNetworkListModel(),
		AppVersion:          appVersion,
		Spinner:             component.NewSpinner(),
		KeyStrokeDuration:   30,
		KeyStrokeAnimFrames: 5,
	}
}

// KeyStrokeEvent represents a single key press for the keystroke display column.
type KeyStrokeEvent struct {
	Key    string
	Action string
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
