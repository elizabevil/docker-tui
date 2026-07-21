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
}

type NetworkService interface {
	List(context.Context, NetworkListOptions) ([]Network, error)
}
