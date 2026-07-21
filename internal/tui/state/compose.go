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
}

func (s *ComposeState) Focus(focus int) { s.ComposeFocus = min(1, max(0, focus)) }

func (s *ComposeState) MoveProject(delta, count int) {
	s.ComposeCursor = boundedCursor(s.ComposeCursor+delta, count)
}

func (s *ComposeState) MoveService(delta, count int) {
	s.ComposeServiceCursor = boundedCursor(s.ComposeServiceCursor+delta, count)
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
