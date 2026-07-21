package state

// DialogState owns dialog content, focus and its independent input buffer.
// It is embedded during migration; the final AppModel will use a named field.
type DialogState struct {
	DialogTitle   string
	DialogBody    string
	DialogPreview string
	DialogAction  string
	DialogFocus   int
	Input         QueryInputState
}

func (s *DialogState) OpenInput(text string, focus int) {
	s.DialogFocus = focus
	s.Input.Set(text)
}

func (s *DialogState) Reset() {
	*s = DialogState{}
}
