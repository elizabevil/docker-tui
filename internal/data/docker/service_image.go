package docker

import (
	"context"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

// Images returns the image service facade.
func (c *Client) Images() runtimeapi.ImageService { return imageService{client: c} }

type imageService struct{ client *Client }

func (s imageService) List(ctx context.Context, options runtimeapi.ImageListOptions) ([]runtimeapi.ImageSummary, error) {
	items, err := s.client.ListImagesWithOptionsContext(ctx, options)
	if err == nil {
		items, err = postFilterImages(items, options.Filters)
	}
	return items, mapRuntimeError(err, "image.list", runtimeapi.ResourceRef{Type: runtimeapi.ResourceImage}, s.client.RuntimeType)
}

func (s imageService) Inspect(ctx context.Context, summary runtimeapi.ImageSummary) (*runtimeapi.ImageDetail, error) {
	detail, err := s.client.InspectImageDetailContext(ctx, summary)
	return detail, mapRuntimeError(err, "image.inspect", runtimeapi.ResourceRef{Type: runtimeapi.ResourceImage, ID: summary.ID}, s.client.RuntimeType)
}
