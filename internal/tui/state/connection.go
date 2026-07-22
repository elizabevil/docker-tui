package state

import (
	dockerclient "github.com/elizabevil/docker-tui/internal/data/docker"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

// ConnectionState owns the active runtime, selector and health state.
type ConnectionState struct {
	Engine                  runtimeapi.Engine
	Pool                    *dockerclient.ConnectionPool
	Connecting              bool
	Connected               bool
	ConnectionTarget        string
	ConnectionFailure       dockerclient.ConnectionFailure
	RuntimeSelectorCursor   int
	RuntimeSelectorError    map[string]dockerclient.ConnectionFailure
	RuntimeSelectorDisabled bool
	HealthFailures          int
	HealthDegraded          bool
	RuntimeType             string
	EngineVersion           string
}

func NewConnectionState(engine runtimeapi.Engine) ConnectionState {
	state := ConnectionState{Engine: engine, Connected: engine != nil}
	if engine != nil {
		identity := engine.Identity()
		state.RuntimeType = string(identity.Type)
		state.EngineVersion = identity.Version
	}
	return state
}

func (s *ConnectionState) Begin() {
	s.Connecting = true
}

func (s *ConnectionState) ConnectedTo(name string, engine runtimeapi.Engine) {
	s.Engine = engine
	s.Connecting = false
	s.Connected = engine != nil
	s.ConnectionTarget = name
	s.ConnectionFailure = dockerclient.ConnectionFailure{}
	s.HealthFailures = 0
	s.HealthDegraded = false
	if s.RuntimeSelectorError != nil {
		delete(s.RuntimeSelectorError, name)
	}
	if engine == nil {
		s.RuntimeType = ""
		s.EngineVersion = ""
		return
	}
	identity := engine.Identity()
	s.RuntimeType = string(identity.Type)
	s.EngineVersion = identity.Version
}

func (s *ConnectionState) Failed(name string, err error) {
	s.Engine = nil
	s.Connecting = false
	s.Connected = false
	s.ConnectionTarget = name
	s.ConnectionFailure = dockerclient.ClassifyConnectionError(err)
	s.RuntimeType = ""
	s.EngineVersion = ""
	s.HealthFailures = 0
	s.HealthDegraded = false
}

func (s *ConnectionState) SelectionFailed(name string, err error) {
	s.Connecting = false
	s.SetProbeResult(name, err)
}

func (s *ConnectionState) SetProbeResult(name string, err error) {
	if s.RuntimeSelectorError == nil {
		s.RuntimeSelectorError = make(map[string]dockerclient.ConnectionFailure)
	}
	if err == nil {
		delete(s.RuntimeSelectorError, name)
		return
	}
	s.RuntimeSelectorError[name] = dockerclient.ClassifyConnectionError(err)
}
