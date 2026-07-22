package docker

import (
	"context"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

// Volumes returns the volume service facade.
func (c *Client) Volumes() runtimeapi.VolumeService { return volumeService{client: c} }

type volumeService struct{ client *Client }

func (s volumeService) List(ctx context.Context, options runtimeapi.VolumeListOptions) ([]runtimeapi.Volume, error) {
	items, err := s.client.ListVolumesContext(ctx, options)
	if err == nil {
		items, err = postFilterVolumes(items, options.Filters)
	}
	return items, mapRuntimeError(err, "volume.list", runtimeapi.ResourceRef{Type: runtimeapi.ResourceVolume}, s.client.RuntimeType)
}

func (s volumeService) Inspect(ctx context.Context, name string) (*runtimeapi.VolumeDetail, error) {
	detail, err := s.client.InspectVolumeContext(ctx, name)
	return detail, mapRuntimeError(err, "volume.inspect", runtimeapi.ResourceRef{Type: runtimeapi.ResourceVolume, ID: name}, s.client.RuntimeType)
}

func (s volumeService) Create(ctx context.Context, options runtimeapi.VolumeCreateOptions) (*runtimeapi.Volume, error) {
	return s.client.CreateVolumeContext(ctx, options)
}

func (s volumeService) Remove(ctx context.Context, name string, force bool) error {
	err := s.client.cli.VolumeRemove(ctx, name, force)
	return mapRuntimeError(err, "volume.remove", runtimeapi.ResourceRef{Type: runtimeapi.ResourceVolume, ID: name}, s.client.RuntimeType)
}

func (s volumeService) Prune(ctx context.Context, options runtimeapi.PruneOptions) (runtimeapi.PruneResult, error) {
	return s.client.PruneVolumesContext(ctx, options)
}
