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
	items, err := s.client.listContainersDocker(ctx, queryOptions)
	if err == nil {
		items, err = PostFilterContainers(items, options.Filters)
		if options.Limit > 0 && len(items) > options.Limit {
			items = items[:options.Limit]
		}
	}
	return items, mapContainerErrWithID(err, "list", "")
}

func (s containerService) Inspect(ctx context.Context, id string) (*runtimeapi.ContainerDetail, error) {
	detail, err := s.client.inspectContainerContext(ctx, id)
	return detail, mapContainerErrWithID(err, "inspect", id)
}

func (s containerService) Top(ctx context.Context, id string) (runtimeapi.ContainerProcesses, error) {
	processes, err := s.client.containerTopContext(ctx, id)
	return processes, mapContainerErrWithID(err, "top", id)
}

func (s containerService) Stats(ctx context.Context, id string) (runtimeapi.ContainerStats, error) {
	stats, err := s.client.containerStatsContext(ctx, id)
	return stats, mapContainerErrWithID(err, "stats", id)
}

func (s containerService) Logs(ctx context.Context, id string, options runtimeapi.ContainerLogOptions) (io.ReadCloser, error) {
	reader, err := s.client.containerLogsContext(ctx, id, options)
	return reader, mapContainerErrWithID(err, "logs", id)
}

// mapContainerErrWithID wraps a Docker SDK error with the container
// resource ref. An empty id builds an untyped ref (used for list operations).
func mapContainerErrWithID(err error, op, id string) error {
	if err == nil {
		return nil
	}
	ref := runtimeapi.ResourceRef{Type: runtimeapi.ResourceContainer}
	if id != "" {
		ref.ID = id
	}
	return runtimeapi.MapRuntimeError(err, runtimeapi.Operation(runtimeapi.ResourceContainer, op), ref, runtimeapi.Docker)
}

// TASK-019 advanced container operations. These delegate to the
// matching Client methods in container_advanced.go.

// Update applies resource-limit changes to a running container.
func (s containerService) Update(ctx context.Context, id string, options runtimeapi.ContainerUpdateOptions) (runtimeapi.ContainerUpdateResult, error) {
	return s.client.ContainerUpdate(ctx, id, options)
}

// Diff returns filesystem changes between a container and its base image.
func (s containerService) Diff(ctx context.Context, id string) ([]runtimeapi.ContainerDiffChange, error) {
	return s.client.ContainerDiff(ctx, id)
}

// Export returns a tar stream of the container's filesystem. Caller must close.
func (s containerService) Export(ctx context.Context, id string) (io.ReadCloser, error) {
	return s.client.ContainerExport(ctx, id)
}

// Commit snapshots a container's filesystem changes as a new image.
func (s containerService) Commit(ctx context.Context, id string, options runtimeapi.ContainerCommitOptions) (runtimeapi.ContainerCommitResult, error) {
	return s.client.ContainerCommit(ctx, id, options)
}

// Wait blocks until the container exits or the condition is met.
func (s containerService) Wait(ctx context.Context, id, condition string) (runtimeapi.ContainerWaitResult, error) {
	return s.client.ContainerWait(ctx, id, condition)
}

// CopyFromContainer streams a single file or directory out of a container.
func (s containerService) CopyFromContainer(ctx context.Context, id, srcPath string) (io.ReadCloser, error) {
	return s.client.ContainerCopyFrom(ctx, id, srcPath)
}
