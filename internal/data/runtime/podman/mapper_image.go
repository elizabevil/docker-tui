package podman

import (
	"strings"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

// MapImageSummaries converts Podman image list results into
// runtime.ImageSummary items. Field differences: IsManifestList (*bool)
// is dereferenced; Arch defaults to "\u2014" when empty.
func MapImageSummaries(raw []ImageItem) []runtimeapi.ImageSummary {
	out := make([]runtimeapi.ImageSummary, 0, len(raw))
	for _, image := range raw {
		summary := runtimeapi.ImageSummary{
			ID:       image.ID,
			RepoTags: image.RepoTags,
			Created:  image.Created,
			Size:     image.Size,
			Labels:   image.Labels,
			Arch:     image.Arch,
		}
		if summary.Arch == "" {
			summary.Arch = "\u2014"
		}
		summary.IsManifest = image.IsManifestList != nil && *image.IsManifestList
		summary.Registry, _, _ = splitImageRef(summary.RepoTags)
		out = append(out, summary)
	}
	return out
}

// splitImageRef extracts the registry from the first RepoTag entry.
func splitImageRef(tags []string) (registry, image, tag string) {
	if len(tags) == 0 || tags[0] == "" {
		return "", "", ""
	}
	ref := tags[0]

	// Split tag
	lastColon := strings.LastIndex(ref, ":")
	if lastColon > 0 {
		tag = ref[lastColon+1:]
		ref = ref[:lastColon]
	}

	// Split registry from name
	parts := strings.Split(ref, "/")
	if len(parts) <= 1 {
		return "", ref, tag
	}
	if strings.Contains(parts[0], ".") || strings.Contains(parts[0], ":") || parts[0] == "localhost" {
		return parts[0], strings.Join(parts[1:], "/"), tag
	}
	return "", ref, tag
}
