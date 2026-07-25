package docker

import (
	"context"
	"io"

	"github.com/docker/docker/api/types/container"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

// Update applies resource-limit changes to a running container. TASK-019.
func (c *Client) ContainerUpdate(ctx context.Context, id string, opts runtimeapi.ContainerUpdateOptions) (runtimeapi.ContainerUpdateResult, error) {
	upd := container.UpdateConfig{}
	if opts.Memory != nil {
		upd.Memory = *opts.Memory
	}
	if opts.NanoCPUs != nil {
		upd.NanoCPUs = *opts.NanoCPUs
	}
	if opts.RestartPolicy != nil {
		upd.RestartPolicy.Name = container.RestartPolicyMode(*opts.RestartPolicy)
		if upd.RestartPolicy.Name == container.RestartPolicyOnFailure && opts.RestartMaxRetries != nil {
			upd.RestartPolicy.MaximumRetryCount = *opts.RestartMaxRetries
		}
	}
	res, err := c.cli.ContainerUpdate(ctx, id, upd)
	if err != nil {
		return runtimeapi.ContainerUpdateResult{}, runtimeapi.MapRuntimeError(err, runtimeapi.Operation(runtimeapi.ResourceContainer, "update"), runtimeapi.ResourceRef{Type: runtimeapi.ResourceContainer, ID: id}, runtimeapi.Docker)
	}
	return runtimeapi.ContainerUpdateResult{Warnings: res.Warnings}, nil
}

// Diff returns filesystem changes between a container and its base image. TASK-019.
func (c *Client) ContainerDiff(ctx context.Context, id string) ([]runtimeapi.ContainerDiffChange, error) {
	raw, err := c.cli.ContainerDiff(ctx, id)
	if err != nil {
		return nil, runtimeapi.MapRuntimeError(err, runtimeapi.Operation(runtimeapi.ResourceContainer, "diff"), runtimeapi.ResourceRef{Type: runtimeapi.ResourceContainer, ID: id}, runtimeapi.Docker)
	}
	out := make([]runtimeapi.ContainerDiffChange, 0, len(raw))
	for _, ch := range raw {
		out = append(out, runtimeapi.ContainerDiffChange{
			Kind: mapDiffKind(ch.Kind),
			Path: ch.Path,
		})
	}
	return out, nil
}

// Export returns a tar stream of the container's filesystem. Caller must close. TASK-019.
func (c *Client) ContainerExport(ctx context.Context, id string) (io.ReadCloser, error) {
	rc, err := c.cli.ContainerExport(ctx, id)
	if err != nil {
		return nil, runtimeapi.MapRuntimeError(err, runtimeapi.Operation(runtimeapi.ResourceContainer, "export"), runtimeapi.ResourceRef{Type: runtimeapi.ResourceContainer, ID: id}, runtimeapi.Docker)
	}
	return rc, nil
}

// Commit snapshots a container's filesystem changes as a new image. TASK-019.
func (c *Client) ContainerCommit(ctx context.Context, id string, opts runtimeapi.ContainerCommitOptions) (runtimeapi.ContainerCommitResult, error) {
	cfg := container.CommitOptions{
		Comment: opts.Comment,
		Author:  opts.Author,
		Pause:   opts.Pause,
	}
	if opts.Repository != "" || opts.Tag != "" {
		cfg.Reference = opts.Repository + ":" + opts.Tag
	}
	res, err := c.cli.ContainerCommit(ctx, id, cfg)
	if err != nil {
		return runtimeapi.ContainerCommitResult{}, runtimeapi.MapRuntimeError(err, runtimeapi.Operation(runtimeapi.ResourceContainer, "commit"), runtimeapi.ResourceRef{Type: runtimeapi.ResourceContainer, ID: id}, runtimeapi.Docker)
	}
	return runtimeapi.ContainerCommitResult{ID: res.ID}, nil
}

// Wait blocks until the container exits or the condition is met. TASK-019.
func (c *Client) ContainerWait(ctx context.Context, id, condition string) (runtimeapi.ContainerWaitResult, error) {
	if condition == "" {
		condition = "not-running"
	}
	resCh, errCh := c.cli.ContainerWait(ctx, id, container.WaitCondition(condition))
	select {
	case err := <-errCh:
		if err != nil {
			return runtimeapi.ContainerWaitResult{}, runtimeapi.MapRuntimeError(err, runtimeapi.Operation(runtimeapi.ResourceContainer, "wait"), runtimeapi.ResourceRef{Type: runtimeapi.ResourceContainer, ID: id}, runtimeapi.Docker)
		}
	case res := <-resCh:
		out := runtimeapi.ContainerWaitResult{StatusCode: res.StatusCode}
		if res.Error != nil {
			out.Error = &runtimeapi.ContainerWaitError{Message: res.Error.Message}
		}
		return out, nil
	}
	return runtimeapi.ContainerWaitResult{}, nil
}

// CopyFromContainer streams a single file or directory out of a container. TASK-019.
func (c *Client) ContainerCopyFrom(ctx context.Context, id, srcPath string) (io.ReadCloser, error) {
	rc, _, err := c.cli.CopyFromContainer(ctx, id, srcPath)
	if err != nil {
		return nil, runtimeapi.MapRuntimeError(err, runtimeapi.Operation(runtimeapi.ResourceContainer, "copy"), runtimeapi.ResourceRef{Type: runtimeapi.ResourceContainer, ID: id}, runtimeapi.Docker)
	}
	return rc, nil
}

// mapDiffKind translates Docker's container.ChangeType enum to our
// runtime enum.
func mapDiffKind(k container.ChangeType) runtimeapi.ChangeKind {
	switch k {
	case container.ChangeAdd:
		return runtimeapi.ChangeAdded
	case container.ChangeDelete:
		return runtimeapi.ChangeDeleted
	default:
		return runtimeapi.ChangeModified
	}
}

// Compile-time check that the adapter implements the extended service.
var _ interface {
	ContainerUpdate(context.Context, string, runtimeapi.ContainerUpdateOptions) (runtimeapi.ContainerUpdateResult, error)
	ContainerDiff(context.Context, string) ([]runtimeapi.ContainerDiffChange, error)
	ContainerExport(context.Context, string) (io.ReadCloser, error)
	ContainerCommit(context.Context, string, runtimeapi.ContainerCommitOptions) (runtimeapi.ContainerCommitResult, error)
	ContainerWait(context.Context, string, string) (runtimeapi.ContainerWaitResult, error)
	ContainerCopyFrom(context.Context, string, string) (io.ReadCloser, error)
} = (*Client)(nil)