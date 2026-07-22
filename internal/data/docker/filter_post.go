package docker

import (
	"strconv"
	"strings"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

func postFilterContainers(items []runtimeapi.ContainerSummary, filters runtimeapi.FilterSet) ([]runtimeapi.ContainerSummary, error) {
	return filterSlice(items, func(item runtimeapi.ContainerSummary) (bool, error) {
		for field, values := range filters {
			for _, value := range values[1:] {
				var matched bool
				switch field {
				case runtimeapi.ContainerFilterID:
					matched = strings.HasPrefix(item.ID, value)
				case runtimeapi.ContainerFilterName:
					matched = strings.Contains(item.Name, value)
				case runtimeapi.ContainerFilterLabel:
					matched = matchesLabel(item.Labels, value)
				case runtimeapi.ContainerFilterStatus:
					matched = item.State == value
				case runtimeapi.ContainerFilterAncestor:
					matched = strings.Contains(item.Image, value)
				case runtimeapi.ContainerFilterNetwork:
					matched = containsString(item.NetworkNames, value)
				case runtimeapi.ContainerFilterVolume:
					matched = containsString(item.MountNames, value)
				}
				if !matched {
					return false, nil
				}
			}
		}
		return true, nil
	})
}

func postFilterImages(items []runtimeapi.ImageSummary, filters runtimeapi.FilterSet) ([]runtimeapi.ImageSummary, error) {
	return filterSlice(items, func(item runtimeapi.ImageSummary) (bool, error) {
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

func postFilterVolumes(items []runtimeapi.Volume, filters runtimeapi.FilterSet) ([]runtimeapi.Volume, error) {
	return filterSlice(items, func(item runtimeapi.Volume) (bool, error) {
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

func postFilterNetworks(items []runtimeapi.Network, filters runtimeapi.FilterSet) ([]runtimeapi.Network, error) {
	return filterSlice(items, func(item runtimeapi.Network) (bool, error) {
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

func isDanglingImage(item runtimeapi.ImageSummary) bool {
	return len(item.RepoTags) == 0 || (len(item.RepoTags) == 1 && item.RepoTags[0] == "<none>:<none>")
}
