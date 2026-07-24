package state

import (
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

// ConnectionState owns the active runtime, selector and health state.
type ConnectionState struct {
	Engine                  runtimeapi.Engine
	Pool                    *runtimeapi.ConnectionPool
	Connecting              bool
	Connected               bool
	ConnectionTarget        string
	ConnectionFailure       runtimeapi.ConnectionFailure
	RuntimeSelectorCursor   int
	RuntimeSelectorError    map[string]runtimeapi.ConnectionFailure
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
	s.ConnectionFailure = runtimeapi.ConnectionFailure{}
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
	s.ConnectionFailure = runtimeapi.ClassifyConnectionError(err)
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
		s.RuntimeSelectorError = make(map[string]runtimeapi.ConnectionFailure)
	}
	if err == nil {
		delete(s.RuntimeSelectorError, name)
		return
	}
	s.RuntimeSelectorError[name] = runtimeapi.ClassifyConnectionError(err)
}
