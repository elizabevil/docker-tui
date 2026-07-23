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
		items, err = PostFilterImages(items, options.Filters)
	}
	return items, runtimeapi.MapRuntimeError(err, runtimeapi.Operation(runtimeapi.ResourceImage, "list"), runtimeapi.ResourceRef{Type: runtimeapi.ResourceImage}, runtimeapi.Docker)
}

func (s imageService) Inspect(ctx context.Context, summary runtimeapi.ImageSummary) (*runtimeapi.ImageDetail, error) {
	detail, err := s.client.InspectImageDetailContext(ctx, summary)
	return detail, runtimeapi.MapRuntimeError(err, runtimeapi.Operation(runtimeapi.ResourceImage, "inspect"), runtimeapi.ResourceRef{Type: runtimeapi.ResourceImage, ID: summary.ID}, runtimeapi.Docker)
}
