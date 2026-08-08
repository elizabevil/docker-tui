package state

type ComposeState struct {
	ComposeCursor          int
	ComposeServiceCursor   int
	ComposeProjectFilter   string
	ComposeServiceFilter   string
	ComposeDetailProject   string
	ComposeFocus           int
	ComposeContainerViewID string
	ComposeContainerCursor int

	// ComposeDownRemoveVolumes toggles `-v` on compose down. Held in
	// state so a future Confirm dialog can edit it before firing.
	ComposeDownRemoveVolumes bool

	// ComposeDownRemoveImages is the `--rmi` selector. Empty means
	// "do not remove images"; "local" / "all" per R08-04 F2.
	ComposeDownRemoveImages string

	// ComposeDownRemoveOrphans toggles `--remove-orphans` (R08-04 F2).
	// dtui orphan = container whose service label is not in the
	// project's aggregated service set (R08-04 F3 simplified).
	ComposeDownRemoveOrphans bool

	// ComposeDownSkipConfirm is set to true when the user has already
	// accepted the down confirm dialog so the second entry into
	// doComposeDown runs the actual down without re-opening the
	// dialog. doConfirmYes sets it; the next doComposeDown clears it.
	ComposeDownSkipConfirm bool

	// ComposeServiceTop / Port / Stats hold the latest per-service
	// query results (R08-07). The UI subviews read from these slices;
	// the data path lives in internal/tui/keyboard/compose_action.go
	// and surfaces via ComposeServiceTopLoaded / PortLoaded / StatsLoaded.
	ComposeServiceTop   []ComposeServiceTopItem
	ComposeServicePort  []ComposeServicePortItem
	ComposeServiceStats []ComposeServiceStatsItem

	// ComposeGroupID is the R08-14 CoLocated group selector for the
	// Action Bar's group-level verbs (Shift+Ctrl+D / Shift+Ctrl+R /
	// Shift+Ctrl+E). Empty means "all containers in the project".
	ComposeGroupID string

	// ComposeSubview picks the right-pane sub-renderer (R08-07).
	// The values correspond to the doComposeTop / Port / Stats verbs;
	// empty renders the standard service list.
	ComposeSubview ComposeSubview
}

// ComposeSubview selects the right-pane sub-renderer for the compose
// panel. The default zero value (ComposeSubviewServices) shows the
// regular service list.
type ComposeSubview int

const (
	ComposeSubviewServices ComposeSubview = iota
	ComposeSubviewTop
	ComposeSubviewPort
	ComposeSubviewStats
)

func (s ComposeSubview) String() string {
	switch s {
	case ComposeSubviewTop:
		return "top"
	case ComposeSubviewPort:
		return "port"
	case ComposeSubviewStats:
		return "stats"
	default:
		return "services"
	}
}

func (s *ComposeState) Focus(focus int) { s.ComposeFocus = min(1, max(0, focus)) }

func (s *ComposeState) MoveProject(delta, count int) {
	s.ComposeCursor = boundedCursor(s.ComposeCursor+delta, count)
}

func (s *ComposeState) MoveService(delta, count int) {
	s.ComposeServiceCursor = boundedCursor(s.ComposeServiceCursor+delta, count)
}

func (s ComposeState) ProjectCursor(count int) int {
	return boundedCursor(s.ComposeCursor, count)
}

func (s ComposeState) ServiceCursor(count int) int {
	return boundedCursor(s.ComposeServiceCursor, count)
}

func (s ComposeState) ContainerCursor(count int) int {
	return boundedCursor(s.ComposeContainerCursor, count)
}

func (s *ComposeState) OpenContainers(service string) {
	s.ComposeContainerViewID = service
	s.ComposeContainerCursor = 0
}

func (s *ComposeState) CloseContainers() {
	s.ComposeContainerViewID = ""
	s.ComposeContainerCursor = 0
}

func boundedCursor(cursor, count int) int {
	if count <= 0 {
		return 0
	}
	return min(max(0, cursor), count-1)
}
