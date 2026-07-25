package audit

import "encoding/json"

type ContainerMeta struct {
	Image string `json:"image,omitempty"`
	State string `json:"state,omitempty"`
}

type ContainerTarget struct {
	ID   string
	Name string
	Meta ContainerMeta
}

func (t ContainerTarget) TargetType() string { return "container" }
func (t ContainerTarget) TargetID() string   { return t.ID }
func (t ContainerTarget) TargetName() string { return fallbackName(t.Name, t.ID) }
func (t ContainerTarget) ToDTO() TargetDTO   { return targetDTO(t, t.Meta) }

type ImageMeta struct {
	RepoTags []string `json:"repo_tags,omitempty"`
}

type ImageTarget struct {
	ID   string
	Name string
	Meta ImageMeta
}

func (t ImageTarget) TargetType() string { return "image" }
func (t ImageTarget) TargetID() string   { return t.ID }
func (t ImageTarget) TargetName() string { return fallbackName(t.Name, t.ID) }
func (t ImageTarget) ToDTO() TargetDTO   { return targetDTO(t, t.Meta) }

type VolumeMeta struct {
	Driver string `json:"driver,omitempty"`
}

type VolumeTarget struct {
	Name string
	Meta VolumeMeta
}

func (t VolumeTarget) TargetType() string { return "volume" }
func (t VolumeTarget) TargetID() string   { return t.Name }
func (t VolumeTarget) TargetName() string { return t.Name }
func (t VolumeTarget) ToDTO() TargetDTO   { return targetDTO(t, t.Meta) }

type NetworkMeta struct {
	Driver string `json:"driver,omitempty"`
}

type NetworkTarget struct {
	ID   string
	Name string
	Meta NetworkMeta
}

func (t NetworkTarget) TargetType() string { return "network" }
func (t NetworkTarget) TargetID() string   { return t.ID }
func (t NetworkTarget) TargetName() string { return fallbackName(t.Name, t.ID) }
func (t NetworkTarget) ToDTO() TargetDTO   { return targetDTO(t, t.Meta) }

type ComposeMeta struct {
	Containers int `json:"containers,omitempty"`
	Volumes    int `json:"volumes,omitempty"`
	Networks   int `json:"networks,omitempty"`
}

type ComposeTarget struct {
	Name string
	Meta ComposeMeta
}

func (t ComposeTarget) TargetType() string { return "compose_project" }
func (t ComposeTarget) TargetID() string   { return t.Name }
func (t ComposeTarget) TargetName() string { return t.Name }
func (t ComposeTarget) ToDTO() TargetDTO   { return targetDTO(t, t.Meta) }

type RuntimeMeta struct {
	Previous string `json:"previous,omitempty"`
}

type RuntimeTarget struct {
	Name string
	Host string
	Meta RuntimeMeta
}

func (t RuntimeTarget) TargetType() string { return "runtime" }
func (t RuntimeTarget) TargetID() string   { return fallbackName(t.Host, t.Name) }
func (t RuntimeTarget) TargetName() string { return fallbackName(t.Name, t.Host) }
func (t RuntimeTarget) ToDTO() TargetDTO   { return targetDTO(t, t.Meta) }

type ExecMeta struct {
	ContainerID string `json:"container_id"`
}

type ExecTarget struct {
	ID   string
	Name string
	Meta ExecMeta
}

func (t ExecTarget) TargetType() string { return "exec_session" }
func (t ExecTarget) TargetID() string   { return fallbackName(t.ID, t.Meta.ContainerID) }
func (t ExecTarget) TargetName() string { return fallbackName(t.Name, t.TargetID()) }
func (t ExecTarget) ToDTO() TargetDTO   { return targetDTO(t, t.Meta) }

func targetDTO(target Target, meta any) TargetDTO {
	raw, _ := json.Marshal(meta) //nolint:errcheck // Meta is opaque user data; marshal failures are not actionable here.
	if string(raw) == "{}" {
		raw = nil
	}
	return TargetDTO{Type: target.TargetType(), ID: target.TargetID(), Name: target.TargetName(), Meta: raw}
}

func fallbackName(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}
