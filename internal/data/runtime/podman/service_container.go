package podman

import (
	"bytes"
	"context"
	"io"
	"net/url"
	"strconv"

	"github.com/docker/docker/pkg/stdcopy"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/driver/podman"
	"github.com/elizabevil/docker-tui/internal/driver/podman/dto"
)

// PodmanContainerService implements ContainerService for the Podman adapter.
type PodmanContainerService struct {
	Client *podman.Client
}

// List returns container summaries, applying client-side post-filtering
// when the Podman API does not support multi-value filter semantics.
func (s PodmanContainerService) List(ctx context.Context, options runtimeapi.ContainerListOptions) ([]runtimeapi.ContainerSummary, error) {
	queryOptions := options
	if options.Filters.HasMultipleValues() {
		queryOptions.Limit = 0
	}
	filters := map[string][]string(queryOptions.Filters)
	raw, err := s.Client.REST.ListContainers(ctx, dto.ContainerListOptions{All: queryOptions.All, Limit: queryOptions.Limit, Filters: filters})
	if err == nil {
		raw, err = PostFilterContainers(raw, options.Filters)
		if options.Limit > 0 && len(raw) > options.Limit {
			raw = raw[:options.Limit]
		}
	}
	if err != nil {
		return nil, mapPodmanContainerErr(err, "list", "")
	}
	return MapContainerSummaries(raw), nil
}

// Inspect returns the full detail view of a container by its ID or name.
func (s PodmanContainerService) Inspect(ctx context.Context, id string) (*runtimeapi.ContainerDetail, error) {
	raw, err := s.Client.REST.InspectContainer(ctx, id)
	if err != nil {
		return nil, mapPodmanContainerErr(err, "inspect", id)
	}
	return MapContainerInspectResponse(*raw), nil
}

// Top returns running processes inside the container (ps-style output).
func (s PodmanContainerService) Top(ctx context.Context, id string) (runtimeapi.ContainerProcesses, error) {
	raw, err := s.Client.REST.ContainerTop(ctx, id)
	if err != nil {
		return runtimeapi.ContainerProcesses{}, mapPodmanContainerErr(err, "top", id)
	}
	return runtimeapi.ContainerProcesses{
		Titles:    raw.Titles,
		Processes: raw.Processes,
	}, nil
}

// Stats returns real-time resource usage statistics for a single container.
func (s PodmanContainerService) Stats(ctx context.Context, id string) (runtimeapi.ContainerStats, error) {
	raw, err := s.Client.REST.ContainerStats(ctx, id)
	if err != nil {
		return runtimeapi.ContainerStats{}, mapPodmanContainerErr(err, "stats", id)
	}
	var networkRx, networkTx float64
	for _, net := range raw.Networks {
		networkRx += float64(net.RxBytes)
		networkTx += float64(net.TxBytes)
	}
	var memPct float64
	if raw.MemoryStats.Limit > 0 {
		memPct = float64(raw.MemoryStats.Usage) / float64(raw.MemoryStats.Limit) * 100
	}
	return runtimeapi.ContainerStats{
		MemoryUsage:   float64(raw.MemoryStats.Usage),
		MemoryLimit:   float64(raw.MemoryStats.Limit),
		MemoryPercent: memPct,
		NetworkRx:     networkRx,
		NetworkTx:     networkTx,
	}, nil
}

// mapPodmanContainerErr wraps a Podman REST error with the container
// resource ref. An empty id builds an untyped ref (used for list operations).
func mapPodmanContainerErr(err error, op, id string) error {
	if err == nil {
		return nil
	}
	ref := runtimeapi.ResourceRef{Type: runtimeapi.ResourceContainer}
	if id != "" {
		ref.ID = id
	}
	return runtimeapi.MapRuntimeError(err,
		runtimeapi.Operation(runtimeapi.ResourceContainer, op), ref, runtimeapi.Podman)
}

// Logs returns a stream of stdout/stderr log lines from the container.
func (s PodmanContainerService) Logs(ctx context.Context, id string, options runtimeapi.ContainerLogOptions) (io.ReadCloser, error) {
	if s.Client.REST == nil {
		return nil, runtimeapi.NewError(runtimeapi.ErrorUnavailable, runtimeapi.Operation(runtimeapi.ResourceContainer, "logs"), id, podman.ErrPodmanRESTNotReady)
	}
	query := url.Values{
		"stdout": {"true"}, "stderr": {"true"}, "follow": {"false"},
		"timestamps": {strconv.FormatBool(options.Timestamps)},
	}
	if options.Since != "" {
		query.Set("since", options.Since)
	}
	if options.Tail != "" {
		query.Set("tail", options.Tail)
	}
	logPath := podman.ContainerPath(id, "/logs")
	reader, err := s.Client.REST.StreamGet(ctx, logPath, query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = reader.Close() }() //nolint:errcheck // log stream consumed by StdCopy.
	var output bytes.Buffer
	if _, err := stdcopy.StdCopy(&output, &output, reader); err != nil {
		return nil, mapPodmanContainerErr(err, "logs", id)
	}
	return io.NopCloser(bytes.NewReader(output.Bytes())), nil
}

// TASK-019 advanced container operations.

