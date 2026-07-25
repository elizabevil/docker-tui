package state

import "github.com/elizabevil/docker-tui/internal/data/audit"

type ConfirmState struct {
	ConfirmAction  string
	ConfirmTarget  string
	ConfirmMessage string
	ConfirmAudit   audit.Trace
}

func (s *ConfirmState) Open(action, target, message string, trace audit.Trace) {
	s.ConfirmAction, s.ConfirmTarget, s.ConfirmMessage, s.ConfirmAudit = action, target, message, trace
}

func (s *ConfirmState) Close() { *s = ConfirmState{} }
