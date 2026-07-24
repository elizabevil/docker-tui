package podman

import (
	"strconv"
	"strings"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/driver/podman/dto"
)

// PostFilterContainers filters Podman container DTOs by the same runtime
// filter contract used by the Docker adapter. Field semantics are kept
// aligned with runtimeapi.ContainerSummary.
func PostFilterContainers(items []dto.ContainerItem, filters runtimeapi.FilterSet) ([]dto.ContainerItem, error) {
	return runtimeapi.FilterSlice(items, func(item dto.ContainerItem) (bool, error) {
		for field, values := range filters {
			for _, value := range values[1:] {
				var matched bool
				switch field {
				case runtimeapi.ContainerFilterID:
					matched = strings.HasPrefix(item.ID, value)
				case runtimeapi.ContainerFilterName:
					name := ""
					if len(item.Names) > 0 {
						name = item.Names[0]
					}
					matched = strings.Contains(name, value)
				case runtimeapi.ContainerFilterLabel:
					matched = matchesLabel(item.Labels, value)
				case runtimeapi.ContainerFilterStatus:
					matched = item.State == value
				case runtimeapi.ContainerFilterAncestor:
					matched = strings.Contains(item.Image, value)
				case runtimeapi.ContainerFilterNetwork:
					matched = containsString(item.Networks, value)
				case runtimeapi.ContainerFilterVolume:
					matched = containsString(item.Mounts, value)
				}
				if !matched {
					return false, nil
				}
			}
		}
		return true, nil
	})
}

// PostFilterImages filters Podman image DTOs by the same runtime filter
// contract used by the Docker adapter.
func PostFilterImages(items []dto.ImageItem, filters runtimeapi.FilterSet) ([]dto.ImageItem, error) {
	return runtimeapi.FilterSlice(items, func(item dto.ImageItem) (bool, error) {
		for field, values := range filters {
			for _, value := range values[1:] {
				var matched bool
				switch field {
				case runtimeapi.ImageFilterReference:
					matched = containsReference(item.RepoTags, value)
				case runtimeapi.ImageFilterLabel:
					matched = matchesLabel(item.Labels, value)
				case runtimeapi.ImageFilterDangling:
					want, err := strconv.ParseBool(value)
					if err != nil {
						return false, runtimeapi.NewError(runtimeapi.ErrorInvalid, "image.list.filter", value, err)
					}
					matched = isDanglingImage(item) == want
				case runtimeapi.ImageFilterBefore, runtimeapi.ImageFilterSince, runtimeapi.ImageFilterUntil:
					return false, runtimeapi.UnsupportedError(runtimeapi.Operation(runtimeapi.ResourceImage, "list.filter") + "." + field + ".and")
				}
				if !matched {
					return false, nil
				}
			}
		}
		return true, nil
	})
}

// PostFilterVolumes filters Podman volume DTOs by the same runtime filter
// contract used by the Docker adapter.
func PostFilterVolumes(items []dto.VolumeItem, filters runtimeapi.FilterSet) ([]dto.VolumeItem, error) {
	return runtimeapi.FilterSlice(items, func(item dto.VolumeItem) (bool, error) {
		for field, values := range filters {
			for _, value := range values[1:] {
				matched := false
				switch field {
				case runtimeapi.VolumeFilterName:
					matched = strings.Contains(item.Name, value)
				case runtimeapi.VolumeFilterDriver:
					matched = item.Driver == value
				case runtimeapi.VolumeFilterLabel:
					matched = matchesLabel(item.Labels, value)
				case runtimeapi.VolumeFilterDangling:
					return false, runtimeapi.UnsupportedError(runtimeapi.Operation(runtimeapi.ResourceVolume, "list.filter.dangling.and"))
				}
				if !matched {
					return false, nil
				}
			}
		}
		return true, nil
	})
}

// PostFilterNetworks filters Podman network DTOs by the same runtime
// filter contract used by the Docker adapter.
func PostFilterNetworks(items []dto.Network, filters runtimeapi.FilterSet) ([]dto.Network, error) {
	return runtimeapi.FilterSlice(items, func(item dto.Network) (bool, error) {
		for field, values := range filters {
			for _, value := range values[1:] {
				matched := false
				switch field {
				case runtimeapi.NetworkFilterID:
					matched = strings.HasPrefix(item.ID, value)
				case runtimeapi.NetworkFilterName:
					matched = strings.Contains(item.Name, value)
				case runtimeapi.NetworkFilterDriver:
					matched = item.Driver == value
				case runtimeapi.NetworkFilterLabel:
					matched = matchesLabel(item.Labels, value)
				case runtimeapi.NetworkFilterScope:
					matched = item.Scope == value
				case runtimeapi.NetworkFilterType:
					return false, runtimeapi.UnsupportedError(runtimeapi.Operation(runtimeapi.ResourceNetwork, "list.filter.type.and"))
				}
				if !matched {
					return false, nil
				}
			}
		}
		return true, nil
	})
}

func matchesLabel(labels map[string]string, expression string) bool {
	return runtimeapi.MatchesLabel(labels, expression)
}

func containsString(values []string, expected string) bool {
	return runtimeapi.ContainsString(values, expected)
}

func containsReference(references []string, expected string) bool {
	return runtimeapi.ContainsReference(references, expected)
}

func isDanglingImage(item dto.ImageItem) bool {
	return runtimeapi.IsDanglingImage(item.RepoTags)
}
