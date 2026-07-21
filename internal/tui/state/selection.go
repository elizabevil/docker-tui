package state

import "github.com/elizabevil/docker-tui/internal/data/audit"

type SelectionState struct {
	MarkedIDs             map[string]bool
	PendingImagePull      string
	PendingImagePullAudit audit.Trace
}

func (s *SelectionState) Toggle(id string) {
	if s.MarkedIDs == nil {
		s.MarkedIDs = make(map[string]bool)
	}
	if s.MarkedIDs[id] {
		delete(s.MarkedIDs, id)
	} else {
		s.MarkedIDs[id] = true
	}
}

func (s *SelectionState) ClearMarks() { s.MarkedIDs = make(map[string]bool) }

func (s *SelectionState) QueueImagePull(ref string, trace audit.Trace) {
	s.PendingImagePull, s.PendingImagePullAudit = ref, trace
}

func (s *SelectionState) TakeImagePull() (string, audit.Trace) {
	ref, trace := s.PendingImagePull, s.PendingImagePullAudit
	s.PendingImagePull, s.PendingImagePullAudit = "", audit.Trace{}
	return ref, trace
}
