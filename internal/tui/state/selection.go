package state

import "github.com/elizabevil/docker-tui/internal/data/audit"

type SelectionState struct {
	PanelMarks            map[PanelType]map[string]bool
	PendingImagePull      string
	PendingImagePullAudit audit.Trace
}

func (s *SelectionState) Toggle(panel PanelType, id string) {
	if s.PanelMarks == nil {
		s.PanelMarks = make(map[PanelType]map[string]bool)
	}
	marks := s.PanelMarks[panel]
	if marks == nil {
		marks = make(map[string]bool)
		s.PanelMarks[panel] = marks
	}
	if marks[id] {
		delete(marks, id)
	} else {
		marks[id] = true
	}
}

func (s *SelectionState) ClearPanelMarks(panel PanelType) {
	if s.PanelMarks != nil {
		delete(s.PanelMarks, panel)
	}
}

func (s *SelectionState) MarkedCount(panel PanelType) int {
	return len(s.PanelMarks[panel])
}

func (s *SelectionState) ClearMarks() {
	s.PanelMarks = make(map[PanelType]map[string]bool)
}

func (s *SelectionState) QueueImagePull(ref string, trace audit.Trace) {
	s.PendingImagePull, s.PendingImagePullAudit = ref, trace
}

func (s *SelectionState) TakeImagePull() (string, audit.Trace) {
	ref, trace := s.PendingImagePull, s.PendingImagePullAudit
	s.PendingImagePull, s.PendingImagePullAudit = "", audit.Trace{}
	return ref, trace
}
