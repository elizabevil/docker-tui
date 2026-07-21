package state

import "unicode"

// NavigationState owns top-level routing and each navigation input buffer.
// It is embedded during migration; the final AppModel will use a named field.
type NavigationState struct {
	ActivePanel       PanelType
	PrevPanel         PanelType
	Mode              AppMode
	FilterInput       QueryInputState
	SearchInput       QueryInputState
	CommandInput      QueryInputState
	FilterExitPending bool
	FilterExitToken   uint64
}

func NewNavigationState() NavigationState {
	return NavigationState{ActivePanel: PanelContainers, Mode: ModeNormal}
}

func (s *NavigationState) EnterMode(mode AppMode) {
	s.Mode = mode
}

func (s *NavigationState) LeaveMode() {
	s.Mode = ModeNormal
}

func (s *NavigationState) BeginFilterExit() uint64 {
	s.FilterExitPending = true
	s.FilterExitToken++
	return s.FilterExitToken
}

func (s *NavigationState) ClearFilterExit() {
	s.FilterExitPending = false
}

func (s *NavigationState) CancelFilterExit() {
	s.FilterExitPending = false
	s.FilterExitToken++
}

// QueryInputState owns editable text and its rune-based cursor.
type QueryInputState struct {
	Text   string
	Cursor int
}

func NewQueryInput(text string) QueryInputState {
	return QueryInputState{Text: text, Cursor: len([]rune(text))}
}

func (q *QueryInputState) Set(text string) {
	q.Text = text
	q.Cursor = len([]rune(text))
}

func (q *QueryInputState) Reset() {
	q.Text = ""
	q.Cursor = 0
}

func (q *QueryInputState) Clamp() []rune {
	runes := []rune(q.Text)
	q.Cursor = max(0, min(q.Cursor, len(runes)))
	return runes
}

func (q *QueryInputState) Move(delta int) {
	runes := q.Clamp()
	q.Cursor = max(0, min(q.Cursor+delta, len(runes)))
}

func (q *QueryInputState) MoveHome() { q.Cursor = 0 }

func (q *QueryInputState) MoveEnd() { q.Cursor = len([]rune(q.Text)) }

func (q *QueryInputState) Insert(text string) bool {
	inserted := []rune(text)
	if len(inserted) == 0 {
		return false
	}
	runes := q.Clamp()
	runes = append(runes[:q.Cursor], append(inserted, runes[q.Cursor:]...)...)
	q.Cursor += len(inserted)
	q.Text = string(runes)
	return true
}

func (q *QueryInputState) DeleteBackward() bool {
	runes := q.Clamp()
	if q.Cursor == 0 {
		return false
	}
	runes = append(runes[:q.Cursor-1], runes[q.Cursor:]...)
	q.Cursor--
	q.Text = string(runes)
	return true
}

func (q *QueryInputState) DeleteForward() bool {
	runes := q.Clamp()
	if q.Cursor >= len(runes) {
		return false
	}
	runes = append(runes[:q.Cursor], runes[q.Cursor+1:]...)
	q.Text = string(runes)
	return true
}

func (q *QueryInputState) DeleteToStart() bool {
	runes := q.Clamp()
	if q.Cursor == 0 {
		return false
	}
	q.Text = string(runes[q.Cursor:])
	q.Cursor = 0
	return true
}

func (q *QueryInputState) DeleteToEnd() bool {
	runes := q.Clamp()
	if q.Cursor >= len(runes) {
		return false
	}
	q.Text = string(runes[:q.Cursor])
	return true
}

func (q *QueryInputState) DeleteWordBackward() bool {
	return q.DeleteDelimitedBackward(func(r rune) bool { return unicode.IsSpace(r) })
}

func (q *QueryInputState) DeleteDelimitedBackward(delimiter func(rune) bool) bool {
	if delimiter == nil {
		return false
	}
	runes := q.Clamp()
	start := q.Cursor
	for start > 0 && delimiter(runes[start-1]) {
		start--
	}
	for start > 0 && !delimiter(runes[start-1]) {
		start--
	}
	if start == q.Cursor {
		return false
	}
	q.Text = string(append(runes[:start], runes[q.Cursor:]...))
	q.Cursor = start
	return true
}

func (q *QueryInputState) MoveWordBackward() {
	runes := q.Clamp()
	for q.Cursor > 0 && unicode.IsSpace(runes[q.Cursor-1]) {
		q.Cursor--
	}
	for q.Cursor > 0 && !unicode.IsSpace(runes[q.Cursor-1]) {
		q.Cursor--
	}
}

func (q *QueryInputState) MoveWordForward() {
	runes := q.Clamp()
	for q.Cursor < len(runes) && unicode.IsSpace(runes[q.Cursor]) {
		q.Cursor++
	}
	for q.Cursor < len(runes) && !unicode.IsSpace(runes[q.Cursor]) {
		q.Cursor++
	}
}
