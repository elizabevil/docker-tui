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

// ContainerUpdateOptions is the request body for
// /libpod/containers/{id}/update.
type ContainerUpdateOptions struct {
	Memory            int64
	NanoCPUs          int64
	RestartPolicy     string
	RestartMaxRetries int
}

// ContainerUpdateResponse is the response for
// /libpod/containers/{id}/update. The Podman REST API returns a list
// of human-readable warnings.
type ContainerUpdateResponse struct {
	Warnings []string `json:"Warnings"`
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