// Update applies resource-limit changes to a running container.
func (s PodmanContainerService) Update(ctx context.Context, id string, options runtimeapi.ContainerUpdateOptions) (runtimeapi.ContainerUpdateResult, error) {
	if s.Client.REST == nil {
		return runtimeapi.ContainerUpdateResult{}, runtimeapi.NewError(runtimeapi.ErrorUnavailable,
			runtimeapi.Operation(runtimeapi.ResourceContainer, "update"), id, podman.ErrPodmanRESTNotReady)
	}
	upd := dto.ContainerUpdateOptions{}
	if options.Memory != nil {
		upd.Memory = *options.Memory
	}
	if options.NanoCPUs != nil {
		upd.NanoCPUs = *options.NanoCPUs
	}
	if options.RestartPolicy != nil {
		upd.RestartPolicy = *options.RestartPolicy
	}
	if options.RestartMaxRetries != nil {
		upd.RestartMaxRetries = *options.RestartMaxRetries
	}
	res, err := s.Client.REST.ContainerUpdate(ctx, id, upd)
	if err != nil {
		return runtimeapi.ContainerUpdateResult{}, mapPodmanContainerErr(err, "update", id)
	}
	if res == nil {
		return runtimeapi.ContainerUpdateResult{}, nil
	}
	return runtimeapi.ContainerUpdateResult{Warnings: res.Warnings}, nil
}

// Diff returns filesystem changes between a container and its base image.
func (s PodmanContainerService) Diff(ctx context.Context, id string) ([]runtimeapi.ContainerDiffChange, error) {
	if s.Client.REST == nil {
		return nil, runtimeapi.NewError(runtimeapi.ErrorUnavailable,
			runtimeapi.Operation(runtimeapi.ResourceContainer, "diff"), id, podman.ErrPodmanRESTNotReady)
	}
	raw, err := s.Client.REST.ContainerDiff(ctx, id)
	if err != nil {
		return nil, mapPodmanContainerErr(err, "diff", id)
	}
	out := make([]runtimeapi.ContainerDiffChange, 0, len(raw))
	for _, ch := range raw {
		out = append(out, runtimeapi.ContainerDiffChange{
			Kind: mapPodmanChangeKind(ch.Kind),
			Path: ch.Path,
		})
	}
	return out, nil
}

// Export returns a tar stream of the container's filesystem. Caller must close.
func (s PodmanContainerService) Export(ctx context.Context, id string) (io.ReadCloser, error) {
	if s.Client.REST == nil {
		return nil, runtimeapi.NewError(runtimeapi.ErrorUnavailable,
			runtimeapi.Operation(runtimeapi.ResourceContainer, "export"), id, podman.ErrPodmanRESTNotReady)
	}
	rc, err := s.Client.REST.ContainerExport(ctx, id)
	if err != nil {
		return nil, mapPodmanContainerErr(err, "export", id)
	}
	return rc, nil
}

// Commit snapshots a container's filesystem changes as a new image.
func (s PodmanContainerService) Commit(ctx context.Context, id string, options runtimeapi.ContainerCommitOptions) (runtimeapi.ContainerCommitResult, error) {
	if s.Client.REST == nil {
		return runtimeapi.ContainerCommitResult{}, runtimeapi.NewError(runtimeapi.ErrorUnavailable,
			runtimeapi.Operation(runtimeapi.ResourceContainer, "commit"), id, podman.ErrPodmanRESTNotReady)
	}
	cfg := dto.ContainerCommitOptions{
		Repository: options.Repository,
		Tag:        options.Tag,
		Comment:    options.Comment,
		Author:     options.Author,
		Pause:      options.Pause,
	}
	res, err := s.Client.REST.ContainerCommit(ctx, id, cfg)
	if err != nil {
		return runtimeapi.ContainerCommitResult{}, mapPodmanContainerErr(err, "commit", id)
	}
	if res == nil {
		return runtimeapi.ContainerCommitResult{}, nil
	}
	return runtimeapi.ContainerCommitResult{ID: res.ID}, nil
}

// Wait blocks until the container exits or the condition is met.
func (s PodmanContainerService) Wait(ctx context.Context, id, condition string) (runtimeapi.ContainerWaitResult, error) {
	if s.Client.REST == nil {
		return runtimeapi.ContainerWaitResult{}, runtimeapi.NewError(runtimeapi.ErrorUnavailable,
			runtimeapi.Operation(runtimeapi.ResourceContainer, "wait"), id, podman.ErrPodmanRESTNotReady)
	}
	res, err := s.Client.REST.ContainerWait(ctx, id, condition)
	if err != nil {
		return runtimeapi.ContainerWaitResult{}, mapPodmanContainerErr(err, "wait", id)
	}
	if res == nil {
		return runtimeapi.ContainerWaitResult{}, nil
	}
	out := runtimeapi.ContainerWaitResult{StatusCode: res.StatusCode}
	if res.Error != "" {
		out.Error = &runtimeapi.ContainerWaitError{Message: res.Error}
	}
	return out, nil
}

// CopyFromContainer streams a single file or directory out of a container.
func (s PodmanContainerService) CopyFromContainer(ctx context.Context, id, srcPath string) (io.ReadCloser, error) {
	if s.Client.REST == nil {
		return nil, runtimeapi.NewError(runtimeapi.ErrorUnavailable,
			runtimeapi.Operation(runtimeapi.ResourceContainer, "copy"), id, podman.ErrPodmanRESTNotReady)
	}
	rc, err := s.Client.REST.ContainerCopyFrom(ctx, id, srcPath)
	if err != nil {
		return nil, mapPodmanContainerErr(err, "copy", id)
	}
	return rc, nil
}

// mapPodmanChangeKind translates the podman dto ChangeKind into the
// runtime enum. The dto mirrors go.podman.io/image/v5/archive.Change.
func mapPodmanChangeKind(k dto.ChangeKind) runtimeapi.ChangeKind {
	switch k {
	case dto.ArchiveChangeAdded:
		return runtimeapi.ChangeAdded
	case dto.ArchiveChangeDeleted:
		return runtimeapi.ChangeDeleted
	default:
		return runtimeapi.ChangeModified
	}
}
