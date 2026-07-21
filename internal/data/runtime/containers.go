package runtime

import "context"

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

type ContainerService interface {
	List(context.Context, ContainerListOptions) ([]ContainerSummary, error)
}
