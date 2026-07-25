package runtime

// NetworkDetail holds the structured inspect result for a network.
// Both Docker and Podman adapters map their native inspect responses
// into this type, ensuring the UI never parses raw SDK JSON.
type NetworkDetail struct {
	Name       string
	ID         string
	Created    string
	Scope      string
	Driver     string
	EnableIPv4 bool
	EnableIPv6 bool
	IPAM       NetworkIPAM
	Internal   bool
	Containers map[string]NetworkEndpoint
	Options    map[string]string
	Labels     map[string]string
}

// NetworkIPAM holds IP Address Management configuration for a network.
type NetworkIPAM struct {
	Config []NetworkIPAMConfig
}

// NetworkIPAMConfig holds a single IPAM configuration entry.
type NetworkIPAMConfig struct {
	Subnet     string
	Gateway    string
	IPRange    string
	AuxAddress map[string]string
}

// NetworkEndpoint holds information about a container connected to a network.
type NetworkEndpoint struct {
	Name        string
	EndpointID  string
	MacAddress  string
	IPv4Address string
	IPv6Address string
	Gateway     string
}
