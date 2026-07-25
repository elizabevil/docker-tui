package state

import (
	"context"
	"net"
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/audit"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

type testExecSession struct{ net.Conn }

func (s *testExecSession) ID() string                                            { return "exec-id" }
func (s *testExecSession) Resize(context.Context, runtimeapi.TerminalSize) error { return nil }

func TestExecStateLifecycle(t *testing.T) {
	var s ExecState
	s.SetShell("")
	if s.ExecShell != "/bin/sh" {
		t.Fatalf("default shell = %q", s.ExecShell)
	}

	connection, server := net.Pipe()
	client := &testExecSession{Conn: connection}
	defer client.Close()
	defer server.Close()
	output := make(chan string)
	done := make(chan struct{})
	trace := audit.Trace{ID: "trace", Action: "exec"}
	s.Start("exec-id", client, output, done, trace)
	if s.ExecID != "exec-id" || s.ExecConn != client || s.ExecCh != output || s.ExecDone != done || s.ExecAudit.ID != trace.ID {
		t.Fatalf("started exec = %#v", s)
	}

	s.Append("hello")
	if s.ExecBuf == nil || s.ExecBuf.Len() == 0 {
		t.Fatal("exec output was not appended")
	}
	s.Reset()
	if s.ExecConn != nil || s.ExecID != "" || s.ExecCh != nil || s.ExecDone != nil || s.ExecShell != "" || s.ExecAudit.Valid() {
		t.Fatalf("reset exec = %#v", s)
	}
	if s.ExecBuf == nil || s.ExecBuf.Len() != 0 {
		t.Fatalf("buffer reset left unexpected state: len=%d", s.ExecBuf.Len())
	}
}
