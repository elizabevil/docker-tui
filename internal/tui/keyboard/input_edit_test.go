package keyboard

import (
	"testing"

	"github.com/elizabevil/docker-tui/internal/tui/state"
)

func TestEditTextInputShellBindings(t *testing.T) {
	text, cursor := "hello world", len([]rune("hello world"))
	if handled, changed := editTextInput("ctrl+w", &text, &cursor); !handled || !changed {
		t.Fatal("ctrl+w was not handled")
	}
	if text != "hello " || cursor != 6 {
		t.Fatalf("ctrl+w result text=%q cursor=%d", text, cursor)
	}
	editTextInput("ctrl+a", &text, &cursor)
	editTextInput("X", &text, &cursor)
	if text != "Xhello " || cursor != 1 {
		t.Fatalf("insert result text=%q cursor=%d", text, cursor)
	}
	editTextInput("ctrl+k", &text, &cursor)
	if text != "X" {
		t.Fatalf("ctrl+k result text=%q", text)
	}
}

func TestEditTextInputUsesRuneCursor(t *testing.T) {
	text, cursor := "世界", 2
	editTextInput("ctrl+h", &text, &cursor)
	if text != "世" || cursor != 1 {
		t.Fatalf("unicode backspace text=%q cursor=%d", text, cursor)
	}
}

func TestCommandInputEditsAtCursor(t *testing.T) {
	app := &state.AppModel{Mode: state.ModeCommand, FilterText: "imags", FilterCursor: 4}
	handleCommandInput("e", app)
	if app.FilterText != "images" || app.FilterCursor != 5 {
		t.Fatalf("command edit text=%q cursor=%d", app.FilterText, app.FilterCursor)
	}
}
