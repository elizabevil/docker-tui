package docker

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

// ListImagesWithOptionsContext returns images with a caller-provided context and filter options.
func (c *Client) ListImagesWithOptionsContext(ctx context.Context, options runtimeapi.ImageListOptions) ([]runtimeapi.ImageSummary, error) {
	return c.listImagesDocker(ctx, options)
}

// listImagesDocker uses the Docker SDK's ImageList (compat API).
func (c *Client) listImagesDocker(ctx context.Context, options runtimeapi.ImageListOptions) ([]runtimeapi.ImageSummary, error) {
	nativeFilters, err := options.NativeFilters()
	if err != nil {
		return nil, err
	}
	filterArgs := filters.NewArgs()
	for field, values := range nativeFilters {
		for _, value := range values {
			filterArgs.Add(field, value)
		}
	}
	images, err := c.cli.ImageList(ctx, image.ListOptions{All: options.All, Filters: filterArgs})
	if err != nil {
		return nil, fmt.Errorf("list images: %w", err)
	}
	result := make([]runtimeapi.ImageSummary, 0, len(images))
	for _, img := range images {
		s := runtimeapi.ImageSummary{
			ID:       img.ID,
			RepoTags: img.RepoTags,
			Created:  img.Created,
			Size:     img.Size,
			Labels:   img.Labels,
		}
		// Arch from descriptor (OCI Descriptor.Platform)
		if img.Descriptor != nil && img.Descriptor.Platform != nil {
			s.Arch = img.Descriptor.Platform.Architecture
		} else {
			s.Arch = "\u2014"
		}
		// Populate manifests from Docker API
		for _, m := range img.Manifests {
			entry := runtimeapi.ImageManifestEntry{
				Digest:    m.ID,
				Size:      m.Size.Content,
				Available: m.Available,
			}
			if m.ImageData != nil && m.ImageData.Platform.OS != "" {
				entry.Platform = runtimeapi.ManifestPlatform{
					OS:           m.ImageData.Platform.OS,
					Architecture: m.ImageData.Platform.Architecture,
					Variant:      m.ImageData.Platform.Variant,
				}
			}
			s.Manifests = append(s.Manifests, entry)
		}

		// Detect manifest list and resolve arch
		if img.Descriptor != nil {
			mt := img.Descriptor.MediaType
			s.IsManifest = mt == "application/vnd.docker.distribution.manifest.list.v2+json" ||
				mt == "application/vnd.oci.image.index.v1+json"
		}
		// Fallback: if multiple manifests → it's a manifest list
		if len(s.Manifests) > 1 && !s.IsManifest {
			s.IsManifest = true
		}
		// Arch fallback: use first manifest's platform if descriptor not available
		if s.Arch == "\u2014" && len(s.Manifests) > 0 && s.Manifests[0].Platform.Architecture != "" {
			s.Arch = s.Manifests[0].Platform.Architecture
		}
		s.Registry, _, _ = splitImageRef(s.RepoTags)
		result = append(result, s)
	}
	return result, nil
}

// SplitImageRef splits first RepoTag into registry, name, tag.
func SplitImageRef(tags []string) (registry, name, tag string) {
	return runtimeapi.SplitImageRef(tags)
}

func splitImageRef(tags []string) (registry, name, tag string) {
	return runtimeapi.SplitImageRef(tags)
}

