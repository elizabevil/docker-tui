package podman

import (
	"context"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

// SendEventItem sends an event item to the output channel without blocking.
// Returns false when the context is cancelled.
func SendEventItem(ctx context.Context, output chan<- runtimeapi.EventItem, item runtimeapi.EventItem) bool {
	select {
	case output <- item:
		return true
	case <-ctx.Done():
		return false
	}
}
