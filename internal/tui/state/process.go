package state

type ProcessState struct {
	ContainerID   string
	ContainerName string
	Titles        []string
	Rows          [][]string
	Cursor        int
	ViewOffset    int
	Loading       bool
	Error         string
	Generation    uint64
}

func (s *ProcessState) Open(id, name string) {
	generation := s.Generation + 1
	*s = ProcessState{ContainerID: id, ContainerName: name, Loading: true, Generation: generation}
}

func (s *ProcessState) Apply(id string, titles []string, rows [][]string, err error) bool {
	if id != s.ContainerID {
		return false
	}
	s.Loading = false
	if err != nil {
		s.Error = err.Error()
		return true
	}
	s.Error = ""
	s.Titles = append([]string(nil), titles...)
	s.Rows = rows
	s.Cursor = boundedCursor(s.Cursor, len(rows))
	return true
}

func (s *ProcessState) Move(delta int) { s.Cursor = boundedCursor(s.Cursor+delta, len(s.Rows)) }
func (s *ProcessState) Close() {
	generation := s.Generation + 1
	*s = ProcessState{Generation: generation}
}
