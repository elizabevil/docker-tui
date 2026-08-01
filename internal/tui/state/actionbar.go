package state

// ActionBarState owns the runtime state of the per-page Action Bar.
// Pure state only — no key references — so the state package has no
// dependency on internal/tui/keys (avoids import cycle).
//
// Field naming note: Visible (not Open) because Go does not allow a
// field and a method to share the same name on a struct; using
// Visible keeps Open() and Close() as the public lifecycle methods.
//
// The widget layer (ui/widget/actionbar) translates this state into
// the typed []ActionItem list and back.
type ActionBarState struct {
	Visible   bool   // whether the Action Bar overlay is shown
	Filtering bool   // whether the user is currently typing into the filter input
	Filter    string // fuzzy / substring filter; "" means no filter
	Selected  int    // cursor index over the current filtered action list
}

// Open resets transient state and marks the bar visible. Filtering
// is forced to false (browse mode, not filter input mode).
func (s *ActionBarState) Open() {
	s.Visible = true
	s.Filtering = false
	s.Filter = ""
	s.Selected = 0
}

// Close hides the bar and resets transient state. Safe to call
// when already closed.
func (s *ActionBarState) Close() {
	s.Visible = false
	s.Filtering = false
	s.Filter = ""
	s.Selected = 0
}

// EnterFilter switches the bar to filter-input mode without closing it.
// Used by handleActionBarKeys when the user presses '/'.
func (s *ActionBarState) EnterFilter() {
	s.Filtering = true
	s.Selected = 0
}

// ExitFilter leaves filter-input mode and discards the typed filter.
// The bar stays open (Visible remains true) so the user returns to
// browsing the full action list.
func (s *ActionBarState) ExitFilter() {
	s.Filtering = false
	s.Selected = 0
}

// Reset clears the entire state to the zero value.
func (s *ActionBarState) Reset() {
	*s = ActionBarState{}
}

// MoveSelection moves the cursor by delta, wrapping inside [0, total).
// When total <= 0, Selected is reset to 0. Only meaningful when
// Filtering == false; while in filter-input mode the keyboard
// handler should route printable keys into Filter instead.
func (s *ActionBarState) MoveSelection(delta, total int) {
	if total <= 0 {
		s.Selected = 0
		return
	}
	s.Selected = (s.Selected + delta + total) % total
}

// SetFilter updates the filter string and resets the cursor to 0.
// The keyboard handler calls this on each character typed in filter
// mode; VisibleItems re-filters the action list from this.
func (s *ActionBarState) SetFilter(s2 string) {
	s.Filter = s2
	s.Selected = 0
}
