package docker

import (
	"context"
	"io"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

// Containers returns the container service facade.
func (c *Client) Containers() runtimeapi.ContainerService {
	return containerService{client: c}
}

type containerService struct {
	client *Client
}

func (s containerService) List(ctx context.Context, options runtimeapi.ContainerListOptions) ([]runtimeapi.ContainerSummary, error) {
	queryOptions := options
	if options.Filters.HasMultipleValues() {
		queryOptions.Limit = 0
	}
	items, err := s.client.ListContainersContext(ctx, queryOptions)
	if err == nil {
		items, err = postFilterContainers(items, options.Filters)
		if options.Limit > 0 && len(items) > options.Limit {
			items = items[:options.Limit]
		}
	}
	return items, mapRuntimeError(err, "container.list", runtimeapi.ResourceRef{Type: runtimeapi.ResourceContainer}, s.client.RuntimeType)
}

func (s containerService) Inspect(ctx context.Context, id string) (*runtimeapi.ContainerDetail, error) {
	detail, err := s.client.inspectContainerContext(ctx, id)
	return detail, mapRuntimeError(err, "container.inspect", runtimeapi.ResourceRef{Type: runtimeapi.ResourceContainer, ID: id}, s.client.RuntimeType)
}

func (s containerService) Top(ctx context.Context, id string) (runtimeapi.ContainerProcesses, error) {
	processes, err := s.client.containerTopContext(ctx, id)
	return processes, mapRuntimeError(err, "container.top", runtimeapi.ResourceRef{Type: runtimeapi.ResourceContainer, ID: id}, s.client.RuntimeType)
}

func (s containerService) Stats(ctx context.Context, id string) (runtimeapi.ContainerStats, error) {
	stats, err := s.client.containerStatsContext(ctx, id)
	return stats, mapRuntimeError(err, "container.stats", runtimeapi.ResourceRef{Type: runtimeapi.ResourceContainer, ID: id}, s.client.RuntimeType)
}

func (s containerService) Logs(ctx context.Context, id string, options runtimeapi.ContainerLogOptions) (io.ReadCloser, error) {
	reader, err := s.client.containerLogsContext(ctx, id, options)
	return reader, mapRuntimeError(err, "container.logs", runtimeapi.ResourceRef{Type: runtimeapi.ResourceContainer, ID: id}, s.client.RuntimeType)
}
