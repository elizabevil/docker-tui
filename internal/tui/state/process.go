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
}

func (s *ProcessState) Open(id, name string) {
	*s = ProcessState{ContainerID: id, ContainerName: name, Loading: true}
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
func (s *ProcessState) Close()         { *s = ProcessState{} }
