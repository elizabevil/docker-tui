package runtime

import (
	"context"
	"fmt"
	"io"
	"time"
)

// Well-known container list filter field names.
const (
	ContainerFilterID       = "id"
	ContainerFilterName     = "name"
	ContainerFilterLabel    = "label"
	ContainerFilterStatus   = "status"
	ContainerFilterAncestor = "ancestor"
	ContainerFilterNetwork  = "network"
	ContainerFilterVolume   = "volume"
)

// ContainerListOptions holds the parameters for listing containers.
type ContainerListOptions struct {
	All     bool
	Limit   int
	Filters FilterSet
}

// NativeFilters converts the filter set to the native Docker/Podman representation.
func (o ContainerListOptions) NativeFilters() (map[string][]string, error) {
	filters := make(map[string][]string, len(o.Filters))
	for field, values := range o.Filters {
		if !isContainerFilter(field) {
			return nil, NewError(ErrorInvalid, "container.list.filter", field, fmt.Errorf("unknown container filter"))
		}
		if len(values) > 0 {
			filters[field] = []string{values[0]}
		}
	}
	return filters, nil
}

func isContainerFilter(field string) bool {
	switch field {
	case ContainerFilterID, ContainerFilterName, ContainerFilterLabel,
		ContainerFilterStatus, ContainerFilterAncestor, ContainerFilterNetwork, ContainerFilterVolume:
		return true
	default:
		return false
	}
}

// ContainerSummary is the runtime-neutral representation of a listed container.
type ContainerSummary struct {
	ID             string
	Name           string
	Image          string
	Status         string
	State          string
	Created        int64
	PortBindings   []PortBinding
	IPs            []string
	MountCount     int
	MountNames     []string
	NetworkNames   []string
	Labels         map[string]string
	ComposeProject string
	ComposeService string
}

// PortBinding maps a container port to a host IP and port.
type PortBinding struct {
	ContainerPort uint16
	Protocol      string
	HostIP        string
	HostPort      uint16
}

// ContainerDetail is the runtime-neutral inspect representation consumed by
// the UI. Adapters own conversion from their native inspect responses.
type ContainerDetail struct {
	ID           string
	Name         string
	Image        string
	Created      string
	Platform     string
	State        ContainerState
	RestartCount int
	Resources    ContainerResources
	Networks     map[string]ContainerNetwork
	Ports        map[string][]ContainerPortBinding
	Mounts       []ContainerMount
	Config       ContainerConfig
}

// ContainerState captures the runtime state of a container.
type ContainerState struct {
	Status     string
	PID        int
	StartedAt  string
	FinishedAt string
}

// ContainerResources describes the resource limits applied to a container.
type ContainerResources struct {
	CPUShares         int64
	Memory            int64
	NanoCPUs          int64
	NetworkMode       string
	RestartPolicy     string
	MaximumRetryCount int
}

// ContainerNetwork holds the IP configuration for a container on a single network.
type ContainerNetwork struct {
	IPAddress  string
	Gateway    string
	MACAddress string
}

// ContainerPortBinding represents a host-side port mapping.
type ContainerPortBinding struct {
	HostIP   string
	HostPort string
}

// ContainerMount describes a volume or bind mount attached to a container.
type ContainerMount struct {
	Source      string
	Destination string
	Mode        string
	ReadWrite   bool
}

// ContainerConfig holds the static configuration of a container.
type ContainerConfig struct {
	WorkingDir   string
	User         string
	Entrypoint   []string
	Command      []string
	Environment  []string
	ExposedPorts []string
	Labels       map[string]string
}

// ContainerProcesses is the output of a top-like process listing.
type ContainerProcesses struct {
	Titles    []string
	Processes [][]string
}

// ContainerStats is one calculated snapshot, not a runtime-specific JSON
// stream. Values use bytes and percentages consistently across drivers.
type ContainerStats struct {
	ReadAt        time.Time
	CPUPercent    float64
	MemoryUsage   float64
	MemoryLimit   float64
	MemoryPercent float64
	NetworkRx     float64
	NetworkTx     float64
}

// ContainerLogOptions controls a finite container log read.
type ContainerLogOptions struct {
	Since      string
	Tail       string
	Timestamps bool
}

// ContainerService provides container lifecycle and inspection operations.
type ContainerService interface {
	List(context.Context, ContainerListOptions) ([]ContainerSummary, error)
	Inspect(context.Context, string) (*ContainerDetail, error)
	Top(context.Context, string) (ContainerProcesses, error)
	Stats(context.Context, string) (ContainerStats, error)
	Logs(context.Context, string, ContainerLogOptions) (io.ReadCloser, error)
}
