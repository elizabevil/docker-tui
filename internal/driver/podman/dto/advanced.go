package dto

// ArchiveChange is a Podman filesystem change descriptor returned by
// /libpod/containers/{id}/changes. Mirrors go.podman.io/image/v5/archive.Change.
type ArchiveChange struct {
	Kind ChangeKind `json:"Kind"`
	Path string     `json:"Path"`
}

// ChangeKind maps to the engine-side change type enum (0 modify, 1 add,
// 2 delete). Defined separately from runtimeapi.ChangeKind so the
// podman driver does not import the runtime package.
type ChangeKind int

const (
	ArchiveChangeModified ChangeKind = 0
	ArchiveChangeAdded    ChangeKind = 1
	ArchiveChangeDeleted  ChangeKind = 2
)

// ContainerUpdateOptions captures the update knobs surfaced to the
// caller. Resource limits (Memory, NanoCPUs) are sent in the
// UpdateEntities JSON body; restart knobs map to the query parameters
// restartPolicy / restartRetries as defined by ContainerUpdateLibpod.
type ContainerUpdateOptions struct {
	Memory            int64
	NanoCPUs          int64
	RestartPolicy     string
	RestartMaxRetries int
}

// ContainerUpdateResponse is the response for
// /libpod/containers/{id}/update. The swagger definition
// (containerUpdateResponse) declares a single ID field.
type ContainerUpdateResponse struct {
	ID string `json:"ID"`
}

// UpdateEntities is the request body for /libpod/containers/{id}/update
// (swagger definitions.UpdateEntities). Only the resource-limit subset
// surfaced by ContainerUpdateOptions is modelled; unrecognized fields
// are ignored by the daemon.
type UpdateEntities struct {
	Memory *LinuxMemory `json:"memory,omitempty"`
	CPU    *LinuxCPU    `json:"cpu,omitempty"`
}

// LinuxMemory mirrors swagger definitions.LinuxMemory. Limit is in bytes.
type LinuxMemory struct {
	Limit             int64  `json:"limit,omitempty"`
	Reservation       int64  `json:"reservation,omitempty"`
	Swap              int64  `json:"swap,omitempty"`
	Kernel            int64  `json:"kernel,omitempty"`
	KernelTCP         int64  `json:"kernelTCP,omitempty"`
	Swappiness        uint64 `json:"swappiness,omitempty"`
	DisableOOMKiller  bool   `json:"disableOOMKiller,omitempty"`
	UseHierarchy      bool   `json:"useHierarchy,omitempty"`
	CheckBeforeUpdate bool   `json:"checkBeforeUpdate,omitempty"`
}

// LinuxCPU mirrors swagger definitions.LinuxCPU. Quota/Period use the
// CFS bandwidth accounting units (microseconds); Cpus is a cpuset list.
type LinuxCPU struct {
	Quota           int64  `json:"quota,omitempty"`
	Period          uint64 `json:"period,omitempty"`
	Shares          uint64 `json:"shares,omitempty"`
	Burst           uint64 `json:"burst,omitempty"`
	Idle            int64  `json:"idle,omitempty"`
	Cpus            string `json:"cpus,omitempty"`
	Mems            string `json:"mems,omitempty"`
	RealtimePeriod  uint64 `json:"realtimePeriod,omitempty"`
	RealtimeRuntime int64  `json:"realtimeRuntime,omitempty"`
}

// ContainerCommitOptions captures parameters for
// /libpod/commit.
type ContainerCommitOptions struct {
	Repository string
	Tag        string
	Comment    string
	Author     string
	Pause      bool
}

// ContainerCommitResponse is the response for the commit endpoint.
type ContainerCommitResponse struct {
	ID string `json:"Id"`
}

// ContainerWaitResponse is the response for
// /libpod/containers/{id}/wait.
type ContainerWaitResponse struct {
	StatusCode int64  `json:"StatusCode"`
	Error      string `json:"Error,omitempty"`
}
