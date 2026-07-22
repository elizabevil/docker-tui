package runtime

import (
	"context"
	"io"
)

// ExecOptions configures an exec session inside a running container.
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

// TerminalSize represents the dimensions of a TTY attached to an exec session.
type TerminalSize struct {
	Width  uint
	Height uint
}

// ExecSession is an interactive I/O stream connected to an exec process.
// Callers read/write through the embedded io.ReadWriteCloser and may resize
// the TTY via Resize.
type ExecSession interface {
	io.ReadWriteCloser
	ID() string
	Resize(context.Context, TerminalSize) error
}

// ExecService opens and manages exec sessions in running containers.
type ExecService interface {
	Open(context.Context, string, ExecOptions) (ExecSession, error)
}

// EventOptions configures an event subscription with optional filters.
type EventOptions struct {
	Filters FilterSet
}

// Event is a single runtime event (container, image, network, etc.).
type Event struct {
	ResourceType string
	Action       string
	ActorID      string
	Attributes   map[string]string
	Scope        string
	Time         int64
	TimeNano     int64
}

// EventItem wraps an Event with an optional error from the event stream.
// A non-nil Error indicates the stream has terminated; the Error is the cause.
type EventItem struct {
	Event Event
	Error error
}

// EventService subscribes to real-time runtime events.
type EventService interface {
	Subscribe(context.Context, EventOptions) (<-chan EventItem, error)
}
