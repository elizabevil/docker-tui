package docker

import (
	"context"

	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

type eventService struct{ client *Client }

func (c *Client) Events() runtimeapi.EventService { return eventService{client: c} }

func (s eventService) Subscribe(ctx context.Context, options runtimeapi.EventOptions) (<-chan runtimeapi.EventItem, error) {
	nativeFilters := filters.NewArgs()
	for field, values := range options.Filters {
		for _, value := range values {
			nativeFilters.Add(field, value)
		}
	}
	msgCh, errCh := s.client.cli.Events(ctx, events.ListOptions{Filters: nativeFilters})
	out := make(chan runtimeapi.EventItem, 100)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgCh:
				if !ok {
					return
				}
				out <- runtimeapi.EventItem{Event: mapDockerEvent(msg)}
			case err, ok := <-errCh:
				if !ok {
					return
				}
				out <- runtimeapi.EventItem{Error: mapRuntimeError(err, "events.subscribe", runtimeapi.ResourceRef{}, s.client.RuntimeType)}
			}
		}
	}()
	return out, nil
}

func mapDockerEvent(message events.Message) runtimeapi.Event {
	return runtimeapi.Event{
		ResourceType: string(message.Type), Action: string(message.Action), ActorID: message.Actor.ID,
		Attributes: message.Actor.Attributes, Scope: message.Scope, Time: message.Time, TimeNano: message.TimeNano,
	}
}
