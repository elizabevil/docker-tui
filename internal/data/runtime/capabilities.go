package runtime

// Support describes how completely a driver implements a capability.
type Support uint8

// Support levels describe the degree of implementation for a capability.
const (
	Unsupported Support = iota
	Degraded
	Available
)

// Capability is a stable feature identifier used by business and UI layers.
type Capability string

// Well-known capability identifiers. Aggregate capabilities (Containers,
// Images, etc.) remain during the facade migration; fine-grained capabilities
// (ContainerListFilter, etc.) replace them as adapters are migrated.
const (
	// Legacy aggregate capabilities remain during facade migration.
	CapabilityContainers Capability = "containers"
	CapabilityImages     Capability = "images"
	CapabilityVolumes    Capability = "volumes"
	CapabilityNetworks   Capability = "networks"
	CapabilityEvents     Capability = "events"
	CapabilityExec       Capability = "exec"
	CapabilityFiltering  Capability = "native_filtering"

	CapabilityContainerListFilter Capability = "container.list.filter"
	CapabilityImageListFilter     Capability = "image.list.filter"
	CapabilityVolumeListFilter    Capability = "volume.list.filter"
	CapabilityNetworkListFilter   Capability = "network.list.filter"
	CapabilityEventFilter         Capability = "events.filter"
	CapabilityExecResize          Capability = "exec.resize"

	CapabilityComposeAggregate    Capability = "compose.aggregate"
	CapabilityComposeProjectLevel Capability = "compose.project_level"
	CapabilityComposeUp           Capability = "compose.up"
	CapabilityComposeBuild        Capability = "compose.build"
	CapabilityComposeRun          Capability = "compose.run"
	CapabilityComposeConfig       Capability = "compose.config"
	CapabilityComposePull         Capability = "compose.pull"
	CapabilityComposePush         Capability = "compose.push"
	CapabilityComposePodScope     Capability = "compose.pod_scope"
	CapabilityComposeCoLocated    Capability = "compose.co_located"
	CapabilityStatsStream         Capability = "stats.stream"
)

// CapabilityInfo records the support level and human-readable reason for a
// single capability. The ReasonCode is a machine-readable identifier for
// structured error projection.
type CapabilityInfo struct {
	Support    Support
	Reason     string
	ReasonCode string
}

// CapabilitySet is a snapshot of all capabilities reported by an engine.
type CapabilitySet map[Capability]CapabilityInfo

// SupportFor returns the support level for the given capability. Returns
// Unsupported when the capability is not present in the set.
func (s CapabilitySet) SupportFor(capability Capability) Support {
	info, ok := s[capability]
	if !ok {
		return Unsupported
	}
	return info.Support
}

// Supports reports whether the capability is available or better.
func (s CapabilitySet) Supports(capability Capability) bool {
	return s.SupportFor(capability) != Unsupported
}
