package state

import "context"

// ContainerWaitState owns the context for an in-flight blocking container
// Wait. The keyboard layer registers the wait's context here (Begin) so a
// long-running Wait can be cancelled when the user switches runtime, quits
// the app, or otherwise aborts the operation — it must never hang on
// context.Background(). Mirrors ImageTransferState.
type ContainerWaitState struct {
	Generation uint64
	TargetID   string
	Context    context.Context
	Cancel     context.CancelFunc
}

// Begin starts a new cancellable wait, cancelling any previous one.
func (s *ContainerWaitState) Begin(targetID string) (context.Context, uint64) {
	s.Stop()
	s.Generation++
	s.TargetID = targetID
	s.Context, s.Cancel = context.WithCancel(context.Background())
	return s.Context, s.Generation
}

// Stop cancels any active wait and detaches its context.
func (s *ContainerWaitState) Stop() {
	if s.Cancel != nil {
		s.Cancel()
	}
	s.Context = nil
	s.Cancel = nil
	s.TargetID = ""
}

// Current reports whether a wait with the given generation is still active.
func (s *ContainerWaitState) Current(generation uint64) bool {
	return generation != 0 && generation == s.Generation && s.Cancel != nil
}

// IsActive reports whether a wait is currently running. The renderer
// layer uses this to decide whether to paint the wait status indicator
// in the message rail.
func (s *ContainerWaitState) IsActive() bool {
	return s != nil && s.Cancel != nil
}
