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

func (s PodmanNetworkService) List(ctx context.Context, options runtimeapi.NetworkListOptions) ([]runtimeapi.Network, error) {
	filters := map[string][]string(options.Filters)
	raw, err := s.Client.REST.ListNetworks(ctx, dto.NetworkListOptions{Filters: filters})
	if err == nil {
		raw, err = PostFilterNetworks(raw, options.Filters)
	}
	if err != nil {
		return nil, runtimeapi.MapRuntimeError(err, runtimeapi.Operation(runtimeapi.ResourceNetwork, "list"), runtimeapi.ResourceRef{Type: runtimeapi.ResourceNetwork}, runtimeapi.Podman)
	}
	return MapNetworks(raw), nil
}

func (s PodmanNetworkService) Inspect(ctx context.Context, id string) (*runtimeapi.NetworkDetail, error) {
	raw, err := s.Client.REST.InspectNetwork(ctx, id)
	if err != nil {
		return nil, runtimeapi.MapRuntimeError(err, runtimeapi.Operation(runtimeapi.ResourceNetwork, "inspect"), runtimeapi.ResourceRef{Type: runtimeapi.ResourceNetwork, ID: id}, runtimeapi.Podman)
	}
	return MapNetworkInspect(*raw), nil
}

func (s PodmanNetworkService) Create(ctx context.Context, options runtimeapi.NetworkCreateOptions) (*runtimeapi.Network, error) {
	raw, err := s.Client.REST.CreateNetwork(ctx, dto.Network{Name: options.Name, Driver: options.Driver, Internal: options.Internal, IPv6Enabled: options.EnableIPv6, Labels: options.Labels, Options: options.Options})
	if err != nil {
		return nil, runtimeapi.MapRuntimeError(err, runtimeapi.Operation(runtimeapi.ResourceNetwork, "create"), runtimeapi.ResourceRef{Type: runtimeapi.ResourceNetwork, ID: options.Name}, runtimeapi.Podman)
	}
	networks := MapNetworks([]podman.Network{*raw})
	if len(networks) == 0 {
		return nil, nil
	}
	return &networks[0], nil
}

func (s PodmanNetworkService) Remove(ctx context.Context, id string) error {
	err := s.Client.REST.RemoveNetwork(ctx, id)
	return runtimeapi.MapRuntimeError(err, runtimeapi.Operation(runtimeapi.ResourceNetwork, "remove"), runtimeapi.ResourceRef{Type: runtimeapi.ResourceNetwork, ID: id}, runtimeapi.Podman)
}

func (s PodmanNetworkService) Prune(ctx context.Context, options runtimeapi.PruneOptions) (runtimeapi.PruneResult, error) {
	filters := map[string][]string(options.Filters)
	reports, err := s.Client.REST.PruneNetworks(ctx, filters)
	if err != nil {
		return runtimeapi.PruneResult{}, runtimeapi.MapRuntimeError(err, runtimeapi.Operation(runtimeapi.ResourceNetwork, "prune"), runtimeapi.ResourceRef{Type: runtimeapi.ResourceNetwork}, runtimeapi.Podman)
	}
	return assembleNetworkPruneResult(reports), nil
}
