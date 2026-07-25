package podman

import (
	"context"
	"io"
	"net/url"
	"strconv"

	"github.com/elizabevil/docker-tui/internal/driver/podman/dto"
)

// ContainerUpdate applies resource-limit changes via POST /libpod/containers/{id}/update.
// TASK-019.
func (c *RESTClient) ContainerUpdate(ctx context.Context, id string, options dto.ContainerUpdateOptions) (*dto.ContainerUpdateResponse, error) {
	query := url.Values{}
	if options.Memory > 0 {
		query.Set("memory", strconv.FormatInt(options.Memory, 10))
	}
	if options.NanoCPUs > 0 {
		query.Set("cpus", strconv.FormatFloat(float64(options.NanoCPUs)/1e9, 'f', 6, 64))
	}
	if options.RestartPolicy != "" {
		query.Set("restartPolicy", options.RestartPolicy)
	}
	if options.RestartMaxRetries > 0 {
		query.Set("restartMaxRetries", strconv.Itoa(options.RestartMaxRetries))
	}
	var raw dto.ContainerUpdateResponse
	if err := c.Post(ctx, ContainerUpdatePath(id), query, nil, &raw); err != nil {
		return nil, err
	}
	return &raw, nil
}

// ContainerDiff (a.k.a. changes) reports filesystem changes vs the image.
// TASK-019.
func (c *RESTClient) ContainerDiff(ctx context.Context, id string) ([]dto.ArchiveChange, error) {
	var raw []dto.ArchiveChange
	if err := c.Get(ctx, ContainerChangesPath(id), nil, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

// ContainerExport streams the container's filesystem as a tar archive.
// Caller must close. TASK-019.
func (c *RESTClient) ContainerExport(ctx context.Context, id string) (io.ReadCloser, error) {
	return c.StreamGet(ctx, ContainerExportPath(id), nil)
}

// ContainerCommit snapshots the container as a new image. TASK-019.
func (c *RESTClient) ContainerCommit(ctx context.Context, id string, options dto.ContainerCommitOptions) (*dto.ContainerCommitResponse, error) {
	query := url.Values{}
	if options.Repository != "" || options.Tag != "" {
		repo := options.Repository + ":" + options.Tag
		query.Set("repo", repo)
	}
	if options.Comment != "" {
		query.Set("comment", options.Comment)
	}
	if options.Author != "" {
		query.Set("author", options.Author)
	}
	if options.Pause {
		query.Set("pause", "true")
	} else {
		query.Set("pause", "false")
	}
	var raw dto.ContainerCommitResponse
	if err := c.Post(ctx, ContainerCommitPath(id), query, nil, &raw); err != nil {
		return nil, err
	}
	return &raw, nil
}

// ContainerWait blocks until the container exits or the condition is met.
// TASK-019.
func (c *RESTClient) ContainerWait(ctx context.Context, id, condition string) (*dto.ContainerWaitResponse, error) {
	query := url.Values{}
	if condition != "" {
		query.Set("condition", condition)
	}
	var raw dto.ContainerWaitResponse
	if err := c.Post(ctx, ContainerWaitPath(id), query, nil, &raw); err != nil {
		return nil, err
	}
	return &raw, nil
}

// ContainerCopyFrom streams a path out of the container as a tar archive.
// Caller must close. TASK-019.
func (c *RESTClient) ContainerCopyFrom(ctx context.Context, id, srcPath string) (io.ReadCloser, error) {
	query := url.Values{"path": {srcPath}}
	return c.StreamGet(ctx, ContainerArchivePath(id), query)
}