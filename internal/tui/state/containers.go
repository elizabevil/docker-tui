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

	ContainerStateRunning  = "running"
	ContainerStateExited   = "exited"
	ContainerStateCreated  = "created"
	ContainerStateStopped  = "stopped"
	ContainerStateStopping = "stopping"
	ContainerStatePaused   = "paused"
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

	// BatchActioned is the cross-panel batch summary emitted by
	// container/image/volume/network batch and compose project operations.
	// It supersedes the per-target single-action messages for actions that
	// fan out across multiple resources so the UI can render one aggregate
	// toast/audit instead of N independent notifications.
	BatchActioned struct {
		Scope      string // e.g. "container.batch.start", "bulk-delete", "compose.start"
		Resource   string // e.g. "container", "image", "volume", "network", "compose_project"
		Total      int    // total targets attempted
		Success    int    // succeeded
		Failed     int    // failed (engine error)
		Skipped    int    // rejected pre-flight (state-incompatible, etc.)
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

	StatsTick struct{}

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

	ToastTick struct{}

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
	var selectedID string
	items := m.FilteredItems()
	if m.Cursor >= 0 && m.Cursor < len(items) {
		selectedID = items[m.Cursor].ID
	}

	sortContainerSlice(items, m.SortBy, m.SortAsc, m.Stats)

	if selectedID != "" {
		for i, item := range items {
			if item.ID == selectedID {
				m.Cursor = i
				break
			}
		}
	}

	return items
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func sortContainerSlice(items []runtimeapi.ContainerSummary, col ContainerSortColumn, asc bool, stats map[string]ContainerStats) {
	sort.SliceStable(items, func(i, j int) bool {
		less := false
		switch col {
		case ContainerSortByName:
			less = items[i].Name < items[j].Name
		case ContainerSortByID:
			less = items[i].ID < items[j].ID
		case ContainerSortByState:
			less = items[i].State < items[j].State
		case ContainerSortByCreated:
			less = items[i].Created < items[j].Created
		case ContainerSortByCPU:
			si, sj := stats[items[i].ID], stats[items[j].ID]
			less = si.CPU < sj.CPU
		case ContainerSortByMem:
			si, sj := stats[items[i].ID], stats[items[j].ID]
			less = si.MemPerc < sj.MemPerc
		default:
			less = items[i].Name < items[j].Name
		}
		if asc {
			return less
		}
		return !less
	})
}
