package runtime

import (
	"context"
	"fmt"
)

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

type VolumeListOptions struct{ Filters FilterSet }
type NetworkListOptions struct{ Filters FilterSet }

type VolumeCreateOptions struct {
	Name    string
	Driver  string
	Labels  map[string]string
	Options map[string]string
}

type NetworkCreateOptions struct {
	Name       string
	Driver     string
	Internal   bool
	EnableIPv6 bool
	Labels     map[string]string
	Options    map[string]string
}

type PruneOptions struct{ Filters FilterSet }

type ResourceResult struct {
	ID    string
	Error error
}

type PruneResult struct {
	Resources      []ResourceResult
	SpaceReclaimed uint64
}

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

func (o VolumeListOptions) NativeFilters() (map[string][]string, error) {
	return nativeSingleValueFilters("volume.list.filter", o.Filters, map[string]struct{}{
		VolumeFilterName: {}, VolumeFilterDriver: {}, VolumeFilterLabel: {}, VolumeFilterDangling: {},
	})
}

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
		if len(values) > 1 {
			return nil, UnsupportedError(operation + "." + field)
		}
		if len(values) == 1 {
			filters[field] = append([]string(nil), values...)
		}
	}
	return filters, nil
}

type Volume struct {
	Name       string
	Driver     string
	Mountpoint string
	Labels     map[string]string
	Scope      string
	CreatedAt  string
}

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

type VolumeService interface {
	List(context.Context, VolumeListOptions) ([]Volume, error)
	Create(context.Context, VolumeCreateOptions) (*Volume, error)
	Prune(context.Context, PruneOptions) (PruneResult, error)
}

type NetworkService interface {
	List(context.Context, NetworkListOptions) ([]Network, error)
	Create(context.Context, NetworkCreateOptions) (*Network, error)
	Prune(context.Context, PruneOptions) (PruneResult, error)
}
