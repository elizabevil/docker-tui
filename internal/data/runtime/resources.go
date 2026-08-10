package runtime

import (
	"context"
	"fmt"
	"time"
)

// Well-known filter field names for volume and network list operations.
const (
	VolumeFilterName     = "name"
	VolumeFilterDriver   = "driver"
	VolumeFilterLabel    = "label"
	VolumeFilterDangling = "dangling"

	NetworkFilterID     = "id"
	NetworkFilterName   = "name"
	NetworkFilterDriver = "driver"
	NetworkFilterLabel  = "label"
	NetworkFilterScope  = "scope"
	NetworkFilterType   = "type"
)

// VolumeListOptions holds the filter set for volume list requests.
type VolumeListOptions struct{ Filters FilterSet }

// NetworkListOptions holds the filter set for network list requests.
type NetworkListOptions struct{ Filters FilterSet }

// VolumeCreateOptions contains the parameters for creating a volume.
type VolumeCreateOptions struct {
	Name    string
	Driver  string
	Labels  map[string]string
	Options map[string]string
}

// NetworkCreateOptions contains the parameters for creating a network.
type NetworkCreateOptions struct {
	Name       string
	Driver     string
	Internal   bool
	EnableIPv6 bool
	Labels     map[string]string
	Options    map[string]string
}

// PruneOptions holds filters for resource pruning operations.
type PruneOptions struct{ Filters FilterSet }

// ResourceResult records the outcome of a single resource in a prune operation.
type ResourceResult struct {
	ID    string
	Error error
}

// PruneResult aggregates the results of a prune operation across resources.
type PruneResult struct {
	Resources      []ResourceResult
	SpaceReclaimed uint64
}

// Counts returns the number of successfully and unsuccessfully pruned resources.
func (r PruneResult) Counts() (succeeded, failed int) {
	for _, resource := range r.Resources {
		if resource.Error != nil {
			failed++
		} else {
			succeeded++
		}
	}
	return succeeded, failed
}

// NativeFilters converts the volume filter set to the native Docker/Podman
// representation. Adapters post-filter all values using AND semantics.
func (o VolumeListOptions) NativeFilters() (map[string][]string, error) {
	return nativeSingleValueFilters("volume.list.filter", o.Filters, map[string]struct{}{
		VolumeFilterName: {}, VolumeFilterDriver: {}, VolumeFilterLabel: {}, VolumeFilterDangling: {},
	})
}

// NativeFilters converts the network filter set to the native Docker/Podman
// representation. Adapters post-filter all values using AND semantics.
func (o NetworkListOptions) NativeFilters() (map[string][]string, error) {
	return nativeSingleValueFilters("network.list.filter", o.Filters, map[string]struct{}{
		NetworkFilterID: {}, NetworkFilterName: {}, NetworkFilterDriver: {},
		NetworkFilterLabel: {}, NetworkFilterScope: {}, NetworkFilterType: {},
	})
}

func nativeSingleValueFilters(operation string, input FilterSet, allowed map[string]struct{}) (map[string][]string, error) {
	filters := make(map[string][]string, len(input))
	for field, values := range input {
		if _, ok := allowed[field]; !ok {
			return nil, NewError(ErrorInvalid, operation, field, fmt.Errorf("unknown filter"))
		}
		if len(values) > 0 {
			filters[field] = []string{values[0]}
		}
	}
	return filters, nil
}

// Volume is the runtime-neutral summary of a Docker or Podman volume.
type Volume struct {
	Name       string
	Driver     string
	Mountpoint string
	Labels     map[string]string
	Scope      string
	CreatedAt  time.Time
}

// Network is the runtime-neutral summary of a Docker or Podman network.
type Network struct {
	Name       string
	ID         string
	Driver     string
	Scope      string
	IPAM       []string
	Containers int
	Created    int64
	Internal   bool
	Labels     map[string]string
}

// VolumeService provides lifecycle operations for volumes.
type VolumeService interface {
	List(context.Context, VolumeListOptions) ([]Volume, error)
	Inspect(context.Context, string) (*VolumeDetail, error)
	Create(context.Context, VolumeCreateOptions) (*Volume, error)
	Remove(context.Context, string, bool) error
	Prune(context.Context, PruneOptions) (PruneResult, error)
}

// NetworkService provides lifecycle operations for networks.
type NetworkService interface {
	List(context.Context, NetworkListOptions) ([]Network, error)
	Inspect(context.Context, string) (*NetworkDetail, error)
	Create(context.Context, NetworkCreateOptions) (*Network, error)
	Remove(context.Context, string) error
	Prune(context.Context, PruneOptions) (PruneResult, error)
}
