package runtime

// VolumeDetail holds the structured inspect result for a volume.
// Both Docker and Podman adapters map their native inspect responses
// into this type, ensuring the UI never parses raw SDK JSON.
type VolumeDetail struct {
	Name       string            `json:"Name"`
	Driver     string            `json:"Driver"`
	Mountpoint string            `json:"Mountpoint"`
	CreatedAt  string            `json:"CreatedAt"`
	Labels     map[string]string `json:"Labels"`
	Scope      string            `json:"Scope"`
	Options    map[string]string `json:"Options"`
	Status     map[string]any    `json:"Status,omitempty"`
}
