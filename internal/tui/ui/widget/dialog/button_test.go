package dialog

import (
	"strings"
	"testing"

	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
)

func TestButtonRender(t *testing.T) {
	t.Run("normal state shows key and label without indicator", func(t *testing.T) {
		got := NewButton("esc", "Cancel").Render()
		if !strings.Contains(got, "esc") || !strings.Contains(got, "Cancel") {
			t.Fatalf("normal button missing key/label: %q", got)
		}
		if strings.Contains(got, component.ButtonIndicator) {
			t.Fatalf("normal button should not show indicator: %q", got)
		}
	})

	t.Run("focused state shows indicator", func(t *testing.T) {
		b := NewButton("enter", "Confirm")
		b.State = ButtonFocused
		got := b.Render()
		if !strings.Contains(got, component.ButtonIndicator) {
			t.Fatalf("focused button missing indicator: %q", got)
		}
		if !strings.Contains(got, "Confirm") {
			t.Fatalf("focused button missing label: %q", got)
		}
	})

	t.Run("disabled state renders faint", func(t *testing.T) {
		b := NewButton("enter", "Confirm")
		b.State = ButtonDisabled
		got := b.Render()
		if strings.Contains(got, component.ButtonIndicator) {
			t.Fatalf("disabled button should not show indicator: %q", got)
		}
	})

	t.Run("danger state shows indicator", func(t *testing.T) {
		b := NewButton("enter", "Force")
		b.State = ButtonDanger
		got := b.Render()
		if !strings.Contains(got, component.ButtonIndicator) {
			t.Fatalf("danger button missing indicator: %q", got)
		}
	})
}

func TestButtonStateDistinctStyling(t *testing.T) {
	normal := NewButton("k", "Action").Render()
	focused := NewButton("k", "Action")
	focused.State = ButtonFocused
	if normal == focused.Render() {
		t.Fatalf("normal and focused renders must differ")
	}
}
