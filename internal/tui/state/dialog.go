package state

// DialogKind identifies dialog semantics independently from presentation text.
type DialogKind int

const (
	DialogNone DialogKind = iota
	DialogImagePull
	DialogImageExport
	DialogImageDebug
	DialogExec
)

func (k DialogKind) Mode() AppMode {
	switch k {
	case DialogImagePull:
		return ModeImagePull
	case DialogImageExport:
		return ModeExport
	case DialogImageDebug:
		return ModeDebug
	case DialogExec:
		return ModeExec
	default:
		return ModeNormal
	}
}

func (k DialogKind) IsSelection() bool {
	return k == DialogImageExport || k == DialogImageDebug
}

// DialogSpec contains all state needed to open one dialog atomically.
type DialogSpec struct {
	Kind    DialogKind
	Title   string
	Body    string
	Preview string
	Action  string
	Input   string
	Focus   int
}

// DialogState owns dialog content, focus and its independent input buffer.
// It is embedded during migration; the final AppModel will use a named field.
type DialogState struct {
	Kind    DialogKind
	Title   string
	Body    string
	Preview string
	Action  string
	Focus   int
	Input   QueryInputState
}

func (s *DialogState) Open(spec DialogSpec) {
	*s = DialogState{
		Kind:    spec.Kind,
		Title:   spec.Title,
		Body:    spec.Body,
		Preview: spec.Preview,
		Action:  spec.Action,
		Focus:   spec.Focus,
		Input:   NewQueryInput(spec.Input),
	}
}

func (s *DialogState) MoveFocus(delta, count int) {
	if count <= 0 {
		s.Focus = 0
		return
	}
	s.Focus = (s.Focus + delta%count + count) % count
}

func (s *DialogState) Close() {
	*s = DialogState{}
}
