package state

import (
	"errors"
	"testing"

	dockerclient "github.com/elizabevil/docker-tui/internal/data/runtime"
)

func sampleLayers() []dockerclient.ImageHistoryLayer {
	return []dockerclient.ImageHistoryLayer{
		{ID: "a", Created: 100, CreatedBy: "RUN /bin/sh", Size: 1, Comment: ""},
		{ID: "b", Created: 200, CreatedBy: "RUN apt-get install vim", Size: 2, Comment: "install"},
		{ID: "c", Created: 300, CreatedBy: "CMD [\"/bin/sh\"]", Size: 0, Comment: ""},
	}
}

func TestHistoryOpenResets(t *testing.T) {
	s := HistoryState{Loading: false, Error: "old", Cursor: 5, ViewOffset: 3, Filter: "x"}
	s.Open("sha256:abc", "nginx:latest")
	if s.ImageID != "sha256:abc" || s.ImageRef != "nginx:latest" || s.Loading != true || s.Error != "" || s.Cursor != 0 || s.ViewOffset != 0 || s.Filter != "" || s.Source != dockerclient.ImageHistoryPending {
		t.Errorf("Open did not reset state: %+v", s)
	}
}

func TestHistoryApplyClearsErrorOnSuccess(t *testing.T) {
	s := HistoryState{Error: "boom"}
	s.Apply(sampleLayers(), dockerclient.ImageHistoryLayerAPI, nil)
	if s.Error != "" {
		t.Errorf("Error after success = %q", s.Error)
	}
	if len(s.Layers) != 3 {
		t.Errorf("Layers = %d, want 3", len(s.Layers))
	}
	if s.Loading {
		t.Errorf("Loading should be false after Apply")
	}
	if s.Source != dockerclient.ImageHistoryLayerAPI {
		t.Errorf("Source = %q, want %q", s.Source, dockerclient.ImageHistoryLayerAPI)
	}
}

func TestHistoryApplyKeepsErrorOnFailure(t *testing.T) {
	s := HistoryState{}
	s.Apply(nil, "", errors.New("socket hung up"))
	if s.Error == "" {
		t.Errorf("Error after failure = empty")
	}
	if s.Layers != nil {
		t.Errorf("Layers should remain nil on failure, got %d", len(s.Layers))
	}
	if s.Loading {
		t.Errorf("Loading should be false after Apply")
	}
}

func TestHistoryMoveCursorClamps(t *testing.T) {
	s := HistoryState{}
	s.MoveCursor(+1, 0, 5)
	if s.Cursor != 0 {
		t.Errorf("empty: Cursor = %d, want 0", s.Cursor)
	}
	s.MoveCursor(+1, 5, 5)
	if s.Cursor != 1 {
		t.Errorf("step 1: Cursor = %d, want 1", s.Cursor)
	}
	s.MoveCursor(+1, 5, 5)
	if s.Cursor != 2 {
		t.Errorf("step 2: Cursor = %d, want 2", s.Cursor)
	}
	s.MoveCursor(+100, 5, 5)
	if s.Cursor != 4 {
		t.Errorf("clamp forward: Cursor = %d, want 4", s.Cursor)
	}
	s.MoveCursor(-100, 5, 5)
	if s.Cursor != 0 {
		t.Errorf("clamp back: Cursor = %d, want 0", s.Cursor)
	}
}

func TestHistoryMoveCursorEnsuresVisible(t *testing.T) {
	// Cursor 4 is visible in [0,5); ViewOffset stays at 0.
	s := HistoryState{}
	s.MoveCursor(+4, 20, 5)
	if s.Cursor != 4 || s.ViewOffset != 0 {
		t.Errorf("step 1: cursor=%d, ViewOffset=%d, want 4/0", s.Cursor, s.ViewOffset)
	}
	// Cursor 9 at end of [ViewOffset, ViewOffset+5) → ViewOffset = 5.
	s.MoveCursor(+5, 20, 5)
	if s.Cursor != 9 || s.ViewOffset != 5 {
		t.Errorf("step 2: cursor=%d, ViewOffset=%d, want 9/5", s.Cursor, s.ViewOffset)
	}
	// Cursor 0 with ViewOffset=5 should scroll back to 0.
	s.Cursor = 0
	s.ViewOffset = 5
	s.MoveCursor(0, 20, 5) // cursor unchanged
	// Manually re-apply ensure for test clarity: cursor 0, ViewOffset 5
	s.EnsureVisible(5)
	if s.ViewOffset != 0 {
		t.Errorf("step 3: ViewOffset = %d, want 0", s.ViewOffset)
	}
}

func TestHistoryEnsureVisibleClamps(t *testing.T) {
	s := HistoryState{Cursor: 0, ViewOffset: 0}
	s.EnsureVisible(0)
	if s.ViewOffset != 0 {
		t.Errorf("visible=0: ViewOffset = %d, want 0", s.ViewOffset)
	}
	s.EnsureVisible(-1)
	if s.ViewOffset != 0 {
		t.Errorf("visible<0: ViewOffset = %d, want 0", s.ViewOffset)
	}
	s.Cursor = 10
	s.ViewOffset = 0
	s.EnsureVisible(5)
	if s.ViewOffset != 6 {
		t.Errorf("cursor past window: ViewOffset = %d, want 6", s.ViewOffset)
	}
}

func TestHistorySetFilterResetsCursor(t *testing.T) {
	s := HistoryState{Cursor: 5, ViewOffset: 3, Filter: "old"}
	s.SetFilter("new")
	if s.Filter != "new" || s.Cursor != 0 || s.ViewOffset != 0 {
		t.Errorf("SetFilter did not reset: %+v", s)
	}
}

func TestHistoryCloseResets(t *testing.T) {
	s := HistoryState{ImageID: "x", Cursor: 5}
	s.Close()
	if s.ImageID != "" || s.Cursor != 0 {
		t.Errorf("Close did not reset: %+v", s)
	}
}