// InspectImageDetailContext returns full image detail with a caller-provided context.
func (c *Client) InspectImageDetailContext(ctx context.Context, summary runtimeapi.ImageSummary) (*runtimeapi.ImageDetail, error) {
	detail := NewImageDetailData(summary)
	info, _, err := c.cli.ImageInspectWithRaw(ctx, summary.ID)
	if err != nil {
		if detail.IsManifest && len(detail.ManifestVariants) > 0 {
			detail.HistoryError = err.Error()
			return detail, nil
		}
		return nil, fmt.Errorf("inspect image %s: %w", summary.ID, err)
	}

	detail.ID = info.ID
	detail.RepoTags = append([]string(nil), info.RepoTags...)
	detail.RepoDigests = append([]string(nil), info.RepoDigests...)
	detail.Registry, detail.Name, detail.Tag = SplitImageRef(info.RepoTags)
	detail.Created = info.Created
	detail.Size = info.Size
	detail.Architecture = info.Architecture
	detail.OS = info.Os
	detail.OSVersion = info.OsVersion
	detail.Author = info.Author
	detail.Comment = info.Comment
	detail.Driver = info.GraphDriver.Name
	detail.LayerCount = len(info.RootFS.Layers)
	if info.Config != nil {
		detail.Runtime = runtimeapi.ImageRuntimeConfig{
			WorkingDir:  info.Config.WorkingDir,
			User:        info.Config.User,
			StopSignal:  info.Config.StopSignal,
			Entrypoint:  append([]string(nil), info.Config.Entrypoint...),
			Cmd:         append([]string(nil), info.Config.Cmd...),
			Shell:       append([]string(nil), info.Config.Shell...),
			OnBuild:     append([]string(nil), info.Config.OnBuild...),
			Environment: append([]string(nil), info.Config.Env...),
		}
		for port := range info.Config.ExposedPorts {
			detail.Runtime.ExposedPorts = append(detail.Runtime.ExposedPorts, string(port))
		}
		for volume := range info.Config.Volumes {
			detail.Runtime.Volumes = append(detail.Runtime.Volumes, volume)
		}
		sort.Strings(detail.Runtime.ExposedPorts)
		sort.Strings(detail.Runtime.Volumes)
		if health := info.Config.Healthcheck; health != nil {
			detail.Runtime.Healthcheck = append(detail.Runtime.Healthcheck, "Test: "+strings.Join(health.Test, " "))
			if health.Interval > 0 {
				detail.Runtime.Healthcheck = append(detail.Runtime.Healthcheck, "Interval: "+health.Interval.String())
			}
			if health.Timeout > 0 {
				detail.Runtime.Healthcheck = append(detail.Runtime.Healthcheck, "Timeout: "+health.Timeout.String())
			}
			if health.Retries > 0 {
				detail.Runtime.Healthcheck = append(detail.Runtime.Healthcheck, fmt.Sprintf("Retries: %d", health.Retries))
			}
		}
		detail.Labels = cloneStringMap(info.Config.Labels)
	}
	if len(detail.Labels) == 0 {
		detail.Labels = cloneStringMap(info.ContainerConfig.Labels)
	}

	if detail.IsManifest {
		detail.HistorySource = runtimeapi.ImageHistoryManifest
		return detail, nil
	}

	detail.HistorySource = runtimeapi.ImageHistoryLayerAPI
	history, historyErr := c.cli.ImageHistory(ctx, summary.ID)
	if historyErr != nil {
		detail.HistoryError = historyErr.Error()
	} else {
		for _, layer := range history {
			detail.History = append(detail.History, runtimeapi.ImageHistoryLayer{
				ID: layer.ID, Created: layer.Created, CreatedBy: layer.CreatedBy,
				Size: layer.Size, Comment: layer.Comment,
			})
		}
	}
	return detail, nil
}

// NewImageDetailData creates an ImageDetailData pre-populated from a summary.
func NewImageDetailData(summary runtimeapi.ImageSummary) *runtimeapi.ImageDetail {
	registry, name, tag := SplitImageRef(summary.RepoTags)
	return &runtimeapi.ImageDetail{
		ID: summary.ID, RepoTags: append([]string(nil), summary.RepoTags...),
		Registry: registry, Name: name, Tag: tag, Size: summary.Size,
		Architecture: summary.Arch, Labels: cloneStringMap(summary.Labels),
		IsManifest:       summary.IsManifest,
		ManifestVariants: append([]runtimeapi.ImageManifestEntry(nil), summary.Manifests...),
		HistorySource:    imageHistorySource(summary.IsManifest),
	}
}

func imageHistorySource(isManifest bool) runtimeapi.ImageHistorySource {
	if isManifest {
		return runtimeapi.ImageHistoryManifest
	}
	return runtimeapi.ImageHistoryPending
}

func cloneStringMap(source map[string]string) map[string]string {
	if len(source) == 0 {
		return nil
	}
	result := make(map[string]string, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}
