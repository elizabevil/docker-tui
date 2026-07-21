//go:build cgo

package docker

import (
	"fmt"

	"go.podman.io/podman/v6/pkg/bindings"
	"go.podman.io/podman/v6/pkg/bindings/images"
)

func (c *Client) listImagesPodman() ([]ImageSummary, error) {
	ctx, err := bindings.NewConnection(c.ctx, c.Host)
	if err != nil {
		return nil, fmt.Errorf("podman connect: %w", err)
	}
	list, err := images.List(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("podman list: %w", err)
	}
	raw := make([]podmanImageSummary, 0, len(list))
	for _, p := range list {
		raw = append(raw, podmanImageSummary{
			ID:             p.ID,
			RepoTags:       p.RepoTags,
			Created:        p.Created,
			Size:           p.Size,
			Labels:         p.Labels,
			Architecture:   p.Arch,
			IsManifestList: p.IsManifestList,
		})
	}
	return mapPodmanImageSummaries(raw), nil
}
