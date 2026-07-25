//go:build !cgo

package dto

import (
	"time"
)

// ContainerItem is the Podman Libpod container list response item.
// Field shape and JSON tags match go.podman.io/podman/v6/pkg/domain/entities/types.ListContainer.
type ContainerItem struct {
	AutoRemove   bool                `json:"AutoRemove,omitempty"`
	Command      []string            `json:"Command,omitempty"`
	Created      time.Time           `json:"Created"`
	CreatedAt    string              `json:"CreatedAt,omitempty"`
	CIDFile      string              `json:"CIDFile,omitempty"`
	Exited       bool                `json:"Exited"`
	ExitedAt     int64               `json:"ExitedAt"`
	ExitCode     int32               `json:"ExitCode"`
	ExposedPorts map[uint16][]string `json:"ExposedPorts,omitempty"`
	ID           string              `json:"Id"`
	Image        string              `json:"Image"`
	ImageID      string              `json:"ImageID,omitempty"`
	IsInfra      bool                `json:"IsInfra,omitempty"`
	Labels       map[string]string   `json:"Labels,omitempty"`
	Mounts       []string            `json:"Mounts,omitempty"`
	Names        []string            `json:"Names,omitempty"`
	Networks     []string            `json:"Networks,omitempty"`
	Pid          int                 `json:"Pid"`
	Pod          string              `json:"Pod,omitempty"`
	PodName      string              `json:"PodName,omitempty"`
	Ports        []ContainerPort     `json:"Ports,omitempty"`
	Restarts     uint                `json:"Restarts"`
	StartedAt    int64               `json:"StartedAt"`
	State        string              `json:"State"`
	Status       string              `json:"Status"`
}

// ContainerPort mirrors go.podman.io/common/libnetwork/types.PortMapping.
type ContainerPort struct {
	HostIP        string `json:"host_ip,omitempty"`
	ContainerPort uint16 `json:"container_port"`
	HostPort      uint16 `json:"host_port"`
	Range         uint16 `json:"range"`
	Protocol      string `json:"protocol,omitempty"`
}
