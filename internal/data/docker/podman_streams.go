package docker

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

type podmanContainerService struct{ client *Client }

func (s podmanContainerService) List(ctx context.Context, options runtimeapi.ContainerListOptions) ([]runtimeapi.ContainerSummary, error) {
	queryOptions := options
	if options.Filters.HasMultipleValues() {
		queryOptions.Limit = 0
	}
	items, err := s.client.listContainersPodman(ctx, queryOptions)
	if err == nil {
		items, err = postFilterContainers(items, options.Filters)
		if options.Limit > 0 && len(items) > options.Limit {
			items = items[:options.Limit]
		}
	}
	return items, mapRuntimeError(err, "container.list", runtimeapi.ResourceRef{Type: runtimeapi.ResourceContainer}, RuntimePodman)
}

func (s podmanContainerService) Inspect(ctx context.Context, id string) (*runtimeapi.ContainerDetail, error) {
	detail, err := s.client.inspectContainerPodmanREST(ctx, id)
	return detail, mapRuntimeError(err, "container.inspect", runtimeapi.ResourceRef{Type: runtimeapi.ResourceContainer, ID: id}, RuntimePodman)
}

func (s podmanContainerService) Top(ctx context.Context, id string) (runtimeapi.ContainerProcesses, error) {
	processes, err := s.client.containerTopPodmanREST(ctx, id)
	return processes, mapRuntimeError(err, "container.top", runtimeapi.ResourceRef{Type: runtimeapi.ResourceContainer, ID: id}, RuntimePodman)
}

func (s podmanContainerService) Stats(ctx context.Context, id string) (runtimeapi.ContainerStats, error) {
	stats, err := s.client.containerStatsPodmanREST(ctx, id)
	return stats, mapRuntimeError(err, "container.stats", runtimeapi.ResourceRef{Type: runtimeapi.ResourceContainer, ID: id}, RuntimePodman)
}

func (s podmanContainerService) Logs(ctx context.Context, id string, options runtimeapi.ContainerLogOptions) (io.ReadCloser, error) {
	if s.client.podmanREST == nil {
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
	reader, err := s.client.podmanREST.StreamGet(ctx, "container.logs", "/containers/"+url.PathEscape(id)+"/logs", query)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	var output bytes.Buffer
	if _, err := stdcopy.StdCopy(&output, &output, reader); err != nil {
		return nil, mapRuntimeError(err, "container.logs", runtimeapi.ResourceRef{Type: runtimeapi.ResourceContainer, ID: id}, RuntimePodman)
	}
	return io.NopCloser(bytes.NewReader(output.Bytes())), nil
}

type podmanEventService struct{ client *Client }

type podmanEvent struct {
	Type       string            `json:"Type"`
	Status     string            `json:"Status"`
	Action     string            `json:"Action"`
	ID         string            `json:"ID"`
	Name       string            `json:"Name"`
	Time       time.Time         `json:"Time"`
	TimeNano   int64             `json:"TimeNano"`
	Attributes map[string]string `json:"Attributes"`
}

func (s podmanEventService) Subscribe(ctx context.Context, options runtimeapi.EventOptions) (<-chan runtimeapi.EventItem, error) {
	if s.client.podmanREST == nil {
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
	reader, err := s.client.podmanREST.StreamGet(ctx, "events.subscribe", "/events", query)
	if err != nil {
		return nil, err
	}
	output := make(chan runtimeapi.EventItem, 100)
	go func() {
		defer close(output)
		defer reader.Close()
		decoder := json.NewDecoder(reader)
		for {
			var event podmanEvent
			if err := decoder.Decode(&event); err != nil {
				if err != io.EOF && ctx.Err() == nil {
					sendEventItem(ctx, output, runtimeapi.EventItem{Error: mapRuntimeError(err, "events.subscribe", runtimeapi.ResourceRef{}, RuntimePodman)})
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
			if !sendEventItem(ctx, output, item) {
				return
			}
		}
	}()
	return output, nil
}
