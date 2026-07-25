package podman

import (
	"context"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/driver/podman"
	"github.com/elizabevil/docker-tui/internal/driver/podman/dto"
)

// PodmanNetworkService implements NetworkService for the Podman adapter.
type PodmanNetworkService struct {
	Client *podman.Client
}

// List returns network summaries, applying client-side post-filtering for
// multi-value filter conditions that the Podman API does not support natively.
func (s PodmanNetworkService) List(ctx context.Context, options runtimeapi.NetworkListOptions) ([]runtimeapi.Network, error) {
	filters := map[string][]string(options.Filters)
	raw, err := s.Client.REST.ListNetworks(ctx, dto.NetworkListOptions{Filters: filters})
	if err == nil {
		raw, err = PostFilterNetworks(raw, options.Filters)
	}
	if err != nil {
		return nil, mapPodmanNetworkErr(err, "list", "")
	}
	return MapNetworks(raw), nil
}

// Inspect returns the full detail view of a network by its ID or name.
func (s PodmanNetworkService) Inspect(ctx context.Context, id string) (*runtimeapi.NetworkDetail, error) {
	raw, err := s.Client.REST.InspectNetwork(ctx, id)
	if err != nil {
		return nil, mapPodmanNetworkErr(err, "inspect", id)
	}
	return MapNetworkInspect(*raw), nil
}

// Create provisions a new Podman network with the given options.
func (s PodmanNetworkService) Create(ctx context.Context, options runtimeapi.NetworkCreateOptions) (*runtimeapi.Network, error) {
	createOpts := dto.Network{
		Name:        options.Name,
		Driver:      options.Driver,
		Internal:    options.Internal,
		IPv6Enabled: options.EnableIPv6,
		Labels:      options.Labels,
		Options:     options.Options,
	}
	raw, err := s.Client.REST.CreateNetwork(ctx, createOpts)
	if err != nil {
		return nil, mapPodmanNetworkErr(err, "create", options.Name)
	}
	networks := MapNetworks([]podman.Network{*raw})
	if len(networks) == 0 {
		return nil, nil
	}
	return &networks[0], nil
}

// Remove deletes a Podman network by its ID or name.
func (s PodmanNetworkService) Remove(ctx context.Context, id string) error {
	err := s.Client.REST.RemoveNetwork(ctx, id)
	return mapPodmanNetworkErr(err, "remove", id)
}

// Prune removes unused Podman networks, returning counts of removed items.
func (s PodmanNetworkService) Prune(ctx context.Context, options runtimeapi.PruneOptions) (runtimeapi.PruneResult, error) {
	filters := map[string][]string(options.Filters)
	reports, err := s.Client.REST.PruneNetworks(ctx, filters)
	if err != nil {
		return runtimeapi.PruneResult{}, mapPodmanNetworkErr(err, "prune", "")
	}
	return assembleNetworkPruneResult(reports), nil
}

// mapPodmanNetworkErr wraps a Podman REST error with the network resource ref.
// An empty id builds an untyped ref (used for list/prune operations).
func mapPodmanNetworkErr(err error, op, id string) error {
	if err == nil {
		return nil
	}
	ref := runtimeapi.ResourceRef{Type: runtimeapi.ResourceNetwork}
	if id != "" {
		ref.ID = id
	}
	return runtimeapi.MapRuntimeError(err,
		runtimeapi.Operation(runtimeapi.ResourceNetwork, op), ref, runtimeapi.Podman)
}
