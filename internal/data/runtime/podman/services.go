package podman

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/pkg/stdcopy"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

// ContainerService implements runtime.ContainerService for the Podman adapter.
type ContainerService struct {
	Client *Client
}

func (s ContainerService) List(ctx context.Context, options runtimeapi.ContainerListOptions) ([]runtimeapi.ContainerSummary, error) {
	queryOptions := options
	if options.Filters.HasMultipleValues() {
		queryOptions.Limit = 0
	}
	filters := map[string][]string(queryOptions.Filters)
	raw, err := ListContainersREST(ctx, s.Client, queryOptions.All, queryOptions.Limit, filters)
	if err == nil {
		raw, err = PostFilterContainers(raw, options.Filters)
		if options.Limit > 0 && len(raw) > options.Limit {
			raw = raw[:options.Limit]
		}
	}
	if err != nil {
		return nil, mapRuntimeError(err, "container.list", runtimeapi.ResourceRef{Type: runtimeapi.ResourceContainer}, runtimeapi.Podman)
	}
	return MapContainerSummaries(raw), nil
}

func (s ContainerService) Inspect(ctx context.Context, id string) (*runtimeapi.ContainerDetail, error) {
	raw, err := InspectContainerREST(ctx, s.Client, id)
	if err != nil {
		return nil, mapRuntimeError(err, "container.inspect", runtimeapi.ResourceRef{Type: runtimeapi.ResourceContainer, ID: id}, runtimeapi.Podman)
	}
	return MapContainerInspectResponse(*raw), nil
}

func (s ContainerService) Top(ctx context.Context, id string) (runtimeapi.ContainerProcesses, error) {
	raw, err := ContainerTopREST(ctx, s.Client, id)
	if err != nil {
		return runtimeapi.ContainerProcesses{}, mapRuntimeError(err, "container.top", runtimeapi.ResourceRef{Type: runtimeapi.ResourceContainer, ID: id}, runtimeapi.Podman)
	}
	return runtimeapi.ContainerProcesses{
		Titles:    raw.Titles,
		Processes: raw.Processes,
	}, nil
}

func (s ContainerService) Stats(ctx context.Context, id string) (runtimeapi.ContainerStats, error) {
	raw, err := ContainerStatsREST(ctx, s.Client, id)
	if err != nil {
		return runtimeapi.ContainerStats{}, mapRuntimeError(err, "container.stats", runtimeapi.ResourceRef{Type: runtimeapi.ResourceContainer, ID: id}, runtimeapi.Podman)
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

func (s ContainerService) Logs(ctx context.Context, id string, options runtimeapi.ContainerLogOptions) (io.ReadCloser, error) {
	if s.Client.REST == nil {
		return nil, runtimeapi.NewError(runtimeapi.ErrorUnavailable, "container.logs", id, errPodmanRESTNotReady)
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
	path := ContainerPath(id, "/logs")
	reader, err := s.Client.REST.StreamGet(ctx, "container.logs", path, query)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	var output bytes.Buffer
	if _, err := stdcopy.StdCopy(&output, &output, reader); err != nil {
		return nil, mapRuntimeError(err, "container.logs", runtimeapi.ResourceRef{Type: runtimeapi.ResourceContainer, ID: id}, runtimeapi.Podman)
	}
	return io.NopCloser(bytes.NewReader(output.Bytes())), nil
}

// EventService implements runtime.EventService for the Podman adapter.
type EventService struct {
	Client *Client
}

func (s EventService) Subscribe(ctx context.Context, options runtimeapi.EventOptions) (<-chan runtimeapi.EventItem, error) {
	if s.Client.REST == nil {
		return nil, runtimeapi.NewError(runtimeapi.ErrorUnavailable, "events.subscribe", "", errPodmanRESTNotReady)
	}
	query := url.Values{"stream": {"true"}}
	if len(options.Filters) > 0 {
		encoded, err := json.Marshal(options.Filters)
		if err != nil {
			return nil, runtimeapi.NewError(runtimeapi.ErrorInvalid, "events.subscribe", "", err)
		}
		query.Set("filters", string(encoded))
	}
	reader, err := s.Client.REST.StreamGet(ctx, "events.subscribe", PathEvents, query)
	if err != nil {
		return nil, err
	}
	output := make(chan runtimeapi.EventItem, 100)
	go func() {
		defer close(output)
		defer reader.Close()
		decoder := json.NewDecoder(reader)
		for {
			var event EventItem
			if err := decoder.Decode(&event); err != nil {
				if err != io.EOF && ctx.Err() == nil {
					SendEventItem(ctx, output, runtimeapi.EventItem{Error: mapRuntimeError(err, "events.subscribe", runtimeapi.ResourceRef{}, runtimeapi.Podman)})
				}
				return
			}
			action := event.Action
			if action == "" {
				action = event.Status
			}
			attributes := event.Attributes
			if attributes == nil {
				attributes = make(map[string]string)
			}
			if event.Name != "" {
				attributes["name"] = event.Name
			}
			timeNano := event.TimeNano
			if timeNano == 0 && !event.Time.IsZero() {
				timeNano = event.Time.UnixNano()
			}
			item := runtimeapi.EventItem{Event: runtimeapi.Event{
				ResourceType: strings.ToLower(event.Type), Action: strings.ToLower(action), ActorID: event.ID,
				Attributes: attributes, Scope: "local", Time: timeNano / int64(time.Second), TimeNano: timeNano,
			}}
			if !SendEventItem(ctx, output, item) {
				return
			}
		}
	}()
	return output, nil
}
