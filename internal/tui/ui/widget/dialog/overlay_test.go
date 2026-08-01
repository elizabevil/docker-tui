package dialog

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestPlaceDialogPreservesContentOutsideDialog(t *testing.T) {
	cfg := dialogConfig{XOffset: 50, YOffset: 50}
	content := strings.Join([]string{"ABCDEFGHIJ", "ABCDEFGHIJ", "ABCDEFGHIJ"}, "\n")
	got := PlaceDialog(content, "XX", 10, 3, "", cfg)
	lines := strings.Split(got, "\n")
	if len(lines) != 3 {
		t.Fatalf("line count = %d", len(lines))
	}
	if lines[0] != "ABCDEFGHIJ" || lines[2] != "ABCDEFGHIJ" {
		t.Fatalf("unmodified lines = %#v", lines)
	}
	if ansi.StringWidth(lines[1]) != 10 || lines[1] != "ABCDXXGHIJ" {
		t.Fatalf("dialog placement = %q (width %d)", lines[1], ansi.StringWidth(lines[1]))
	}
}
