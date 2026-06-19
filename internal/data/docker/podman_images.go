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
	out := make([]ImageSummary, 0, len(list))
	for _, p := range list {
		s := ImageSummary{
			ID: p.ID, RepoTags: p.RepoTags, Created: p.Created,
			Size: p.Size, Labels: p.Labels,
		}
		if p.Arch != "" {
			s.Arch = p.Arch
		} else {
			s.Arch = "\u2014"
		}
		if p.IsManifestList != nil && *p.IsManifestList {
			s.IsManifest = true
		}
		s.Registry, _, _ = splitImageRef(s.RepoTags)
		out = append(out, s)
	}
	return out, nil
}
