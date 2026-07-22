package runtime

// NetworkDetail holds the structured inspect result for a network.
// Both Docker and Podman adapters map their native inspect responses
// into this type, ensuring the UI never parses raw SDK JSON.
type NetworkDetail struct {
	Name       string                     `json:"Name"`
	ID         string                     `json:"Id"`
	Created    string                     `json:"Created"`
	Scope      string                     `json:"Scope"`
	Driver     string                     `json:"Driver"`
	EnableIPv4 bool                       `json:"EnableIPv4"`
	EnableIPv6 bool                       `json:"EnableIPv6"`
	IPAM       NetworkIPAM                `json:"IPAM"`
	Internal   bool                       `json:"Internal"`
	Containers map[string]NetworkEndpoint `json:"Containers"`
	Options    map[string]string          `json:"Options"`
	Labels     map[string]string          `json:"Labels"`
}

// NetworkIPAM holds IP Address Management configuration for a network.
type NetworkIPAM struct {
	Config []NetworkIPAMConfig `json:"Config"`
}

// NetworkIPAMConfig holds a single IPAM configuration entry.
type NetworkIPAMConfig struct {
	Subnet     string            `json:"Subnet"`
	Gateway    string            `json:"Gateway,omitempty"`
	IPRange    string            `json:"IPRange,omitempty"`
	AuxAddress map[string]string `json:"AuxAddress,omitempty"`
}

// NetworkEndpoint holds information about a container connected to a network.
type NetworkEndpoint struct {
	Name        string `json:"Name"`
	EndpointID  string `json:"EndpointID"`
	MacAddress  string `json:"MacAddress"`
	IPv4Address string `json:"IPv4Address"`
	IPv6Address string `json:"IPv6Address"`
	Gateway     string `json:"Gateway,omitempty"`
}
