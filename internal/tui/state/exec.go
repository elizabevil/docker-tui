package state

import (
	"net"

	"github.com/elizabevil/docker-tui/internal/data/audit"
	"github.com/elizabevil/docker-tui/internal/tui/term"
)

type ExecState struct {
	ExecConn   net.Conn
	ExecID     string
	ExecCh     chan string
	ExecDone   chan struct{}
	ExecBuf    *term.Buffer
	ExecScroll int
	ExecShell  string
	ExecAudit  audit.Trace
}

func (s *ExecState) SetShell(shell string) {
	if shell == "" {
		shell = "/bin/sh"
	}
	s.ExecShell = shell
}

func (s *ExecState) Start(id string, conn net.Conn, output chan string, done chan struct{}, trace audit.Trace) {
	s.ExecID, s.ExecConn, s.ExecCh, s.ExecDone, s.ExecAudit = id, conn, output, done, trace
	s.ExecScroll = 0
}

func (s *ExecState) Append(data string) {
	if s.ExecBuf == nil {
		s.ExecBuf = term.NewBuffer(2000)
	}
	s.ExecBuf.Write(data)
}

func (s *ExecState) Reset() {
	if s.ExecBuf != nil {
		s.ExecBuf.Reset()
	}
	s.ExecConn = nil
	s.ExecID = ""
	s.ExecCh = nil
	s.ExecDone = nil
	s.ExecScroll = 0
	s.ExecShell = ""
	s.ExecAudit = audit.Trace{}
}
