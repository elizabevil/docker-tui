package docker

import (
	"context"

	"github.com/docker/docker/api/types/image"
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

func (s imageService) History(ctx context.Context, summary runtimeapi.ImageSummary) ([]runtimeapi.ImageHistoryLayer, error) {
	raw, err := s.client.cli.ImageHistory(ctx, summary.ID)
	if err != nil {
		return nil, runtimeapi.MapRuntimeError(err, runtimeapi.Operation(runtimeapi.ResourceImage, "history"), runtimeapi.ResourceRef{Type: runtimeapi.ResourceImage, ID: summary.ID}, runtimeapi.Docker)
	}
	return mapDockerHistory(raw), nil
}

// mapDockerHistory converts the Docker SDK's history response items into
// the runtime-neutral ImageHistoryLayer slice.
func mapDockerHistory(raw []image.HistoryResponseItem) []runtimeapi.ImageHistoryLayer {
	layers := make([]runtimeapi.ImageHistoryLayer, 0, len(raw))
	for _, h := range raw {
		layers = append(layers, runtimeapi.ImageHistoryLayer{
			ID:        h.ID,
			Created:   h.Created,
			CreatedBy: h.CreatedBy,
			Size:      h.Size,
			Comment:   h.Comment,
		})
	}
	return layers
}
