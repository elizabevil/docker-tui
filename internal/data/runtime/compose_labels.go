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
