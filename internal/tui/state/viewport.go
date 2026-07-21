package state

type ViewportState struct {
	Width         int
	Height        int
	HeaderVisible bool
}

func (s *ViewportState) Resize(width, height int) {
	s.Width, s.Height = max(0, width), max(0, height)
}

func (s *ViewportState) ToggleHeader() { s.HeaderVisible = !s.HeaderVisible }
