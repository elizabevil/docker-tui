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
		return nil, runtimeapi.MapRuntimeError(err, runtimeapi.Operation(runtimeapi.ResourceContainer, "list"), runtimeapi.ResourceRef{Type: runtimeapi.ResourceContainer}, runtimeapi.Podman)
	}
	return MapContainerSummaries(raw), nil
}

func (s PodmanContainerService) Inspect(ctx context.Context, id string) (*runtimeapi.ContainerDetail, error) {
	raw, err := s.Client.REST.InspectContainer(ctx, id)
	if err != nil {
		return nil, runtimeapi.MapRuntimeError(err, runtimeapi.Operation(runtimeapi.ResourceContainer, "inspect"), runtimeapi.ResourceRef{Type: runtimeapi.ResourceContainer, ID: id}, runtimeapi.Podman)
	}
	return MapContainerInspectResponse(*raw), nil
}

func (s PodmanContainerService) Top(ctx context.Context, id string) (runtimeapi.ContainerProcesses, error) {
	raw, err := s.Client.REST.ContainerTop(ctx, id)
	if err != nil {
		return runtimeapi.ContainerProcesses{}, runtimeapi.MapRuntimeError(err, runtimeapi.Operation(runtimeapi.ResourceContainer, "top"), runtimeapi.ResourceRef{Type: runtimeapi.ResourceContainer, ID: id}, runtimeapi.Podman)
	}
	return runtimeapi.ContainerProcesses{
		Titles:    raw.Titles,
		Processes: raw.Processes,
	}, nil
}

func (s PodmanContainerService) Stats(ctx context.Context, id string) (runtimeapi.ContainerStats, error) {
	raw, err := s.Client.REST.ContainerStats(ctx, id)
	if err != nil {
		return runtimeapi.ContainerStats{}, runtimeapi.MapRuntimeError(err, runtimeapi.Operation(runtimeapi.ResourceContainer, "stats"), runtimeapi.ResourceRef{Type: runtimeapi.ResourceContainer, ID: id}, runtimeapi.Podman)
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
	defer reader.Close()
	var output bytes.Buffer
	if _, err := stdcopy.StdCopy(&output, &output, reader); err != nil {
		return nil, runtimeapi.MapRuntimeError(err, runtimeapi.Operation(runtimeapi.ResourceContainer, "logs"), runtimeapi.ResourceRef{Type: runtimeapi.ResourceContainer, ID: id}, runtimeapi.Podman)
	}
	return io.NopCloser(bytes.NewReader(output.Bytes())), nil
}
