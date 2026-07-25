package dto

// ExecCreateRequest is the Podman Libpod REST request body for creating an
// exec session inside a running container.
type ExecCreateRequest struct {
	AttachStdin  bool     `json:"AttachStdin"`
	AttachStdout bool     `json:"AttachStdout"`
	AttachStderr bool     `json:"AttachStderr"`
	TTY          bool     `json:"Tty"`
	Command      []string `json:"Cmd"`
	Environment  []string `json:"Env,omitempty"`
	WorkingDir   string   `json:"WorkingDir,omitempty"`
	User         string   `json:"User,omitempty"`
}

// ExecCreateResponse is the Podman Libpod REST response after creating an
// exec session.
type ExecCreateResponse struct {
	ID string `json:"Id"`
}

// ExecStartRequest is the Podman Libpod REST request body for starting an
// exec session.
type ExecStartRequest struct {
	Detach bool `json:"Detach"`
	TTY    bool `json:"Tty"`
}
