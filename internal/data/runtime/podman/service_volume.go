package podman

import (
	"context"
	"fmt"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/driver/podman"
	"github.com/elizabevil/docker-tui/internal/driver/podman/dto"
)

// PodmanVolumeService implements VolumeService for the Podman adapter.
type PodmanVolumeService struct {
	Client *podman.Client
}

func (s PodmanVolumeService) List(ctx context.Context, options runtimeapi.VolumeListOptions) ([]runtimeapi.Volume, error) {
	filters := map[string][]string(options.Filters)
	raw, err := s.Client.REST.ListVolumes(ctx, dto.VolumeListOptions{Filters: filters})
	if err == nil {
		raw, err = PostFilterVolumes(raw, options.Filters)
	}
	if err != nil {
		return nil, runtimeapi.MapRuntimeError(err, runtimeapi.Operation(runtimeapi.ResourceVolume, "list"), runtimeapi.ResourceRef{Type: runtimeapi.ResourceVolume}, runtimeapi.Podman)
	}
	return MapVolumes(raw), nil
}

func (s PodmanVolumeService) Inspect(ctx context.Context, name string) (*runtimeapi.VolumeDetail, error) {
	raw, err := s.Client.REST.InspectVolume(ctx, name)
	if err != nil {
		return nil, runtimeapi.MapRuntimeError(err, runtimeapi.Operation(runtimeapi.ResourceVolume, "inspect"), runtimeapi.ResourceRef{Type: runtimeapi.ResourceVolume, ID: name}, runtimeapi.Podman)
	}
	return MapVolumeInspect(*raw), nil
}

func (s PodmanVolumeService) Create(ctx context.Context, options runtimeapi.VolumeCreateOptions) (*runtimeapi.Volume, error) {
	raw, err := s.Client.REST.CreateVolume(ctx, dto.VolumeItem{Name: options.Name, Driver: options.Driver, Labels: options.Labels, Options: options.Options})
	if err != nil {
		return nil, runtimeapi.MapRuntimeError(err, runtimeapi.Operation(runtimeapi.ResourceVolume, "create"), runtimeapi.ResourceRef{Type: runtimeapi.ResourceVolume, ID: options.Name}, runtimeapi.Podman)
	}
	volumes := MapVolumes([]podman.VolumeItem{*raw})
	if len(volumes) == 0 {
		return nil, nil
	}
	return &volumes[0], nil
}

func (s PodmanVolumeService) Remove(ctx context.Context, name string, force bool) error {
	err := s.Client.REST.RemoveVolume(ctx, name, force)
	return runtimeapi.MapRuntimeError(err, runtimeapi.Operation(runtimeapi.ResourceVolume, "remove"), runtimeapi.ResourceRef{Type: runtimeapi.ResourceVolume, ID: name}, runtimeapi.Podman)
}

func (s PodmanVolumeService) Prune(ctx context.Context, options runtimeapi.PruneOptions) (runtimeapi.PruneResult, error) {
	filters := map[string][]string(options.Filters)
	reports, err := s.Client.REST.PruneVolumes(ctx, filters)
	if err != nil {
		return runtimeapi.PruneResult{}, runtimeapi.MapRuntimeError(err, runtimeapi.Operation(runtimeapi.ResourceVolume, "prune"), runtimeapi.ResourceRef{Type: runtimeapi.ResourceVolume}, runtimeapi.Podman)
	}
	return assembleVolumePruneResult(reports), nil
}

// assembleVolumePruneResult converts raw Podman volume prune reports into a PruneResult.
func assembleVolumePruneResult(reports []dto.VolumePruneReportItem) runtimeapi.PruneResult {
	result := runtimeapi.PruneResult{}
	for _, report := range reports {
		if report.Err != "" {
			result.Resources = append(result.Resources, runtimeapi.ResourceResult{ID: report.ID, Error: fmt.Errorf("%s", report.Err)})
		} else {
			result.Resources = append(result.Resources, runtimeapi.ResourceResult{ID: report.ID})
			if report.Size > 0 {
				result.SpaceReclaimed += uint64(report.Size)
			}
		}
	}
	return result
}

// assembleNetworkPruneResult converts raw Podman network prune reports into a PruneResult.
func assembleNetworkPruneResult(reports []dto.NetworkPruneReportItem) runtimeapi.PruneResult {
	result := runtimeapi.PruneResult{}
	for _, report := range reports {
		if report.Error != "" {
			result.Resources = append(result.Resources, runtimeapi.ResourceResult{ID: report.Name, Error: fmt.Errorf("%s", report.Error)})
		} else {
			result.Resources = append(result.Resources, runtimeapi.ResourceResult{ID: report.Name})
		}
	}
	return result
}
