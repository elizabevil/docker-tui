//go:build cgo

// This file compiles only when CGO is available. It uses the official Podman
// Go bindings (go.podman.io/podman/v6/pkg/bindings/images) which require CGO
// for local libpod communication. TLS connections and explicit API overrides
// use the shared REST implementation because bindings cannot express all of
// the configured transport semantics.

package docker

import (
	"context"
	"fmt"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"go.podman.io/podman/v6/pkg/bindings"
	"go.podman.io/podman/v6/pkg/bindings/images"
)

// listImagesPodman returns images from the Podman runtime via CGO bindings.
func (c *Client) listImagesPodman(ctx context.Context, options runtimeapi.ImageListOptions) ([]ImageSummary, error) {
	if c.usePodmanRESTTransport() {
		return c.listImagesPodmanREST(ctx, options)
	}
	bindingContext, err := bindings.NewConnection(ctx, c.Host)
	if err != nil {
		return nil, fmt.Errorf("podman connect: %w", err)
	}
	nativeFilters, err := options.NativeFilters()
	if err != nil {
		return nil, err
	}
	listOptions := new(images.ListOptions).WithAll(options.All).WithFilters(nativeFilters)
	list, err := images.List(bindingContext, listOptions)
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
