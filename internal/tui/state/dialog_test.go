package state

import "testing"

func TestDialogStateOpensAtomically(t *testing.T) {
	var dialog DialogState
	dialog.Open(DialogSpec{
		Kind:    DialogExec,
		Title:   "Exec",
		Body:    "Choose a shell",
		Preview: "/bin/世界",
		Action:  "Run",
		Input:   "/bin/世界",
		Focus:   ExecFocusInput,
	})
	if dialog.Kind != DialogExec || dialog.Kind.Mode() != ModeExec {
		t.Fatalf("dialog kind = %v", dialog.Kind)
	}
	if dialog.Focus != ExecFocusInput || dialog.Input.Cursor != 7 {
		t.Fatalf("opened dialog = %#v", dialog)
	}
	if dialog.Title != "Exec" || dialog.Action != "Run" {
		t.Fatalf("dialog content = %#v", dialog)
	}
}

func TestDialogStateFocusAndCloseLifecycle(t *testing.T) {
	dialog := DialogState{Kind: DialogImageExport, Focus: 0, Title: "Export"}
	dialog.MoveFocus(-1, 2)
	if dialog.Focus != 1 {
		t.Fatalf("wrapped focus = %d", dialog.Focus)
	}
	dialog.Close()
	if dialog.Kind != DialogNone || dialog.Title != "" || dialog.Input.Text != "" || dialog.Focus != 0 {
		t.Fatalf("closed dialog = %#v", dialog)
	}
}
