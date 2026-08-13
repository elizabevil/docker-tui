package state

import (
	"strings"

	"github.com/elizabevil/docker-tui/internal/data/audit"
	"github.com/elizabevil/docker-tui/internal/data/config"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/term"
)

const (
	execBatchSize    = 8192
	execBatchTimeout = 16
)

type ExecState struct {
	ExecConn   runtimeapi.ExecSession
	ExecID     string
	ExecCh     chan string
	ExecDone   chan struct{}
	ExecBuf    *term.Buffer
	ExecScroll int
	ExecShell  string
	ExecAudit  audit.Trace

	batchBuf  strings.Builder
	batchSize int
}

func (s *ExecState) SetShell(shell string) {
	if shell == "" {
		shell = config.DefaultShell
	}
	s.ExecShell = shell
}

func (s *ExecState) Start(id string, conn runtimeapi.ExecSession, output chan string, done chan struct{}, trace audit.Trace) {
	s.ExecID, s.ExecConn, s.ExecCh, s.ExecDone, s.ExecAudit = id, conn, output, done, trace
	s.ExecScroll = 0
	s.batchBuf.Reset()
	s.batchSize = 0
}

func (s *ExecState) Append(data string) {
	if s.ExecBuf == nil {
		s.ExecBuf = term.NewBuffer(2000)
	}
	s.batchBuf.WriteString(data)
	s.batchSize += len(data)

	if s.batchSize >= execBatchSize || strings.Contains(data, "\n") || len(data) < 512 {
		s.flushBatch()
	}
}

func (s *ExecState) flushBatch() {
	if s.batchBuf.Len() > 0 {
		s.ExecBuf.Write(s.batchBuf.String())
		s.batchBuf.Reset()
		s.batchSize = 0
	}
}

func (s *ExecState) Flush() {
	s.flushBatch()
}

func (s *ExecState) Reset() {
	s.flushBatch()
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
