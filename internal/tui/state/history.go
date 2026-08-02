package state

import (
	dockerclient "github.com/elizabevil/docker-tui/internal/data/runtime"
)

// HistoryLoadedMsg is the async reply from historyFetchCmd. The handler
// drops responses whose ImageID does not match the currently selected
// image (user switched images mid-flight).
type HistoryLoadedMsg struct {
	ImageID string
	Layers  []dockerclient.ImageHistoryLayer
	Err     error
}

type HistoryState struct {
	ImageID    string
	ImageRef   string
	Layers     []dockerclient.ImageHistoryLayer
	Loading    bool
	Error      string
	Source     dockerclient.ImageHistorySource
	Cursor     int
	ViewOffset int
	Filter     string
	Filtering  bool
}

func (s *HistoryState) Open(imageID, imageRef string) {
	s.ImageID = imageID
	s.ImageRef = imageRef
	s.Layers = nil
	s.Source = dockerclient.ImageHistoryPending
	s.Error = ""
	s.Loading = true
	s.Cursor = 0
	s.ViewOffset = 0
	s.Filter = ""
	s.Filtering = false
}

func (s *HistoryState) Apply(layers []dockerclient.ImageHistoryLayer, src dockerclient.ImageHistorySource, err error) {
	s.Layers = layers
	s.Loading = false
	if err != nil {
		s.Error = err.Error()
	} else {
		s.Error = ""
		s.Source = src
	}
	s.Cursor = 0
	s.ViewOffset = 0
}

func (s *HistoryState) Close() {
	*s = HistoryState{}
}

func (s *HistoryState) MoveCursor(delta, total, visible int) {
	if total <= 0 {
		s.Cursor = 0
		return
	}
	s.Cursor = min(max(s.Cursor+delta, 0), total-1)
	s.EnsureVisible(visible)
}

func (s *HistoryState) EnsureVisible(visible int) {
	if visible <= 0 {
		s.ViewOffset = 0
		return
	}
	if s.Cursor < s.ViewOffset {
		s.ViewOffset = s.Cursor
	} else if s.Cursor >= s.ViewOffset+visible {
		s.ViewOffset = s.Cursor - visible + 1
	}
	if s.ViewOffset < 0 {
		s.ViewOffset = 0
	}
}

func (s *HistoryState) SetFilter(f string) {
	s.Filter = f
	s.Cursor = 0
	s.ViewOffset = 0
}
