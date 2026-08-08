// Package mockengine provides a runtime.Engine stub used by batch-action
// and event-handling tests across the codebase. It is intentionally
// exported so other internal packages can wire it into AppModel fixtures.
package mockengine

import (
	"context"
	"errors"
	"sync"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

// Engine is a runtime.Engine stub. It records every Execute call and can
// be configured to fail specific resource IDs.
type Engine struct {
	mu           sync.Mutex
	execCalls    []runtimeapi.ResourceRef
	removeCalls  []string // records volume/network Remove calls
	failIDs      map[string]error
	failAll      error
	identity     runtimeapi.Identity
	capabilities runtimeapi.CapabilitySet
}

// New returns a fresh mock engine with sensible defaults for unit tests.
func New() *Engine {
	return &Engine{
		failIDs: map[string]error{},
		identity: runtimeapi.Identity{
			Type: runtimeapi.Docker,
			Name: "test",
		},
		capabilities: runtimeapi.CapabilitySet{
			runtimeapi.CapabilityEvents: {Support: runtimeapi.Available},
		},
	}
}

// Fail marks a specific resource ID as failing with err.
func (m *Engine) Fail(id string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failIDs[id] = err
}

// FailAll makes every Execute call fail with err.
func (m *Engine) FailAll(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failAll = err
}

// CallCount returns the number of recorded Execute calls.
func (m *Engine) CallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.execCalls)
}

// Calls returns a copy of the recorded Execute calls (ResourceRefs).
func (m *Engine) Calls() []runtimeapi.ResourceRef {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]runtimeapi.ResourceRef, len(m.execCalls))
	copy(out, m.execCalls)
	return out
}

// RemoveCount returns the number of recorded Volume/Network Remove calls.
func (m *Engine) RemoveCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.removeCalls)
}

// Identity implements runtimeapi.Engine.
func (m *Engine) Identity() runtimeapi.Identity { return m.identity }

// SetIdentity overrides the identity reported by the engine. Tests use
// this to exercise runtime-dependent UI branches (e.g. Podman-specific
// compose scale interception).
func (m *Engine) SetIdentity(identity runtimeapi.Identity) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.identity = identity
}

// Capabilities implements runtimeapi.Engine.
func (m *Engine) Capabilities() runtimeapi.CapabilitySet { return m.capabilities }

// SetCapability records one Capability's support level in the stub's
// CapabilitySet. Tests use this to simulate a runtime with or without
// a given capability. Setting Support=Unsupported removes the entry,
// matching the loader's "unsupported = absent" convention.
func (m *Engine) SetCapability(c runtimeapi.Capability, info runtimeapi.CapabilityInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.capabilities == nil {
		m.capabilities = runtimeapi.CapabilitySet{}
	}
	if info.Support == runtimeapi.Unsupported {
		delete(m.capabilities, c)
		return
	}
	m.capabilities[c] = info
}

// PingContext implements runtimeapi.Engine.
func (m *Engine) PingContext(context.Context) error { return nil }

// Containers implements runtimeapi.Engine.
func (m *Engine) Containers() runtimeapi.ContainerService { return nil }

// Volumes implements runtimeapi.Engine.
func (m *Engine) Volumes() runtimeapi.VolumeService { return &volumeService{engine: m} }

// Networks implements runtimeapi.Engine.
func (m *Engine) Networks() runtimeapi.NetworkService { return &networkService{engine: m} }

// Images implements runtimeapi.Engine.
func (m *Engine) Images() runtimeapi.ImageService { return nil }

// ImageTransfers implements runtimeapi.Engine.
func (m *Engine) ImageTransfers() runtimeapi.ImageTransferService { return nil }

// Actions returns a ResourceActionService stub honoring Fail/FailAll.
func (m *Engine) Actions() runtimeapi.ResourceActionService { return &actionService{engine: m} }

// Exec implements runtimeapi.Engine.
func (m *Engine) Exec() runtimeapi.ExecService { return nil }

// Events implements runtimeapi.Engine.
func (m *Engine) Events() runtimeapi.EventService { return nil }

func (m *Engine) Compose() runtimeapi.ComposeService { return composeStub{} }

func (m *Engine) Pods() runtimeapi.PodService { return podStub{} }

// Close implements runtimeapi.Engine.
func (m *Engine) Close() error { return nil }

type composeStub struct{}

