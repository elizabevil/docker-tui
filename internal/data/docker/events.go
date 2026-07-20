package docker

import (
	"github.com/docker/docker/api/types/events"
)

type EventsMsg struct {
	events.Message
	Error error
}

func (c *Client) Events() (<-chan EventsMsg, error) {
	msgCh, errCh := c.cli.Events(c.ctx, events.ListOptions{})
	out := make(chan EventsMsg, 100)
	go func() {
		defer close(out)
		for {
			select {
			case <-c.ctx.Done():
				return
			case msg, ok := <-msgCh:
				if !ok {
					return
				}
				out <- EventsMsg{Message: msg}
			case err, ok := <-errCh:
				if !ok {
					return
				}
				out <- EventsMsg{Error: err}
			}
		}
	}()
	return out, nil
}
