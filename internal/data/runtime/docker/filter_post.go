package docker

import (
	"strconv"
	"strings"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

// PostFilterContainers filters runtime-neutral container summaries by the
// runtime filter contract. Field semantics are kept aligned with the Docker
// SDK's /containers/json response.
func PostFilterContainers(items []runtimeapi.ContainerSummary, filters runtimeapi.FilterSet) ([]runtimeapi.ContainerSummary, error) {
	return runtimeapi.FilterSlice(items, func(item runtimeapi.ContainerSummary) (bool, error) {
		for field, values := range filters {
			for _, value := range values[1:] {
				var matched bool
				switch field {
				case runtimeapi.ContainerFilterID:
					matched = strings.HasPrefix(item.ID, value)
				case runtimeapi.ContainerFilterName:
					matched = strings.Contains(item.Name, value)
				case runtimeapi.ContainerFilterLabel:
					matched = runtimeapi.MatchesLabel(item.Labels, value)
				case runtimeapi.ContainerFilterStatus:
					matched = item.State == value
				case runtimeapi.ContainerFilterAncestor:
					matched = strings.Contains(item.Image, value)
				case runtimeapi.ContainerFilterNetwork:
					matched = runtimeapi.ContainsString(item.NetworkNames, value)
				case runtimeapi.ContainerFilterVolume:
					matched = runtimeapi.ContainsString(item.MountNames, value)
				}
				if !matched {
					return false, nil
				}
			}
		}
		return true, nil
	})
}

// PostFilterImages filters runtime-neutral image summaries by the runtime
// filter contract. Same-field AND is enforced; before/since/until values are
// not supported in client-side filtering.
func PostFilterImages(items []runtimeapi.ImageSummary, filters runtimeapi.FilterSet) ([]runtimeapi.ImageSummary, error) {
	return runtimeapi.FilterSlice(items, func(item runtimeapi.ImageSummary) (bool, error) {
		for field, values := range filters {
			for _, value := range values[1:] {
				var matched bool
				switch field {
				case runtimeapi.ImageFilterReference:
					matched = runtimeapi.ContainsReference(item.RepoTags, value)
				case runtimeapi.ImageFilterLabel:
					matched = runtimeapi.MatchesLabel(item.Labels, value)
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

// PostFilterVolumes filters runtime-neutral volume records by the runtime
// filter contract. Dangling detection is server-side only.
func PostFilterVolumes(items []runtimeapi.Volume, filters runtimeapi.FilterSet) ([]runtimeapi.Volume, error) {
	return runtimeapi.FilterSlice(items, func(item runtimeapi.Volume) (bool, error) {
		for field, values := range filters {
			for _, value := range values[1:] {
				matched := false
				switch field {
				case runtimeapi.VolumeFilterName:
					matched = strings.Contains(item.Name, value)
				case runtimeapi.VolumeFilterDriver:
					matched = item.Driver == value
				case runtimeapi.VolumeFilterLabel:
					matched = runtimeapi.MatchesLabel(item.Labels, value)
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

// PostFilterNetworks filters runtime-neutral network records by the runtime
// filter contract. Type filtering is not supported in client-side filtering.
func PostFilterNetworks(items []runtimeapi.Network, filters runtimeapi.FilterSet) ([]runtimeapi.Network, error) {
	return runtimeapi.FilterSlice(items, func(item runtimeapi.Network) (bool, error) {
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
					matched = runtimeapi.MatchesLabel(item.Labels, value)
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

func isDanglingImage(item runtimeapi.ImageSummary) bool {
	return runtimeapi.IsDanglingImage(item.RepoTags)
}
