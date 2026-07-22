package docker

import (
	"net"
	"time"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

// podmanNetworkItem is the adapter's stable representation of the Libpod
// network list/inspect response. Both CGO bindings and non-CGO REST normalize
// into this type before mapping to runtime.Network or runtime.NetworkDetail.
//
// Podman's native type (commonTypes.Network) uses Subnets []Subnet with
// net.IPNet encoding and stores IPAM options as map[string]string, whereas
// Docker uses IPAM.Config[].Subnet as plain strings. The DTO normalizes these
// differences before the mapper produces the domain model.
type podmanNetworkItem struct {
	Name        string                            `json:"name"`
	ID          string                            `json:"id"`
	Driver      string                            `json:"driver"`
	Created     time.Time                         `json:"created"`
	Subnets     []podmanSubnet                    `json:"subnets,omitempty"`
	IPv6Enabled bool                              `json:"ipv6_enabled"`
	Internal    bool                              `json:"internal"`
	Labels      map[string]string                 `json:"labels,omitempty"`
	Options     map[string]string                 `json:"options,omitempty"`
	Containers  map[string]podmanNetworkContainer `json:"containers,omitempty"`
}

// podmanSubnet is a normalized subnet entry within a Podman network item.
type podmanSubnet struct {
	Subnet  string `json:"subnet"`
	Gateway string `json:"gateway,omitempty"`
}

// podmanNetworkContainer represents a container connected to a Podman network.
type podmanNetworkContainer struct {
	Name       string                        `json:"name"`
	Interfaces map[string]podmanNetInterface `json:"interfaces,omitempty"`
}

// podmanNetInterface is a single network interface on a container.
type podmanNetInterface struct {
	Subnets    []podmanNetAddress `json:"subnets,omitempty"`
	MacAddress string             `json:"mac_address"`
}

// podmanNetAddress holds a single IP address assignment on an interface.
type podmanNetAddress struct {
	IPNet   string `json:"ipnet"`
	Gateway string `json:"gateway,omitempty"`
}

// mapPodmanNetworks converts Podman network list results into runtime.Network
// items. Field differences: Subnets are extracted as CIDR strings for IPAM;
// Containers count is not populated in list responses (always 0); Created
// (time.Time) is converted to Unix timestamp; Scope is absent from Podman
// and defaults to "local".
func mapPodmanNetworks(raw []podmanNetworkItem) []runtimeapi.Network {
	result := make([]runtimeapi.Network, 0, len(raw))
	for _, n := range raw {
		ipam := make([]string, 0, len(n.Subnets))
		for _, s := range n.Subnets {
			if s.Subnet != "" {
				ipam = append(ipam, s.Subnet)
			}
		}
		result = append(result, runtimeapi.Network{
			Name:       n.Name,
			ID:         n.ID,
			Driver:     n.Driver,
			Scope:      "local",
			IPAM:       ipam,
			Containers: len(n.Containers),
			Created:    n.Created.Unix(),
			Internal:   n.Internal,
			Labels:     n.Labels,
		})
	}
	return result
}

// mapPodmanNetworkInspect converts a single Podman network inspect result
// into the structured NetworkDetail domain type for the detail view.
func mapPodmanNetworkInspect(raw podmanNetworkItem) *runtimeapi.NetworkDetail {
	detailContainers := make(map[string]runtimeapi.NetworkEndpoint, len(raw.Containers))
	for id, c := range raw.Containers {
		ep := runtimeapi.NetworkEndpoint{Name: c.Name}
		for _, iface := range c.Interfaces {
			ep.MacAddress = iface.MacAddress
			if len(iface.Subnets) > 0 {
				ep.IPv4Address = iface.Subnets[0].IPNet
				ep.Gateway = iface.Subnets[0].Gateway
			}
			break
		}
		detailContainers[id] = ep
	}

	ipamConfigs := make([]runtimeapi.NetworkIPAMConfig, 0, len(raw.Subnets))
	enableIPv4 := false
	enableIPv6 := raw.IPv6Enabled
	for _, s := range raw.Subnets {
		ipamConfigs = append(ipamConfigs, runtimeapi.NetworkIPAMConfig{
			Subnet:  s.Subnet,
			Gateway: s.Gateway,
		})
		if ip, _, err := net.ParseCIDR(s.Subnet); err == nil {
			if ip.To4() != nil {
				enableIPv4 = true
			} else {
				enableIPv6 = true
			}
		}
	}

	return &runtimeapi.NetworkDetail{
		Name:       raw.Name,
		ID:         raw.ID,
		Created:    raw.Created.Format(time.RFC3339),
		Scope:      "local",
		Driver:     raw.Driver,
		EnableIPv4: enableIPv4,
		EnableIPv6: enableIPv6,
		IPAM: runtimeapi.NetworkIPAM{
			Config: ipamConfigs,
		},
		Internal:   raw.Internal,
		Containers: detailContainers,
		Options:    raw.Options,
		Labels:     raw.Labels,
	}
}
