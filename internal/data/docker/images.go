package docker

import (
	"fmt"
	"io"
	"strings"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
)

var _ ImageLister = (*dockerImageLister)(nil) // compile-time check

func (c *Client) ListImages() ([]ImageSummary, error) {
	if c.imageLister != nil {
		return c.imageLister.ListImages()
	}
	return c.listImagesDocker()
}

// listImagesDocker uses the Docker SDK's ImageList (compat API).
func (c *Client) listImagesDocker() ([]ImageSummary, error) {
	images, err := c.cli.ImageList(c.ctx, image.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list images: %w", err)
	}
	result := make([]ImageSummary, 0, len(images))
	for _, img := range images {
		s := ImageSummary{
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
			entry := ImageManifestEntry{
				Digest:    m.ID,
				Size:      m.Size.Content,
				Available: m.Available,
			}
			if m.ImageData != nil && m.ImageData.Platform.OS != "" {
				entry.Platform = ManifestPlatform{
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
	return splitImageRef(tags)
}

func splitImageRef(tags []string) (registry, name, tag string) {
	if len(tags) == 0 || tags[0] == "" {
		return "—", "<none>", "<none>"
	}
	ref := tags[0]
	if ref == "<none>:<none>" {
		return "—", "<none>", "<none>"
	}

	// Split tag
	lastColon := strings.LastIndex(ref, ":")
	if lastColon > 0 {
		tag = ref[lastColon+1:]
		ref = ref[:lastColon]
	} else {
		tag = "latest"
	}

	// Split registry from name
	// docker.io/library/nginx → registry=docker.io, name=library/nginx
	// nginx → registry=docker.io, name=library/nginx
	parts := strings.Split(ref, "/")
	if len(parts) == 0 {
		return "—", ref, tag
	}
	if len(parts) == 1 {
		return "docker.io", "library/" + parts[0], tag
	}
	// Check if first part looks like a registry (contains . or :)
	if strings.Contains(parts[0], ".") || strings.Contains(parts[0], ":") || parts[0] == "localhost" {
		registry = parts[0]
		name = strings.Join(parts[1:], "/")
	} else {
		registry = "docker.io"
		name = strings.Join(parts, "/")
	}

	// Truncate long registry
	const maxReg = 13
	if len(registry) > maxReg {
		registry = registry[:maxReg] + "..."
	}

	return
}

func (c *Client) RemoveImage(id string, force bool) error {
	_, err := c.cli.ImageRemove(c.ctx, id, image.RemoveOptions{Force: force})
	return err
}

func (c *Client) PullImage(ref string) error {
	reader, err := c.cli.ImagePull(c.ctx, ref, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("pull image %s: %w", ref, err)
	}
	defer reader.Close()
	_, err = io.Copy(io.Discard, reader)
	return err
}

func (c *Client) PruneImages() (uint64, error) {
	report, err := c.cli.ImagesPrune(c.ctx, filters.NewArgs())
	if err != nil {
		return 0, fmt.Errorf("prune images: %w", err)
	}
	return report.SpaceReclaimed, nil
}

func (c *Client) InspectImage(id string) (string, error) {
	info, _, err := c.cli.ImageInspectWithRaw(c.ctx, id)
	if err != nil {
		return "", fmt.Errorf("inspect image %s: %w", id, err)
	}
	var sb strings.Builder

	fmt.Fprintf(&sb, "ID: %s\n", info.ID)
	if len(info.RepoTags) > 0 {
		fmt.Fprintf(&sb, "Tags: %s\n", strings.Join(info.RepoTags, ", "))
		reg, name, tag := SplitImageRef(info.RepoTags)
		fmt.Fprintf(&sb, "Registry: %s\n", reg)
		fmt.Fprintf(&sb, "Name: %s\n", name)
		fmt.Fprintf(&sb, "Tag: %s\n", tag)
	}
	if len(info.RepoDigests) > 0 {
		fmt.Fprintf(&sb, "Digest: %s\n", strings.Join(info.RepoDigests, ", "))
	}
	fmt.Fprintf(&sb, "Created: %s\n", info.Created)
	fmt.Fprintf(&sb, "Size: %d bytes\n", info.Size)

	fmt.Fprintf(&sb, "\n── System ──\n")
	fmt.Fprintf(&sb, "Architecture: %s\n", info.Architecture)
	fmt.Fprintf(&sb, "OS: %s\n", info.Os)
	if info.OsVersion != "" {
		fmt.Fprintf(&sb, "OS Version: %s\n", info.OsVersion)
	}

	if info.Author != "" {
		fmt.Fprintf(&sb, "Author: %s\n", info.Author)
	}
	if info.Comment != "" {
		fmt.Fprintf(&sb, "Comment: %s\n", info.Comment)
	}

	if info.Config != nil {
		fmt.Fprintf(&sb, "\n── Runtime ──\n")
		if info.Config.WorkingDir != "" {
			fmt.Fprintf(&sb, "WorkingDir: %s\n", info.Config.WorkingDir)
		}
		if info.Config.User != "" {
			fmt.Fprintf(&sb, "User: %s\n", info.Config.User)
		}
		if info.Config.StopSignal != "" {
			fmt.Fprintf(&sb, "StopSignal: %s\n", info.Config.StopSignal)
		}

		fmt.Fprintf(&sb, "\n── Entrypoint / Cmd ──\n")
		if len(info.Config.Entrypoint) > 0 {
			fmt.Fprintf(&sb, "Entrypoint: [%s]\n", strings.Join(info.Config.Entrypoint, ", "))
		}
		if len(info.Config.Cmd) > 0 {
			fmt.Fprintf(&sb, "Cmd: [%s]\n", strings.Join(info.Config.Cmd, ", "))
		}
		if len(info.Config.Shell) > 0 {
			fmt.Fprintf(&sb, "Shell: [%s]\n", strings.Join(info.Config.Shell, ", "))
		}
		if len(info.Config.OnBuild) > 0 {
			fmt.Fprintf(&sb, "OnBuild: [%s]\n", strings.Join(info.Config.OnBuild, ", "))
		}

		if len(info.Config.Env) > 0 {
			fmt.Fprintf(&sb, "\n── Environment (%d vars) ──\n", len(info.Config.Env))
			for i, env := range info.Config.Env {
				fmt.Fprintf(&sb, "  %d. %s\n", i+1, env)
			}
		}

		if len(info.Config.ExposedPorts) > 0 {
			ports := make([]string, 0, len(info.Config.ExposedPorts))
			for p := range info.Config.ExposedPorts {
				ports = append(ports, string(p))
			}
			fmt.Fprintf(&sb, "ExposedPorts: %s\n", strings.Join(ports, ", "))
		}

		if len(info.Config.Volumes) > 0 {
			fmt.Fprintf(&sb, "\n── Volumes ──\n")
			for v := range info.Config.Volumes {
				fmt.Fprintf(&sb, "  %s\n", v)
			}
		}

		if info.Config.Healthcheck != nil {
			fmt.Fprintf(&sb, "\n── Healthcheck ──\n")
			if len(info.Config.Healthcheck.Test) > 0 {
				fmt.Fprintf(&sb, "Test: [%s]\n", strings.Join(info.Config.Healthcheck.Test, ", "))
			}
			if info.Config.Healthcheck.Interval > 0 {
				fmt.Fprintf(&sb, "Interval: %s\n", info.Config.Healthcheck.Interval)
			}
			if info.Config.Healthcheck.Timeout > 0 {
				fmt.Fprintf(&sb, "Timeout: %s\n", info.Config.Healthcheck.Timeout)
			}
			if info.Config.Healthcheck.Retries > 0 {
				fmt.Fprintf(&sb, "Retries: %d\n", info.Config.Healthcheck.Retries)
			}
			if info.Config.Healthcheck.StartPeriod > 0 {
				fmt.Fprintf(&sb, "StartPeriod: %s\n", info.Config.Healthcheck.StartPeriod)
			}
		}
	}

	if info.GraphDriver.Name != "" {
		fmt.Fprintf(&sb, "\n── Storage ──\n")
		fmt.Fprintf(&sb, "Driver: %s\n", info.GraphDriver.Name)
		fmt.Fprintf(&sb, "Layers: %d\n", len(info.RootFS.Layers))
	}

	if len(info.ContainerConfig.Labels) > 0 || len(info.Config.Labels) > 0 {
		fmt.Fprintf(&sb, "\n── Labels ──\n")
		labels := info.Config.Labels
		if len(labels) == 0 {
			labels = info.ContainerConfig.Labels
		}
		for k, v := range labels {
			fmt.Fprintf(&sb, "  %s=%s\n", k, v)
		}
	}

	return sb.String(), nil
}
