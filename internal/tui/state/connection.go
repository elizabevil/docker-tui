package state

import dockerclient "github.com/elizabevil/docker-tui/internal/data/docker"

// ConnectionState owns the active runtime, selector and health state.
type ConnectionState struct {
	Docker                  *dockerclient.Client
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

func NewConnectionState(client *dockerclient.Client) ConnectionState {
	state := ConnectionState{Docker: client, Connected: client != nil}
	if client != nil {
		state.RuntimeType = string(client.RuntimeType)
		state.EngineVersion = client.EngineVersion
	}
	return state
}

func (s *ConnectionState) Begin() {
	s.Connecting = true
}

func (s *ConnectionState) ConnectedTo(name string, client *dockerclient.Client) {
	s.Docker = client
	s.Connecting = false
	s.Connected = client != nil
	s.ConnectionTarget = name
	s.ConnectionFailure = dockerclient.ConnectionFailure{}
	s.HealthFailures = 0
	s.HealthDegraded = false
	if s.RuntimeSelectorError != nil {
		delete(s.RuntimeSelectorError, name)
	}
	if client == nil {
		s.RuntimeType = ""
		s.EngineVersion = ""
		return
	}
	s.RuntimeType = string(client.RuntimeType)
	s.EngineVersion = client.EngineVersion
}

func (s *ConnectionState) Failed(name string, err error) {
	s.Docker = nil
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