func (composeStub) ListProjects(context.Context, runtimeapi.ListProjectsOptions) ([]runtimeapi.ComposeProjectSummary, error) {
	return nil, nil
}
func (composeStub) InspectProject(context.Context, string) (runtimeapi.ComposeProjectSummary, error) {
	return runtimeapi.ComposeProjectSummary{}, nil
}
func (composeStub) Config(context.Context, string) (string, error)    { return "", nil }
func (composeStub) Start(context.Context, string, []string) error     { return nil }
func (composeStub) Stop(context.Context, string, []string, int) error { return nil }
func (composeStub) Restart(context.Context, string, []string) error   { return nil }
func (composeStub) Down(context.Context, string, runtimeapi.DownOptions) error {
	return nil
}
func (composeStub) Up(context.Context, string, runtimeapi.UpOptions) error { return nil }
func (composeStub) Run(context.Context, string, string, []string, runtimeapi.RunOptions) error {
	return nil
}
func (composeStub) Exec(context.Context, string, string, []string) error { return nil }
func (composeStub) Top(context.Context, string, string) (runtimeapi.ContainerProcesses, error) {
	return runtimeapi.ContainerProcesses{}, nil
}
func (composeStub) Port(context.Context, string, string, int) (string, error) {
	return "", nil
}
func (composeStub) Stats(context.Context, string, string) (runtimeapi.ContainerStats, error) {
	return runtimeapi.ContainerStats{}, nil
}
func (composeStub) Events(context.Context, string) (<-chan runtimeapi.ComposeEvent, error) {
	return nil, nil
}

type podStub struct{}

func (podStub) ListPods(context.Context, runtimeapi.PodListOptions) ([]runtimeapi.PodSummary, error) {
	return nil, nil
}
func (podStub) InspectPod(context.Context, string) (runtimeapi.PodDetail, error) {
	return runtimeapi.PodDetail{}, nil
}

type actionService struct{ engine *Engine }

func (a *actionService) Execute(_ context.Context, ref runtimeapi.ResourceRef, _ runtimeapi.Action, _ runtimeapi.ActionOptions) (runtimeapi.ActionResult, error) {
	a.engine.mu.Lock()
	defer a.engine.mu.Unlock()
	a.engine.execCalls = append(a.engine.execCalls, ref)
	if err := a.engine.failAll; err != nil {
		return runtimeapi.ActionResult{}, err
	}
	if err := a.engine.failIDs[ref.ID]; err != nil {
		return runtimeapi.ActionResult{}, err
	}
	return runtimeapi.ActionResult{Resource: ref}, nil
}

// volumeService is a stub that honors Fail / FailAll for the IDs passed
// to Remove. Other methods are no-op stubs.
type volumeService struct{ engine *Engine }

func (v *volumeService) Remove(_ context.Context, name string, _ bool) error {
	v.engine.mu.Lock()
	defer v.engine.mu.Unlock()
	v.engine.removeCalls = append(v.engine.removeCalls, name)
	if err := v.engine.failAll; err != nil {
		return err
	}
	return v.engine.failIDs[name]
}

func (v *volumeService) List(context.Context, runtimeapi.VolumeListOptions) ([]runtimeapi.Volume, error) {
	return nil, nil
}
func (v *volumeService) Inspect(context.Context, string) (*runtimeapi.VolumeDetail, error) {
	return &runtimeapi.VolumeDetail{}, nil
}
func (v *volumeService) Create(context.Context, runtimeapi.VolumeCreateOptions) (*runtimeapi.Volume, error) {
	return &runtimeapi.Volume{}, nil
}
func (v *volumeService) Prune(context.Context, runtimeapi.PruneOptions) (runtimeapi.PruneResult, error) {
	return runtimeapi.PruneResult{}, nil
}

// networkService mirrors volumeService for networks.
type networkService struct{ engine *Engine }

func (n *networkService) Remove(_ context.Context, id string) error {
	n.engine.mu.Lock()
	defer n.engine.mu.Unlock()
	n.engine.removeCalls = append(n.engine.removeCalls, id)
	if err := n.engine.failAll; err != nil {
		return err
	}
	return n.engine.failIDs[id]
}

func (n *networkService) List(context.Context, runtimeapi.NetworkListOptions) ([]runtimeapi.Network, error) {
	return nil, nil
}
func (n *networkService) Inspect(context.Context, string) (*runtimeapi.NetworkDetail, error) {
	return &runtimeapi.NetworkDetail{}, nil
}
func (n *networkService) Create(context.Context, runtimeapi.NetworkCreateOptions) (*runtimeapi.Network, error) {
	return &runtimeapi.Network{}, nil
}
func (n *networkService) Prune(context.Context, runtimeapi.PruneOptions) (runtimeapi.PruneResult, error) {
	return runtimeapi.PruneResult{}, nil
}

// Compile-time check.
var _ runtimeapi.Engine = (*Engine)(nil)

// Compile-time check for actions package to verify the error type.
var _ = errors.New
