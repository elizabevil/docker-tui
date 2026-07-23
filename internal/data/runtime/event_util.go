package runtime

import "context"

// SendEventItem sends an event item to the output channel without blocking.
// Returns false when the context is cancelled.
func SendEventItem(ctx context.Context, output chan<- EventItem, item EventItem) bool {
	select {
	case output <- item:
		return true
	case <-ctx.Done():
		return false
	}
}
