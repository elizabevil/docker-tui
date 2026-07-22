package runtime

import (
	"context"
	"io"
)

type ExecOptions struct {
	Command      []string
	Environment  []string
	WorkingDir   string
	User         string
	TTY          bool
	AttachStdin  bool
	AttachStdout bool
	AttachStderr bool
}

type TerminalSize struct {
	Width  uint
	Height uint
}

type ExecSession interface {
	io.ReadWriteCloser
	ID() string
	Resize(context.Context, TerminalSize) error
}

type ExecService interface {
	Open(context.Context, string, ExecOptions) (ExecSession, error)
}

type EventOptions struct {
	Filters FilterSet
}

type Event struct {
	ResourceType string
	Action       string
	ActorID      string
	Attributes   map[string]string
	Scope        string
	Time         int64
	TimeNano     int64
}

type EventItem struct {
	Event Event
	Error error
}

type EventService interface {
	Subscribe(context.Context, EventOptions) (<-chan EventItem, error)
}
