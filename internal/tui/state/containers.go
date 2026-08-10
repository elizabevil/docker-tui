package state

import (
	"sort"
	"strings"

	"github.com/elizabevil/docker-tui/internal/data/audit"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

const (
	ActionStarted   = "started"
	ActionStopped   = "stopped"
	ActionRestarted = "restarted"
	ActionKilled    = "killed"
	ActionPaused    = "paused"
	ActionUnpaused  = "unpaused"
	ActionRenamed   = "renamed"
	ActionRemoved   = "removed"
	ActionPulled    = "pulled"
	ActionPruned    = "pruned"
	ActionExported  = "exported"
	ActionCreated   = "created"

	ContainerStateRunning    = "running"
	ContainerStateExited     = "exited"
	ContainerStateCreated    = "created"
	ContainerStateStopped    = "stopped"
	ContainerStateStopping   = "stopping"
	ContainerStatePaused     = "paused"
	ContainerStateDead       = "dead"
	ContainerStateRestarting = "restarting"
)

// ContainerSortColumn identifies which column to sort by on the containers page.
type ContainerSortColumn int

const (
	ContainerSortByName ContainerSortColumn = iota
	ContainerSortByID
	ContainerSortByCPU
	ContainerSortByMem
	ContainerSortByState
	ContainerSortByCreated
)

type (
	ContainersLoaded struct {
		Containers []runtimeapi.ContainerSummary
		Error      error
	}

	// ComposeProjectsLoaded carries the result of FetchComposeProjects
	// (R08-03). Empty Projects + nil Error means "no projects visible"
	// (R08-03 F2: hidden-empty semantics).
	ComposeProjectsLoaded struct {
		Projects []runtimeapi.ComposeProjectSummary
		Error    error
	}

	ContainerActioned struct {
		Action  string
		ID      string
		Success bool
		Error   error
		Audit   audit.Trace
	}

	ContainerBatchActioned struct {
		Action  string
		Success int
		Skipped int
		Failed  int
		Error   error
		Audit   audit.Trace
	}

	// BatchProgressed is emitted between individual targets of a fan-out
	// batch operation so the UI can render live progress (e.g. "2/5
	// containers stopped"). It carries the aggregate counts seen so far.
	BatchProgressed struct {
		Scope   string // e.g. "container.batch.start"
		Action  string // i18n action label for the indicator, e.g. "stopping"
		Total   int    // total targets
		Current int    // targets completed so far
		Success int
		Failed  int
		Skipped int
		Error   error // joined errors from failed operations so far
	}

	// BatchActioned is the cross-panel batch summary emitted by
	// container/image/volume/network batch and compose project operations.
	// It supersedes the per-target single-action messages for actions that
	// fan out across multiple resources so the UI can render one aggregate
	// toast/audit instead of N independent notifications.
	BatchActioned struct {
		Scope      string // e.g. "container.batch.start", "bulk-delete", "compose.start"
		Resource   ResourceType
		Total      int // total targets attempted
		Success    int // succeeded
		Failed     int // failed (engine error)
		Skipped    int // rejected pre-flight (state-incompatible, etc.)
		FailedIDs  []string
		SkippedIDs []string
		Audit      audit.Trace
		Error      error // joined errors from failed operations
	}

	ContainerProcessesLoaded struct {
		ContainerID string
		Processes   runtimeapi.ContainerProcesses
		Error       error
	}

	ContainerProcessesTick struct {
		ContainerID string
		Generation  uint64
	}

	// ComposeServiceTopLoaded carries the per-container top output
	// for `docker compose top <service>` (R08-07 F1).
	ComposeServiceTopLoaded struct {
		Project string
		Service string
		Items   []ComposeServiceTopItem
		Error   error
	}

	// ComposeServicePortLoaded carries the aggregated port map for
	// `docker compose port <service>` (R08-07 F2). HostPort is "" when
	// no container exposes the requested port.
	ComposeServicePortLoaded struct {
		Project string
		Service string
		Port    int
		Items   []ComposeServicePortItem
		Error   error
	}

	// ComposeServiceStatsLoaded is one snapshot of per-container
	// stats for `docker compose stats <project>` (R08-07 F3). The
	// caller schedules the next snapshot via StatsTick.
	ComposeServiceStatsLoaded struct {
		Project string
		Items   []ComposeServiceStatsItem
		Error   error
	}

	// ComposeServiceRunCompleted signals the end of `docker compose
	// run <service>` (R08-08 F2). The oneoff container is created by
	// the engine and either runs detached or holds an exec session;
	// the carrier only carries the final error so the Update loop can
	// surface failures as a toast.
	ComposeServiceRunCompleted struct {
		Project string
		Service string
		Error   error
	}

	// ComposeServicePushCompleted carries the aggregated outcome of
	// `docker compose push <project>` (R08-05).
	ComposeServicePushCompleted struct {
		Project string
		Success int
		Error   error
	}

	// ComposeServicePullCompleted carries the aggregated outcome of
	// `docker compose pull <project>` (R08-05). Image references are
	// deduplicated before pulling, so Success counts distinct images.
	ComposeServicePullCompleted struct {
		Project string
		Success int
		Error   error
	}

	ComposeServiceTopItem struct {
		ContainerID   string
		ContainerName string
		Processes     runtimeapi.ContainerProcesses
	}

	ComposeServicePortItem struct {
		ContainerID   string
		HostIP        string
		HostPort      string
		ContainerPort uint16
		Protocol      string
	}

	ComposeServiceStatsItem struct {
		ContainerID   string
		ContainerName string
		Stats         runtimeapi.ContainerStats
		Error         error
	}

	LogBatchReceived struct {
		ContainerID string
		Lines       []string
		Error       error
	}

	LogStreamError struct {
		ContainerID string
		Error       error
	}

	LogTick struct {
		ContainerID string
	}

	StatsReceived struct {
		ContainerID string
		CPU         float64
		MemUsage    float64
		MemLimit    float64
		MemPerc     float64
		NetRx       float64
		NetTx       float64
		Error       error
	}

	StatsTick       struct{}
	CursorBlinkTick struct{}

	ImageActioned struct {
		Action  string
		Ref     string
		Success bool
		Error   error
		Audit   audit.Trace
	}

	GenericActioned struct {
		Action  string
		ID      string
		Success bool
		Error   error
		Audit   audit.Trace
	}

	ResourcePruned struct {
		ResourceType runtimeapi.ResourceType
		Result       runtimeapi.PruneResult
		Error        error
		Audit        audit.Trace
	}

	ToastTick struct{ Generation uint64 }

	HostStatsTick       struct{}
	RuntimeHealthTick   struct{}
	RuntimeHealthResult struct {
		Name  string
		Error error
	}
	RuntimeProbeResult struct {
		Name  string
		Error error
	}

	ConnectionProbeAll    struct{}
	ConnectionRefreshTick struct{}

	EscTimeout        struct{}
	FilterExitTimeout struct{ Token uint64 }

	KeyHintTick   struct{}
	KeyStrokeTick struct{}
)

type DockerConnected struct {
	Engine runtimeapi.Engine
	Name   string
	Error  error
	Notice string
}

type ExecFinished struct {
	Err error
}

type ExecOutput struct {
	Data string
}

type ExecDone struct{}

type ContainerListModel struct {
	Items             []runtimeapi.ContainerSummary
	Cursor            int
	ViewOffset        int
	Loading           bool
	Error             error
	Stats             map[string]ContainerStats
	Filter            string
	SortBy            ContainerSortColumn
	SortAsc           bool
	SelectionAnchorID string
}

type ContainerStats struct {
	CPU      float64
	MemUsage float64
	MemLimit float64
	MemPerc  float64
	NetRx    float64 // bytes
	NetTx    float64 // bytes
}

func NewContainerListModel() *ContainerListModel {
	return &ContainerListModel{
		Items:  make([]runtimeapi.ContainerSummary, 0),
		Stats:  make(map[string]ContainerStats),
		Cursor: 0,
	}
}

func (m *ContainerListModel) Selected() *runtimeapi.ContainerSummary {
	items := m.FilteredItems()
	if len(items) == 0 || m.Cursor < 0 || m.Cursor >= len(items) {
		return nil
	}
	return &items[m.Cursor]
}

func (m *ContainerListModel) Len() int {
	return len(m.Items)
}

func (m *ContainerListModel) SetFilter(query string) {
	m.Filter = query
}

func (m *ContainerListModel) FilterText() string {
	return m.Filter
}

func (m *ContainerListModel) FilteredItems() []runtimeapi.ContainerSummary {
	if m.Filter == "" {
		return m.Items
	}
	filtered := make([]runtimeapi.ContainerSummary, 0, len(m.Items))
	for _, c := range m.Items {
		if contains(c.Name, m.Filter) ||
			contains(c.ID, m.Filter) ||
			contains(c.Image, m.Filter) ||
			contains(c.Status, m.Filter) {
			filtered = append(filtered, c)
		}
	}
	return filtered
}

// SortedItems returns filtered items sorted by the current SortBy column.
func (m *ContainerListModel) SortedItems() []runtimeapi.ContainerSummary {
	items := m.FilteredItems()
	sortContainerSlice(items, m.SortBy, m.SortAsc, m.Stats)
	return items
}

// TrackCursor returns the index of selectedID in items, or m.Cursor when
// selectedID is empty, or -1 when the item is no longer present.
func (m *ContainerListModel) TrackCursor(items []runtimeapi.ContainerSummary, selectedID string) int {
	if selectedID == "" {
		return m.Cursor
	}
	for i, item := range items {
		if item.ID == selectedID {
			return i
		}
	}
	return -1
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func sortContainerSlice(items []runtimeapi.ContainerSummary, col ContainerSortColumn, asc bool, stats map[string]ContainerStats) {
	sort.SliceStable(items, func(i, j int) bool {
		primary := false
		switch col {
		case ContainerSortByName:
			primary = items[i].Name < items[j].Name
		case ContainerSortByID:
			primary = items[i].ID < items[j].ID
		case ContainerSortByState:
			primary = items[i].State < items[j].State ||
				(items[i].State == items[j].State && items[i].Name < items[j].Name)
		case ContainerSortByCreated:
			primary = items[i].Created < items[j].Created ||
				(items[i].Created == items[j].Created && items[i].Name < items[j].Name)
		case ContainerSortByCPU:
			si, sj := stats[items[i].ID], stats[items[j].ID]
			primary = si.CPU < sj.CPU ||
				(si.CPU == sj.CPU && items[i].Name < items[j].Name)
		case ContainerSortByMem:
			si, sj := stats[items[i].ID], stats[items[j].ID]
			primary = si.MemPerc < sj.MemPerc ||
				(si.MemPerc == sj.MemPerc && items[i].Name < items[j].Name)
		default:
			primary = items[i].Name < items[j].Name
		}
		if asc {
			return primary
		}
		if primary {
			return true
		}
		if items[i].Name != items[j].Name {
			return items[i].Name > items[j].Name
		}
		return items[i].ID > items[j].ID
	})
}
