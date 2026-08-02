package podman

import (
	"context"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/driver/podman"
	"github.com/elizabevil/docker-tui/internal/driver/podman/dto"
)

// PodmanImageService implements ImageService for the Podman adapter.
type PodmanImageService struct {
	Client *podman.Client
}

// List returns image summaries, applying client-side post-filtering for
// multi-value filter conditions that the Podman API does not support natively.
func (s PodmanImageService) List(ctx context.Context, options runtimeapi.ImageListOptions) ([]runtimeapi.ImageSummary, error) {
	filters := map[string][]string(options.Filters)
	raw, err := s.Client.REST.ListImages(ctx, dto.ImageListOptions{All: options.All, Filters: filters})
	if err == nil {
		raw, err = PostFilterImages(raw, options.Filters)
	}
	if err != nil {
		return nil, runtimeapi.MapRuntimeError(err, runtimeapi.Operation(runtimeapi.ResourceImage, "list"), runtimeapi.ResourceRef{Type: runtimeapi.ResourceImage}, runtimeapi.Podman)
	}
	return MapImageSummaries(raw), nil
}

// Inspect fetches full image detail via /libpod/images/{id}/json.
// Podman 5.x rejects the unversioned inspect path with 404, so the REST
// client transparently retries /v4.0.0/libpod/images/{id}/json on miss.
func (s PodmanImageService) Inspect(ctx context.Context, summary runtimeapi.ImageSummary) (*runtimeapi.ImageDetail, error) {
	inspect, err := s.Client.REST.InspectImage(ctx, summary.ID)
	if err != nil {
		return nil, runtimeapi.MapRuntimeError(err, runtimeapi.Operation(runtimeapi.ResourceImage, "inspect"), runtimeapi.ResourceRef{Type: runtimeapi.ResourceImage, ID: summary.ID}, runtimeapi.Podman)
	}
	// History is intentionally excluded from detail loading. Restore it as a
	// separate, on-demand request when the dedicated history view is added.
	return MapImageInspect(*inspect, nil), nil
}

func (s PodmanImageService) History(ctx context.Context, summary runtimeapi.ImageSummary) ([]runtimeapi.ImageHistoryLayer, error) {
	raw, err := s.Client.REST.ImageHistory(ctx, summary.ID)
	if err != nil {
		return nil, runtimeapi.MapRuntimeError(err, runtimeapi.Operation(runtimeapi.ResourceImage, "history"), runtimeapi.ResourceRef{Type: runtimeapi.ResourceImage, ID: summary.ID}, runtimeapi.Podman)
	}
	return mapPodmanHistory(raw), nil
}

// mapPodmanHistory converts the Podman REST DTO history entries into
// the runtime-neutral ImageHistoryLayer slice.
func mapPodmanHistory(raw []dto.LayerHistoryEntry) []runtimeapi.ImageHistoryLayer {
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
