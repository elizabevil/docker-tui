package state

import "testing"

func TestQueryInputMaintainsRuneCursor(t *testing.T) {
	input := NewQueryInput("世界")
	if input.Cursor != 2 {
		t.Fatalf("cursor = %d, want 2", input.Cursor)
	}
	input.Move(-1)
	if !input.Insert("好") || input.Text != "世好界" || input.Cursor != 2 {
		t.Fatalf("insert result = %#v", input)
	}
	if !input.DeleteBackward() || input.Text != "世界" || input.Cursor != 1 {
		t.Fatalf("backspace result = %#v", input)
	}
	if !input.DeleteForward() || input.Text != "世" || input.Cursor != 1 {
		t.Fatalf("delete result = %#v", input)
	}
}

func TestQueryInputOwnsEditingBoundaries(t *testing.T) {
	input := QueryInputState{Text: "hello world", Cursor: 99}
	input.Clamp()
	if input.Cursor != 11 {
		t.Fatalf("clamped cursor = %d", input.Cursor)
	}
	if !input.DeleteWordBackward() || input.Text != "hello " || input.Cursor != 6 {
		t.Fatalf("delete word result = %#v", input)
	}
	input.DeleteToStart()
	if input.Text != "" || input.Cursor != 0 {
		t.Fatalf("delete to start result = %#v", input)
	}
}

func TestNavigationFilterExitLifecycle(t *testing.T) {
	navigation := NewNavigationState()
	first := navigation.BeginFilterExit()
	second := navigation.BeginFilterExit()
	if !navigation.FilterExitPending || second != first+1 {
		t.Fatalf("filter exit state = %#v", navigation)
	}
	navigation.ClearFilterExit()
	if navigation.FilterExitPending {
		t.Fatal("filter exit remained pending")
	}
}
