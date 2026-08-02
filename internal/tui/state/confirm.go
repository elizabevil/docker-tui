package state

import "github.com/elizabevil/docker-tui/internal/data/audit"

type ConfirmState struct {
	ConfirmAction  string
	ConfirmTarget  string
	ConfirmMessage string
	ConfirmAudit   audit.Trace
	Options        []ChoiceOption
	Focus          int
	ReturnMode     AppMode
}

type ChoiceOption struct {
	ID          string
	Label       string
	Description string
	Disabled    bool
}

func (s *ConfirmState) Open(action, target, message string, trace audit.Trace) {
	s.ConfirmAction, s.ConfirmTarget, s.ConfirmMessage, s.ConfirmAudit = action, target, message, trace
	s.Options = []ChoiceOption{
		{ID: "cancel", Label: "Cancel"},
		{ID: "confirm", Label: "Confirm"},
	}
	s.Focus = 0
}

func (s *ConfirmState) MoveFocus(delta int) {
	if len(s.Options) == 0 {
		return
	}
	for i := 0; i < len(s.Options); i++ {
		s.Focus = (s.Focus + delta + len(s.Options)) % len(s.Options)
		if !s.Options[s.Focus].Disabled {
			return
		}
	}
}

func (s *ConfirmState) Close() { *s = ConfirmState{} }
