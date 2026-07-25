//go:build !cgo

package dto

import "time"

// VolumeItem is the Podman Libpod volume inspect response item.
// Field shape and JSON tags match go.podman.io/podman/v6/libpod/define.InspectVolumeData.
type VolumeItem struct {
	Name        string            `json:"Name"`
	Driver      string            `json:"Driver"`
	Mountpoint  string            `json:"Mountpoint"`
	CreatedAt   time.Time         `json:"CreatedAt"`
	Status      map[string]any    `json:"Status,omitempty"`
	Labels      map[string]string `json:"Labels,omitempty"`
	Scope       string            `json:"Scope"`
	Options     map[string]string `json:"Options"`
	UID         int               `json:"UID,omitempty"`
	GID         int               `json:"GID,omitempty"`
	Anonymous   bool              `json:"Anonymous,omitempty"`
	MountCount  uint              `json:"MountCount"`
	NeedsCopyUp bool              `json:"NeedsCopyUp,omitempty"`
	NeedsChown  bool              `json:"NeedsChown,omitempty"`
	Timeout     uint              `json:"Timeout,omitempty"`
	StorageID   string            `json:"StorageID,omitempty"`
	LockNumber  uint32            `json:"LockNumber"`
}
