package keyboard

import (
	"testing"

	"github.com/elizabevil/docker-tui/internal/tui/state"
)

func TestEditTextInputShellBindings(t *testing.T) {
	input := state.NewQueryInput("hello world")
	if handled, changed := editQueryInput("ctrl+w", &input); !handled || !changed {
		t.Fatal("ctrl+w was not handled")
	}
	if input.Text != "hello " || input.Cursor != 6 {
		t.Fatalf("ctrl+w result=%#v", input)
	}
	editQueryInput("ctrl+a", &input)
	editQueryInput("X", &input)
	if input.Text != "Xhello " || input.Cursor != 1 {
		t.Fatalf("insert result=%#v", input)
	}
	editQueryInput("ctrl+k", &input)
	if input.Text != "X" {
		t.Fatalf("ctrl+k result=%#v", input)
	}
}

func TestEditTextInputUsesRuneCursor(t *testing.T) {
	input := state.NewQueryInput("世界")
	editQueryInput("ctrl+h", &input)
	if input.Text != "世" || input.Cursor != 1 {
		t.Fatalf("unicode backspace=%#v", input)
	}
}

func TestCommandInputEditsAtCursor(t *testing.T) {
	app := &state.AppModel{Navigation: state.NavigationState{Mode: state.ModeCommand, CommandInput: state.QueryInputState{Text: "imags", Cursor: 4}}}
	handleCommandInput("e", app)
	if app.Navigation.CommandInput.Text != "images" || app.Navigation.CommandInput.Cursor != 5 {
		t.Fatalf("command edit input=%#v", app.Navigation.CommandInput)
	}
}

func TestExecInputDeletesPathSegment(t *testing.T) {
	app := &state.AppModel{
		Navigation: state.NavigationState{Mode: state.ModeExec},
		Dialog:     state.DialogState{Kind: state.DialogExec, Focus: state.ExecFocusInput, Input: state.NewQueryInput("/usr/bin/sh")},
	}
	handleExecDialogKeys("ctrl+w", app)
	if app.Dialog.Input.Text != "/usr/bin/" {
		t.Fatalf("exec input = %#v", app.Dialog.Input)
	}
}
