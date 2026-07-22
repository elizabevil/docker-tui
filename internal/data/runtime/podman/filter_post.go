package podman

import (
	"strconv"
	"strings"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

// PostFilterContainers applies same-field AND post-filtering to raw Podman
// container list results.
func PostFilterContainers(items []ContainerItem, filters runtimeapi.FilterSet) ([]ContainerItem, error) {
	return filterSlice(items, func(item ContainerItem) (bool, error) {
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

// PostFilterImages applies same-field AND post-filtering to raw Podman
// image list results.
func PostFilterImages(items []ImageItem, filters runtimeapi.FilterSet) ([]ImageItem, error) {
	return filterSlice(items, func(item ImageItem) (bool, error) {
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
					return false, runtimeapi.UnsupportedError("image.list.filter." + field + ".and")
				}
				if !matched {
					return false, nil
				}
			}
		}
		return true, nil
	})
}

// PostFilterVolumes applies same-field AND post-filtering to raw Podman
// volume list results.
func PostFilterVolumes(items []VolumeItem, filters runtimeapi.FilterSet) ([]VolumeItem, error) {
	return filterSlice(items, func(item VolumeItem) (bool, error) {
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
					return false, runtimeapi.UnsupportedError("volume.list.filter.dangling.and")
				}
				if !matched {
					return false, nil
				}
			}
		}
		return true, nil
	})
}

// PostFilterNetworks applies same-field AND post-filtering to raw Podman
// network list results.
func PostFilterNetworks(items []Network, filters runtimeapi.FilterSet) ([]Network, error) {
	return filterSlice(items, func(item Network) (bool, error) {
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
					return false, runtimeapi.UnsupportedError("network.list.filter.type.and")
				}
				if !matched {
					return false, nil
				}
			}
		}
		return true, nil
	})
}

func filterSlice[T any](items []T, match func(T) (bool, error)) ([]T, error) {
	filtered := make([]T, 0, len(items))
	for _, item := range items {
		matched, err := match(item)
		if err != nil {
			return nil, err
		}
		if matched {
			filtered = append(filtered, item)
		}
	}
	return filtered, nil
}

func matchesLabel(labels map[string]string, expression string) bool {
	key, expected, hasValue := strings.Cut(expression, "=")
	actual, ok := labels[key]
	return ok && (!hasValue || actual == expected)
}

func containsString(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func containsReference(references []string, expected string) bool {
	for _, reference := range references {
		if reference == expected || strings.Contains(reference, expected) {
			return true
		}
	}
	return false
}

func isDanglingImage(item ImageItem) bool {
	return len(item.RepoTags) == 0 || (len(item.RepoTags) == 1 && item.RepoTags[0] == "<none>:<none>")
}
