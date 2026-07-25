package docker

import (
	"context"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

// Networks returns the network service facade.
func (c *Client) Networks() runtimeapi.NetworkService { return networkService{client: c} }

type networkService struct{ client *Client }

func (s networkService) List(ctx context.Context, options runtimeapi.NetworkListOptions) ([]runtimeapi.Network, error) {
	items, err := s.client.ListNetworksContext(ctx, options)
	if err == nil {
		items, err = PostFilterNetworks(items, options.Filters)
	}
	return items, runtimeapi.MapRuntimeError(err, runtimeapi.Operation(runtimeapi.ResourceNetwork, "list"), runtimeapi.ResourceRef{Type: runtimeapi.ResourceNetwork}, runtimeapi.Docker)
}

func (s networkService) Inspect(ctx context.Context, id string) (*runtimeapi.NetworkDetail, error) {
	detail, err := s.client.InspectNetworkContext(ctx, id)
	return detail, runtimeapi.MapRuntimeError(err, runtimeapi.Operation(runtimeapi.ResourceNetwork, "inspect"), runtimeapi.ResourceRef{Type: runtimeapi.ResourceNetwork, ID: id}, runtimeapi.Docker)
}

func (s networkService) Create(ctx context.Context, options runtimeapi.NetworkCreateOptions) (*runtimeapi.Network, error) {
	return s.client.CreateNetworkContext(ctx, options)
}

func (s networkService) Remove(ctx context.Context, id string) error {
	err := s.client.cli.NetworkRemove(ctx, id)
	return runtimeapi.MapRuntimeError(err, runtimeapi.Operation(runtimeapi.ResourceNetwork, "remove"), runtimeapi.ResourceRef{Type: runtimeapi.ResourceNetwork, ID: id}, runtimeapi.Docker)
}

func (s networkService) Prune(ctx context.Context, options runtimeapi.PruneOptions) (runtimeapi.PruneResult, error) {
	return s.client.PruneNetworksContext(ctx, options)
}
