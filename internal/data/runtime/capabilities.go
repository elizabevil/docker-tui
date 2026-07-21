package runtime

// Support describes how completely a driver implements a capability.
type Support uint8

const (
	Unsupported Support = iota
	Degraded
	Available
)

// Capability is a stable feature identifier used by business and UI layers.
type Capability string

const (
	CapabilityContainers Capability = "containers"
	CapabilityImages     Capability = "images"
	CapabilityVolumes    Capability = "volumes"
	CapabilityNetworks   Capability = "networks"
	CapabilityEvents     Capability = "events"
	CapabilityExec       Capability = "exec"
)

type CapabilityInfo struct {
	Support Support
	Reason  string
}

type CapabilitySet map[Capability]CapabilityInfo

func (s CapabilitySet) SupportFor(capability Capability) Support {
	info, ok := s[capability]
	if !ok {
		return Unsupported
	}
	return info.Support
}

func (s CapabilitySet) Supports(capability Capability) bool {
	return s.SupportFor(capability) != Unsupported
}
