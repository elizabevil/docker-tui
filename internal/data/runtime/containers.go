package runtime

import (
	"context"
	"fmt"
	"time"
)

const (
	ContainerFilterID       = "id"
	ContainerFilterName     = "name"
	ContainerFilterLabel    = "label"
	ContainerFilterStatus   = "status"
	ContainerFilterAncestor = "ancestor"
	ContainerFilterNetwork  = "network"
	ContainerFilterVolume   = "volume"
)

type ContainerListOptions struct {
	All     bool
	Limit   int
	Filters FilterSet
}

func (o ContainerListOptions) NativeFilters() (map[string][]string, error) {
	filters := make(map[string][]string, len(o.Filters))
	for field, values := range o.Filters {
		if !isContainerFilter(field) {
			return nil, NewError(ErrorInvalid, "container.list.filter", field, fmt.Errorf("unknown container filter"))
		}
		if len(values) > 1 {
			return nil, UnsupportedError("container.list.filter." + field)
		}
		if len(values) == 1 {
			filters[field] = append([]string(nil), values...)
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
	Labels         map[string]string
	ComposeProject string
	ComposeService string
}

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

type ContainerState struct {
	Status     string
	PID        int
	StartedAt  string
	FinishedAt string
}

type ContainerResources struct {
	CPUShares         int64
	Memory            int64
	NanoCPUs          int64
	NetworkMode       string
	RestartPolicy     string
	MaximumRetryCount int
}

type ContainerNetwork struct {
	IPAddress  string
	Gateway    string
	MACAddress string
}

type ContainerPortBinding struct {
	HostIP   string
	HostPort string
}

type ContainerMount struct {
	Source      string
	Destination string
	Mode        string
	ReadWrite   bool
}

type ContainerConfig struct {
	WorkingDir   string
	User         string
	Entrypoint   []string
	Command      []string
	Environment  []string
	ExposedPorts []string
	Labels       map[string]string
}

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

type ContainerService interface {
	List(context.Context, ContainerListOptions) ([]ContainerSummary, error)
	Inspect(context.Context, string) (*ContainerDetail, error)
	Top(context.Context, string) (ContainerProcesses, error)
	Stats(context.Context, string) (ContainerStats, error)
}
