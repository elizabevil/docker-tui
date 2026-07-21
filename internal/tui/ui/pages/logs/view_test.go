package logs

import (
	"strings"
	"testing"

	"github.com/elizabevil/docker-tui/internal/tui/state"
)

func TestWrapLineUsesRunes(t *testing.T) {
	got := wrapLine("ab世界cd", 3)
	want := []string{"ab世", "界cd"}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("wrapLine() = %#v, want %#v", got, want)
	}
}

func TestRenderViewFitsViewportWhenWrapped(t *testing.T) {
	m := &state.AppModel{
		LogState: state.LogState{LogContainerID: "container", LogContent: []string{strings.Repeat("x", 200), "second"}, LogWrapEnabled: true},
	}
	got := RenderView(m, 8, 60)
	if rows := strings.Count(got, "\n") + 1; rows != 8 {
		t.Fatalf("RenderView rows = %d, want 8", rows)
	}
}

func TestRenderViewCanScrollPastWrappedLine(t *testing.T) {
	m := &state.AppModel{
		LogState: state.LogState{LogContainerID: "container", LogContent: []string{strings.Repeat("x", 200), "second"}, LogWrapEnabled: true, LogViewOffset: 1},
	}
	got := RenderView(m, 8, 60)
	if !strings.Contains(got, "second") {
		t.Fatalf("wrapped viewport did not reach the second raw line: %q", got)
	}
}
