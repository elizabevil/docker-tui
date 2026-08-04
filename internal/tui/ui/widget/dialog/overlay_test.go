package dialog

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/elizabevil/docker-tui/internal/data/config"
)

func TestPlaceDialogPreservesContentOutsideDialog(t *testing.T) {
	cfg := LoadDialogConfig()
	content := strings.Join([]string{"ABCDEFGHIJ", "ABCDEFGHIJ", "ABCDEFGHIJ"}, "\n")
	got := PlaceDialog(content, "XX", 10, 3, "", cfg)
	lines := strings.Split(got, "\n")
	if len(lines) != 3 {
		t.Fatalf("line count = %d", len(lines))
	}
	if lines[0] != "ABCDEFGHIJ" || lines[2] != "ABCDEFGHIJ" {
		t.Fatalf("unmodified lines = %#v", lines)
	}
	if ansi.StringWidth(lines[1]) != 10 || ansi.Strip(lines[1]) != "ABCDXXGHIJ" {
		t.Fatalf("dialog placement = %q (width %d)", lines[1], ansi.StringWidth(lines[1]))
	}
}

func TestResolveOverlayValidatesAndNormalizesColor(t *testing.T) {
	theme := config.DefaultTheme()
	theme.Dialog.Overlay = config.ValueRef(config.Color("#abc"))
	theme.Dialog.OverlayOpacity = 50
	if got := resolveOverlay(theme); got != "#aabbcc7f" {
		t.Fatalf("resolved overlay = %q", got)
	}
}

func TestResolveOverlayFallsBackFromInvalidColor(t *testing.T) {
	theme := config.DefaultTheme()
	theme.Dialog.Overlay = config.ValueRef(config.Color("not-a-color"))
	if got := resolveOverlay(theme); got != "#0d1117cc" {
		t.Fatalf("fallback overlay = %q", got)
	}
}

func TestPlaceDialogResetsStylesAtSpliceBoundaries(t *testing.T) {
	cfg := LoadDialogConfig()
	content := strings.Join([]string{"ABCDEFGHIJ", "\x1b[48;5;1mABCDEFGHIJ", "ABCDEFGHIJ"}, "\n")
	got := PlaceDialog(content, "XX", 10, 3, "", cfg)
	line := strings.Split(got, "\n")[1]
	if !strings.Contains(line, "\x1b[0mXX\x1b[0m") {
		t.Fatalf("dialog splice must isolate inherited styles: %q", line)
	}
}

func TestDialogBoxAppliesOpaqueSurfaceBackground(t *testing.T) {
	box := DialogBox(DialogStyle{Width: 20, OverlayColor: "#0d1117cc", LeftAligned: true}, "content")
	if !strings.Contains(box, "48;") && !strings.Contains(box, "48:") {
		t.Fatalf("dialog box must define its own background: %q", box)
	}
	if got := opaqueDialogColor("#0d1117cc"); got != "#0d1117" {
		t.Fatalf("opaque dialog color = %q", got)
	}
}
