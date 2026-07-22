package runtime

// Type identifies the container engine implementation behind a connection.
type Type string

const (
	// Docker identifies a Docker Engine backend.
	Docker Type = "docker"
	// Podman identifies a Podman backend.
	Podman Type = "podman"
)

// Identity describes a connected engine without exposing SDK-specific types.
type Identity struct {
	Type       Type
	Name       string
	Endpoint   string
	Version    string
	APIVersion string
}
