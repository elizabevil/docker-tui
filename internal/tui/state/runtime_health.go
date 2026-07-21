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
func (m *AppModel) ApplyHealthResult(name string, err error, threshold int) HealthTransition {
	if m == nil || name != m.ConnectionTarget || m.Docker == nil {
		return HealthNoChange
	}
	if threshold <= 0 {
		threshold = 2
	}
	if err != nil {
		m.HealthFailures++
		if m.HealthFailures >= threshold && !m.HealthDegraded {
			m.HealthDegraded = true
			m.Connected = false
			m.ConnectionError = err.Error()
			return HealthDisconnected
		}
		return HealthNoChange
	}
	wasDegraded := m.HealthDegraded
	m.HealthFailures = 0
	m.HealthDegraded = false
	if wasDegraded {
		m.Connected = true
		m.ConnectionError = ""
		return HealthRecovered
	}
	return HealthNoChange
}
