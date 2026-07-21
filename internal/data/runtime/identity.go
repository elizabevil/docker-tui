package runtime

// Type identifies the container engine implementation behind a connection.
type Type string

const (
	Docker Type = "docker"
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
