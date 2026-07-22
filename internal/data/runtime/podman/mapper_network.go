package podman

import (
	"net"
	"time"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

// MapNetworks converts Podman network list results into runtime.Network items.
// Field differences: Subnets (commonTypes.Subnet with IPNet/net.IP) are
// extracted as CIDR strings; Created (time.Time) is converted to Unix
// timestamp; Scope defaults to "local".
// Note: Podman list response does not include a containers field.
// Container count is only available via the inspect endpoint.
func MapNetworks(raw []Network) []runtimeapi.Network {
	result := make([]runtimeapi.Network, 0, len(raw))
	for _, n := range raw {
		ipam := make([]string, 0, len(n.Subnets))
		for _, s := range n.Subnets {
			cidr := s.Subnet.String()
			if cidr != "" {
				ipam = append(ipam, cidr)
			}
		}
		result = append(result, runtimeapi.Network{
			Name:     n.Name,
			ID:       n.ID,
			Driver:   n.Driver,
			Scope:    "local",
			IPAM:     ipam,
			Created:  n.Created.Unix(),
			Internal: n.Internal,
			Labels:   n.Labels,
		})
	}
	return result
}

// MapNetworkInspect converts a single Podman network inspect result
// into the structured NetworkDetail domain type for the detail view.
func MapNetworkInspect(raw NetworkInspectItem) *runtimeapi.NetworkDetail {
	detailContainers := make(map[string]runtimeapi.NetworkEndpoint, len(raw.Containers))
	for id, c := range raw.Containers {
		ep := runtimeapi.NetworkEndpoint{Name: c.Name}
		for _, iface := range c.Interfaces {
			ep.MacAddress = iface.MacAddress.String()
			if len(iface.Subnets) > 0 {
				ep.IPv4Address = iface.Subnets[0].IPNet.String()
				ep.Gateway = netIPString(iface.Subnets[0].Gateway)
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
			Subnet:  s.Subnet.String(),
			Gateway: netIPString(s.Gateway),
		})
		if ip, _, err := net.ParseCIDR(s.Subnet.String()); err == nil {
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

// netIPString converts a net.IP to a string, returning "" for nil.
func netIPString(ip net.IP) string {
	if ip == nil {
		return ""
	}
	return ip.String()
}
