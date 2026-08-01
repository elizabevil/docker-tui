package state

import "testing"

func TestActionBarLifecycle(t *testing.T) {
	state := ActionBarState{Filter: "old", Selected: 4, Filtering: true}
	state.Open()
	if !state.Visible || state.Filtering || state.Filter != "" || state.Selected != 0 {
		t.Fatalf("Open() = %#v", state)
	}
	state.EnterFilter()
	state.SetFilter("image")
	if !state.Filtering || state.Filter != "image" || state.Selected != 0 {
		t.Fatalf("filter state = %#v", state)
	}
	state.ExitFilter()
	if state.Filtering || state.Filter != "image" {
		t.Fatalf("ExitFilter() should preserve query: %#v", state)
	}
	state.Close()
	if state != (ActionBarState{}) {
		t.Fatalf("Close() = %#v", state)
	}
}

func TestActionBarMoveSelectionWraps(t *testing.T) {
	state := ActionBarState{}
	state.MoveSelection(-1, 3)
	if state.Selected != 2 {
		t.Fatalf("wrapped selection = %d", state.Selected)
	}
	state.MoveSelection(1, 3)
	if state.Selected != 0 {
		t.Fatalf("wrapped selection = %d", state.Selected)
	}
	state.MoveSelection(1, 0)
	if state.Selected != 0 {
		t.Fatalf("empty selection = %d", state.Selected)
	}
}
