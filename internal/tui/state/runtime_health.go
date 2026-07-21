package state

// HealthTransition describes a meaningful runtime health state change.
type HealthTransition int

const (
	HealthNoChange HealthTransition = iota
	HealthDisconnected
	HealthRecovered
)

// ApplyHealthResult updates health counters and returns only state transitions;
// presentation decisions remain in the TUI update layer.
func (s *ConnectionState) ApplyHealthResult(name string, err error, threshold int) HealthTransition {
	if s == nil || name != s.ConnectionTarget || s.Docker == nil {
		return HealthNoChange
	}
	if threshold <= 0 {
		threshold = 2
	}
	if err != nil {
		s.HealthFailures++
		if s.HealthFailures >= threshold && !s.HealthDegraded {
			s.HealthDegraded = true
			s.Connected = false
			s.ConnectionError = err.Error()
			return HealthDisconnected
		}
		return HealthNoChange
	}
	wasDegraded := s.HealthDegraded
	s.HealthFailures = 0
	s.HealthDegraded = false
	if wasDegraded {
		s.Connected = true
		s.ConnectionError = ""
		return HealthRecovered
	}
	return HealthNoChange
}
