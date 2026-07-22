package docker

import (
	"context"

	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

// eventService implements runtime.EventService for the Docker/Podman adapter.
type eventService struct{ client *Client }

// Events returns the event service facade.
func (c *Client) Events() runtimeapi.EventService { return eventService{client: c} }

// Subscribe starts listening for runtime events and returns a channel that
// receives EventItem values until the context is cancelled.
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
				if !sendEventItem(ctx, out, runtimeapi.EventItem{Event: mapDockerEvent(msg)}) {
					return
				}
			case err, ok := <-errCh:
				if !ok {
					return
				}
				item := runtimeapi.EventItem{Error: mapRuntimeError(err, "events.subscribe", runtimeapi.ResourceRef{}, s.client.RuntimeType)}
				if !sendEventItem(ctx, out, item) {
					return
				}
			}
		}
	}()
	return out, nil
}

func sendEventItem(ctx context.Context, output chan<- runtimeapi.EventItem, item runtimeapi.EventItem) bool {
	select {
	case output <- item:
		return true
	case <-ctx.Done():
		return false
	}
}

func mapDockerEvent(message events.Message) runtimeapi.Event {
	return runtimeapi.Event{
		ResourceType: string(message.Type), Action: string(message.Action), ActorID: message.Actor.ID,
		Attributes: message.Actor.Attributes, Scope: message.Scope, Time: message.Time, TimeNano: message.TimeNano,
	}
}
