package podman

import (
	"context"
	"encoding/json"
	"io"
	"net/url"
	"strings"
	"time"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	podman "github.com/elizabevil/docker-tui/internal/driver/podman"
)

// PodmanEventService implements EventService for the Podman adapter.
type PodmanEventService struct {
	Client *podman.Client
}

func (s PodmanEventService) Subscribe(ctx context.Context, options runtimeapi.EventOptions) (<-chan runtimeapi.EventItem, error) {
	if s.Client.REST == nil {
		return nil, runtimeapi.NewError(runtimeapi.ErrorUnavailable, runtimeapi.Operation(runtimeapi.ResourceEvent, "subscribe"), "", podman.ErrPodmanRESTNotReady)
	}
	query := url.Values{"stream": {"true"}}
	if len(options.Filters) > 0 {
		encoded, err := json.Marshal(options.Filters)
		if err != nil {
			return nil, runtimeapi.NewError(runtimeapi.ErrorInvalid, runtimeapi.Operation(runtimeapi.ResourceEvent, "subscribe"), "", err)
		}
		query.Set("filters", string(encoded))
	}
	reader, err := s.Client.REST.StreamGet(ctx, podman.PathEvents, query)
	if err != nil {
		return nil, err
	}
	output := make(chan runtimeapi.EventItem, 100)
	go func() {
		defer close(output)
		defer reader.Close()
		decoder := json.NewDecoder(reader)
		for {
			var event podman.EventItem
			if err := decoder.Decode(&event); err != nil {
				if err != io.EOF && ctx.Err() == nil {
					runtimeapi.SendEventItem(ctx, output, runtimeapi.EventItem{Error: runtimeapi.MapRuntimeError(err, runtimeapi.Operation(runtimeapi.ResourceEvent, "subscribe"), runtimeapi.ResourceRef{}, runtimeapi.Podman)})
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
			if !runtimeapi.SendEventItem(ctx, output, item) {
				return
			}
		}
	}()
	return output, nil
}
