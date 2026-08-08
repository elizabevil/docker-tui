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

// ChoiceOption is one row in a confirm dialog.
//
// The Checked field carries a checkbox state (R08-04 F2 down dialog).
// Pure Cancel / Confirm dialogs leave it false on every option; the
// R08-04 down dialog uses it to surface the `-v` / `--rmi` /
// `--remove-orphans` toggles alongside the existing Confirm / Cancel
// buttons.
type ChoiceOption struct {
	ID          string
	Label       string
	Description string
	Disabled    bool
	Checked     bool
}

func (s *ConfirmState) Open(action, target, message string, trace audit.Trace) {
	s.ConfirmAction, s.ConfirmTarget, s.ConfirmMessage, s.ConfirmAudit = action, target, message, trace
	s.Options = []ChoiceOption{
		{ID: "cancel", Label: "Cancel"},
		{ID: "confirm", Label: "Confirm"},
	}
	s.Focus = 0
}

// OpenWithOptions replaces the default Cancel / Confirm pair with a
// caller-supplied slice. Use when the dialog carries extra checkboxes
// (R08-04 F2) or any non-default button layout.
func (s *ConfirmState) OpenWithOptions(action, target, message string, trace audit.Trace, options []ChoiceOption) {
	s.ConfirmAction, s.ConfirmTarget, s.ConfirmMessage, s.ConfirmAudit = action, target, message, trace
	s.Options = options
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
