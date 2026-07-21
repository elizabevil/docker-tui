package runtime

import (
	"context"
	"fmt"
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

type ContainerService interface {
	List(context.Context, ContainerListOptions) ([]ContainerSummary, error)
}
