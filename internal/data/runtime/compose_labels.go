package runtime

import "strings"

// Compose label keys set by docker compose / podman compose on containers,
// volumes and networks. Centralised so adapters, mappers and the TUI share
// one source of truth instead of repeating string literals.
const (
	ComposeLabelProject         = "com.docker.compose.project"
	ComposeLabelService         = "com.docker.compose.service"
	ComposeLabelImage           = "com.docker.compose.image"
	ComposeLabelContainerNumber = "com.docker.compose.container-number"
	ComposeLabelConfigHash      = "com.docker.compose.config-hash"
	ComposeLabelOneoff          = "com.docker.compose.oneoff"
	ComposeLabelVersion         = "com.docker.compose.version"
	ComposeLabelWorkingDir      = "com.docker.compose.project.working_dir"
	ComposeLabelConfigFiles     = "com.docker.compose.project.config_files"
)

// ComposeProjectLabelValue returns the label filter value selecting
// resources that belong to the given compose project.
func ComposeProjectLabelValue(project string) string {
	return ComposeLabelProject + "=" + project
}

// ApplyComposeLabels fills the label-derived fields of a ContainerSummary
// from a container's labels. It is used by both docker and podman adapters
// to populate WorkingDir / ConfigFiles / Version / ConfigHash /
// ContainerNumber / Oneoff without an additional inspect call.
//
// The inspect-derived fields (NetworkMode, PidMode, IpcMode, CoLocatedGroupID,
// CoLocatedGroupSrc) are intentionally NOT filled here — they require a
// per-container inspect and are populated by the CoLocated group resolver
// (R08-14 §F2 + R08-15 §Q2) on demand, with 5s TTL cache.
func ApplyComposeLabels(labels map[string]string, summary *ContainerSummary) {
	summary.WorkingDir = labels[ComposeLabelWorkingDir]
	summary.Version = labels[ComposeLabelVersion]
	summary.ConfigHash = labels[ComposeLabelConfigHash]
	summary.ContainerNumber = labels[ComposeLabelContainerNumber]
	summary.Oneoff = labels[ComposeLabelOneoff] == "true"

	if raw := labels[ComposeLabelConfigFiles]; raw != "" {
		parts := strings.Split(raw, "\n")
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			if p != "" {
				out = append(out, p)
			}
		}
		summary.ConfigFiles = out
	}
}

// ComposeImageFromLabels returns the image reference declared by the
// compose file (com.docker.compose.image label) when present, falling
// back to the container's effective Image. Centralised so the keyboard
// layer's push / top / stats paths and the action bar's
// `compose_tagged` predicate share the same parser.
func ComposeImageFromLabels(c ContainerSummary) string {
	if c.Labels != nil {
		if img := c.Labels[ComposeLabelImage]; img != "" {
			return img
		}
	}
	return c.Image
}

// ComposeServiceFromContainer returns the canonical service label
// value for a container, falling back to "unknown" when the label is
// absent. Compose groups containers by (project, service); the
// "unknown" sentinel keeps aggregations stable for malformed labels.
func ComposeServiceFromContainer(c ContainerSummary) string {
	if c.ComposeService != "" {
		return c.ComposeService
	}
	return "unknown"
}

// ComposeProjectFromContainer returns the canonical project label
// value for a container. Empty string means "not part of any compose
// project" — callers should filter such rows out.
func ComposeProjectFromContainer(c ContainerSummary) string {
	return c.ComposeProject
}
