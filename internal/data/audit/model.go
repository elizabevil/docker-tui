package audit

import (
	"context"
	"encoding/json"
	"time"
)

type Result string

const (
	ResultRequested Result = "requested"
	ResultStarted   Result = "started"
	ResultSucceeded Result = "succeeded"
	ResultFailed    Result = "failed"
	ResultPartial   Result = "partial"
	ResultCancelled Result = "cancelled"
)

type Level string

const (
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
)

type RuntimeContext struct {
	Type string `json:"type"`
	Name string `json:"name"`
	Host string `json:"host"`
}

type UIContext struct {
	Surface string `json:"surface,omitempty"`
	View    string `json:"view,omitempty"`
	Mode    string `json:"mode,omitempty"`
}

type TargetDTO struct {
	Type string          `json:"type"`
	ID   string          `json:"id"`
	Name string          `json:"name"`
	Meta json.RawMessage `json:"meta,omitempty"`
}

type Details struct {
	DurationMs int64  `json:"duration_ms,omitempty"`
	Error      string `json:"error,omitempty"`
	Shell      string `json:"shell,omitempty"`
	ExitCode   *int   `json:"exit_code,omitempty"`
}

type Record struct {
	Time    time.Time      `json:"time"`
	TraceID string         `json:"trace_id"`
	EventID string         `json:"event_id"`
	Action  string         `json:"action"`
	Result  Result         `json:"result"`
	Level   Level          `json:"level"`
	Message string         `json:"message"`
	Runtime RuntimeContext `json:"runtime"`
	UI      UIContext      `json:"ui"`
	Target  TargetDTO      `json:"target"`
	Details Details        `json:"details,omitempty"`
}

type Target interface {
	TargetType() string
	TargetID() string
	TargetName() string
	ToDTO() TargetDTO
}

type Sink interface {
	WriteAudit(context.Context, Record) error
}

type Trace struct {
	ID        string
	Action    string
	StartedAt time.Time
	Runtime   RuntimeContext
	UI        UIContext
	Target    TargetDTO
}

func (t Trace) Valid() bool { return t.ID != "" && t.Action != "" }

type UIMessage struct {
	Level   Level
	Message string
}

// SessionTarget represents the application session itself, used to log
// startup, shutdown, and runtime-connection lifecycle events that are not
// tied to a specific container, image, volume, etc.
type SessionTarget struct {
	ID   string
	Name string
	Meta SessionMeta
}

type SessionMeta struct {
	Version string `json:"version,omitempty"`
	OS      string `json:"os,omitempty"`
	Arch    string `json:"arch,omitempty"`
}

func (t SessionTarget) TargetType() string { return "session" }
func (t SessionTarget) TargetID() string   { return t.ID }
func (t SessionTarget) TargetName() string { return t.Name }
func (t SessionTarget) ToDTO() TargetDTO {
	raw, _ := json.Marshal(t.Meta)
	return TargetDTO{Type: "session", ID: t.ID, Name: t.Name, Meta: raw}
}
