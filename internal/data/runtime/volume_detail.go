package runtime

// VolumeDetail holds the structured inspect result for a volume.
// Both Docker and Podman adapters map their native inspect responses
// into this type, ensuring the UI never parses raw SDK JSON.
type VolumeDetail struct {
	Name       string
	Driver     string
	Mountpoint string
	CreatedAt  string
	Labels     map[string]string
	Scope      string
	Options    map[string]string
	Status     map[string]any
}
