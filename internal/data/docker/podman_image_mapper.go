package docker

// podmanImageSummary is the adapter's stable representation of the Libpod
// image-list response. Both transports normalize into this type before mapping.
type podmanImageSummary struct {
	ID             string            `json:"Id"`
	RepoTags       []string          `json:"RepoTags"`
	Created        int64             `json:"Created"`
	Size           int64             `json:"Size"`
	Labels         map[string]string `json:"Labels"`
	Architecture   string            `json:"Arch,omitempty"`
	IsManifestList *bool             `json:"IsManifestList,omitempty"`
}

// mapPodmanImageSummaries converts Podman image list results into ImageSummary
// items. Field differences: Architecture is derived from the Arch field;
// IsManifestList is a *bool that defaults to false when absent.
func mapPodmanImageSummaries(raw []podmanImageSummary) []ImageSummary {
	out := make([]ImageSummary, 0, len(raw))
	for _, image := range raw {
		summary := ImageSummary{
			ID:       image.ID,
			RepoTags: image.RepoTags,
			Created:  image.Created,
			Size:     image.Size,
			Labels:   image.Labels,
			Arch:     image.Architecture,
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
