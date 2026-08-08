package runtime

import (
	"context"
	"errors"
	"testing"
	"time"
)

type stubEngine struct {
	id      string
	closeOK bool
}

func (s *stubEngine) Identity() Identity                   { return Identity{Endpoint: s.id} }
func (s *stubEngine) Capabilities() CapabilitySet          { return CapabilitySet{} }
func (s *stubEngine) PingContext(_ context.Context) error  { return nil }
func (s *stubEngine) Containers() ContainerService         { return nil }
func (s *stubEngine) Volumes() VolumeService               { return nil }
func (s *stubEngine) Networks() NetworkService             { return nil }
func (s *stubEngine) Images() ImageService                 { return nil }
func (s *stubEngine) ImageTransfers() ImageTransferService { return nil }
func (s *stubEngine) Actions() ResourceActionService       { return nil }
func (s *stubEngine) Exec() ExecService                    { return nil }
func (s *stubEngine) Events() EventService                 { return nil }
func (s *stubEngine) Compose() ComposeService               { return nil }
func (s *stubEngine) Pods() PodService                     { return nil }
func (s *stubEngine) Close() error {
	s.closeOK = true
	return nil
}

// typedNilEngine mirrors the real adapter bug: a function returning
// (*concrete)(nil), error gets wrapped into a non-nil Engine interface.
func typedNilEngine(err error) (Engine, error) {
	var s *stubEngine // typed nil
	return s, err
}

func TestEngineIsUsableCatchesTypedNil(t *testing.T) {
	e, err := typedNilEngine(errors.New("boom"))
	if e == nil {
		t.Fatal("typed-nil pointer wrapped in interface should not equal nil")
	}
	if engineIsUsable(e) {
		t.Fatal("engineIsUsable must reject typed-nil pointer")
	}
	if err == nil {
		t.Fatal("expected an error")
	}
}

func TestEngineIsUsableAcceptsRealImplementation(t *testing.T) {
	e := &stubEngine{id: "ok"}
	if !engineIsUsable(e) {
		t.Fatal("real implementation must be usable")
	}
}

func TestSanitizeEngineCollapsesTypedNil(t *testing.T) {
	e, _ := typedNilEngine(errors.New("boom"))
	if got := sanitizeEngine(e, nil); got != nil {
		t.Fatalf("sanitizeEngine should collapse typed-nil to nil, got %#v", got)
	}
	e2 := &stubEngine{id: "ok"}
	if got := sanitizeEngine(e2, nil); got != e2 {
		t.Fatal("sanitizeEngine must preserve real engines")
	}
}

func TestRefreshAllDoesNotPanicOnAdapterFailure(t *testing.T) {
	pool := NewPool(func(_ ClientConfig) (Engine, error) {
		return typedNilEngine(errors.New("adapter failure"))
	})
	pool.AddHost(ConnectionSpec{Name: "broken", Host: "tcp://example:2375", Runtime: RuntimeDocker})

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("RefreshAll must not panic on typed-nil engine: %v", r)
		}
	}()

	results := pool.RefreshAll(50 * time.Millisecond)
	if len(results) != 1 || results[0].Error == nil {
		t.Fatalf("expected one failure result, got %#v", results)
	}
}
