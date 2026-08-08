package runtime

import (
	"context"
	"time"
)

// ComposeService 是 compose-级操作的统一接口(per R08-02 §F1 + R08-15 Q3 决策表)。
//
// docker 与 podman adapter 必须各自实现此接口,接口形态完全相同,实现路径独立
// (per R08-15 §0 驱动层架构模型)。方法是否返回 ErrComposeUnsupported 取决于该能力
// 能否在 driver 层 wrap 实现(Q3 决策表):
//
//   方法                   可 wrap 自实现     永久 ErrComposeUnsupported
//   ─────────────────────────────────────────────────────────────────────
//   ListProjects            ✓
//   InspectProject          ✓
//   Start / Stop / Restart  ✓
//   Down                    ✓
//   Up                      ✓(per Q3)
//   Run                     ✓(per Q3 + R08-08)
//   Exec                    ✓(per Q3)
//   Top / Port / Stats      ✓
//   Events                  ✓
//
//   Config                  ✗                   ✓
//
// Build / Pull / Push 不在本接口,在 R08-05 build-transfer 域定义(Build 永久 ErrUnsupported;
// Pull / Push wrap-able)。
type ComposeService interface {
	ListProjects(ctx context.Context) ([]ComposeProjectSummary, error)
	InspectProject(ctx context.Context, project string) (ComposeProjectSummary, error)
	Config(ctx context.Context, project string) (string, error)
	Start(ctx context.Context, project string, services []string) error
	Stop(ctx context.Context, project string, services []string, timeoutSec int) error
	Restart(ctx context.Context, project string, services []string) error
	Down(ctx context.Context, project string, opts DownOptions) error
	Up(ctx context.Context, project string, opts UpOptions) error
	Run(ctx context.Context, project string, service string, command []string, opts RunOptions) error
	Exec(ctx context.Context, project string, service string, command []string) error
	Top(ctx context.Context, project string, service string) (ContainerProcesses, error)
	Port(ctx context.Context, project string, service string, port int) (string, error)
	Stats(ctx context.Context, project string, service string) (ContainerStats, error)
	Events(ctx context.Context, project string) (<-chan ComposeEvent, error)
}

type ComposeProjectSummary struct {
	Name        string
	Services    []ComposeServiceSummary
	WorkingDir  string
	ConfigFiles []string
	Version     string
	CreatedAt   time.Time
	HasRunning  bool
	HasStopped  bool
	Source      string
}

type ComposeServiceSummary struct {
	Name     string
	Image    string
	Replicas int
	Running  int
	Stopped  int
}

type DownOptions struct {
	RemoveVolumes    bool
	RemoveImages     string
	RemoveOrphans    bool
	TimeoutSec       int
	ProjectTimeoutMs int
}

type UpOptions struct {
	Detach        bool
	Build         string
	ForceRecreate bool
	NoDeps        bool
	NoBuild       bool
	RemoveOrphans bool
	Scale         map[string]int
}

type RunOptions struct {
	RemoveAfter        bool
	EntrypointOverride string
	EnvOverrides      map[string]string
	Detach            bool
	User              string
}

type ComposeEvent struct {
	Type      string
	Project   string
	Service   string
	Action    string
	Timestamp time.Time
}
