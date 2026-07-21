package state

import "testing"

func TestDialogStateOwnsInputLifecycle(t *testing.T) {
	var dialog DialogState
	dialog.OpenInput("/bin/世界", ExecFocusInput)
	if dialog.DialogFocus != ExecFocusInput || dialog.Input.Cursor != 7 {
		t.Fatalf("opened dialog = %#v", dialog)
	}
	dialog.DialogTitle = "Exec"
	dialog.Reset()
	if dialog.DialogTitle != "" || dialog.Input.Text != "" || dialog.DialogFocus != 0 {
		t.Fatalf("reset dialog = %#v", dialog)
	}
}
