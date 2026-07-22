package docker

import (
	"context"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

// ImageLister provides image listing with full metadata.
// Different runtimes (Docker SDK vs Podman SDK) implement this interface
// to provide runtime-specific metadata (architecture, manifest info, etc.).
type ImageLister interface {
	ListImages(context.Context, runtimeapi.ImageListOptions) ([]ImageSummary, error)
}

// dockerImageLister uses the Docker SDK to list images.
type dockerImageLister struct {
	client *Client
}

func (l *dockerImageLister) ListImages(ctx context.Context, options runtimeapi.ImageListOptions) ([]ImageSummary, error) {
	return l.client.listImagesDocker(ctx, options)
}
