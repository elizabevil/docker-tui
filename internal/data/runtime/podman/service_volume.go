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

// List returns volume summaries, applying client-side post-filtering for
// multi-value filter conditions that the Podman API does not support natively.
func (s PodmanVolumeService) List(ctx context.Context, options runtimeapi.VolumeListOptions) ([]runtimeapi.Volume, error) {
	filters := map[string][]string(options.Filters)
	raw, err := s.Client.REST.ListVolumes(ctx, dto.VolumeListOptions{Filters: filters})
	if err == nil {
		raw, err = PostFilterVolumes(raw, options.Filters)
	}
	if err != nil {
		return nil, mapPodmanVolumeErr(err, "list", "")
	}
	return MapVolumes(raw), nil
}

// Inspect returns the full detail view of a volume by its name.
func (s PodmanVolumeService) Inspect(ctx context.Context, name string) (*runtimeapi.VolumeDetail, error) {
	raw, err := s.Client.REST.InspectVolume(ctx, name)
	if err != nil {
		return nil, mapPodmanVolumeErr(err, "inspect", name)
	}
	return MapVolumeInspect(*raw), nil
}

// Create provisions a new Podman volume with the given options.
func (s PodmanVolumeService) Create(ctx context.Context, options runtimeapi.VolumeCreateOptions) (*runtimeapi.Volume, error) {
	createOpts := dto.VolumeItem{
		Name:    options.Name,
		Driver:  options.Driver,
		Labels:  options.Labels,
		Options: options.Options,
	}
	raw, err := s.Client.REST.CreateVolume(ctx, createOpts)
	if err != nil {
		return nil, mapPodmanVolumeErr(err, "create", options.Name)
	}
	volumes := MapVolumes([]podman.VolumeItem{*raw})
	if len(volumes) == 0 {
		return nil, nil
	}
	return &volumes[0], nil
}

// Remove deletes a Podman volume by its name.
func (s PodmanVolumeService) Remove(ctx context.Context, name string, force bool) error {
	err := s.Client.REST.RemoveVolume(ctx, name, force)
	return mapPodmanVolumeErr(err, "remove", name)
}

// Prune removes unused Podman volumes, returning counts of removed items
// and reclaimed disk space.
func (s PodmanVolumeService) Prune(ctx context.Context, options runtimeapi.PruneOptions) (runtimeapi.PruneResult, error) {
	filters := map[string][]string(options.Filters)
	reports, err := s.Client.REST.PruneVolumes(ctx, filters)
	if err != nil {
		return runtimeapi.PruneResult{}, mapPodmanVolumeErr(err, "prune", "")
	}
	return assembleVolumePruneResult(reports), nil
}

// mapPodmanVolumeErr wraps a Podman REST error with the volume resource ref.
// An empty id builds an untyped ref (used for list/prune operations).
func mapPodmanVolumeErr(err error, op, id string) error {
	if err == nil {
		return nil
	}
	ref := runtimeapi.ResourceRef{Type: runtimeapi.ResourceVolume}
	if id != "" {
		ref.ID = id
	}
	return runtimeapi.MapRuntimeError(err,
		runtimeapi.Operation(runtimeapi.ResourceVolume, op), ref, runtimeapi.Podman)
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
