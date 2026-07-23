package dto

import (
	"encoding/json"
	"net"
	"time"
)

// IPNet mirrors go.podman.io/common/libnetwork/types.IPNet. It carries the
// same JSON text encoding so REST responses can be unmarshalled without the
// CGO dependency.
type IPNet struct {
	net.IPNet
}

// UnmarshalJSON accepts either a CIDR string ("192.168.0.0/24") or a
// raw JSON object form (the libpod response uses text encoding).
func (n *IPNet) UnmarshalJSON(data []byte) error {
	trimmed := trimJSON(data)
	if len(trimmed) == 0 || string(trimmed) == "null" {
		return nil
	}
	if trimmed[0] == '"' {
		var s string
		if err := json.Unmarshal(trimmed, &s); err != nil {
			return err
		}
		ip, ipnet, err := net.ParseCIDR(s)
		if err != nil {
			return err
		}
		if ip != nil {
			ipnet.IP = ip
		}
		n.IPNet = *ipnet
		return nil
	}
	var inner net.IPNet
	if err := json.Unmarshal(data, &inner); err != nil {
		return err
	}
	n.IPNet = inner
	return nil
}

// MarshalJSON emits the IPNet as a CIDR string.
func (n IPNet) MarshalJSON() ([]byte, error) {
	return json.Marshal(n.String())
}

// HardwareAddr mirrors go.podman.io/common/libnetwork/types.HardwareAddr
// so MAC addresses in REST responses are accepted.
type HardwareAddr net.HardwareAddr

// UnmarshalJSON accepts either a hex string ("aa:bb:cc:...") or a JSON
// array of bytes (fallback).
func (h *HardwareAddr) UnmarshalJSON(data []byte) error {
	trimmed := trimJSON(data)
	if len(trimmed) == 0 || string(trimmed) == "null" {
		return nil
	}
	if trimmed[0] == '"' {
		var s string
		if err := json.Unmarshal(trimmed, &s); err != nil {
			return err
		}
		mac, err := net.ParseMAC(s)
		if err != nil {
			return err
		}
		*h = HardwareAddr(mac)
		return nil
	}
	var arr []byte
	if err := json.Unmarshal(data, &arr); err != nil {
		return err
	}
	*h = HardwareAddr(arr)
	return nil
}

// MarshalJSON emits the MAC address as a colon-separated hex string.
func (h HardwareAddr) MarshalJSON() ([]byte, error) {
	return json.Marshal(net.HardwareAddr(h).String())
}

// String mirrors net.HardwareAddr.String.
func (h HardwareAddr) String() string {
	return net.HardwareAddr(h).String()
}

// Subnet is the Podman subnet entry inside Network.Subnets.
type Subnet struct {
	Subnet     IPNet       `json:"subnet"`
	Gateway    net.IP      `json:"gateway,omitempty"`
	LeaseRange *LeaseRange `json:"lease_range,omitempty"`
}

// LeaseRange is the Podman lease range entry inside Subnet.LeaseRange.
type LeaseRange struct {
	StartIP net.IP `json:"start_ip,omitempty"`
	EndIP   net.IP `json:"end_ip,omitempty"`
}

// RouteType is the Podman route type enum; mirrored for JSON encoding.
type RouteType int

// Route is the Podman route entry inside Network.Routes.
type Route struct {
	Destination IPNet     `json:"destination"`
	Gateway     net.IP    `json:"gateway,omitempty"`
	Metric      *uint32   `json:"metric,omitempty"`
	RouteType   RouteType `json:"route_type,omitempty"`
}

// Network is the Podman Libpod network list response item.
// Field shape and JSON tags match go.podman.io/common/libnetwork/types.Network.
type Network struct {
	Name              string            `json:"name"`
	ID                string            `json:"id"`
	Driver            string            `json:"driver"`
	Scope             string            `json:"scope,omitempty"`
	NetworkInterface  string            `json:"network_interface,omitempty"`
	Created           time.Time         `json:"created"`
	Subnets           []Subnet          `json:"subnets,omitempty"`
	Routes            []Route           `json:"routes,omitempty"`
	IPv6Enabled       bool              `json:"ipv6_enabled"`
	Internal          bool              `json:"internal"`
	DNSEnabled        bool              `json:"dns_enabled"`
	NetworkDNSServers []string          `json:"network_dns_servers,omitempty"`
	Labels            map[string]string `json:"labels,omitempty"`
	Options           map[string]string `json:"options,omitempty"`
	IPAMOptions       map[string]string `json:"ipam_options,omitempty"`
}

// NetworkContainerInfo is per-container info in NetworkInspect.Containers.
type NetworkContainerInfo struct {
	Name       string                  `json:"name"`
	Interfaces map[string]NetInterface `json:"interfaces,omitempty"`
}

// NetInterface is per-interface info in NetworkContainerInfo.Interfaces.
type NetInterface struct {
	Subnets    []NetAddress `json:"subnets,omitempty"`
	MacAddress HardwareAddr `json:"mac_address"`
}

// NetAddress is the Podman NetAddress DTO.
type NetAddress struct {
	IPNet   IPNet  `json:"ipnet"`
	Gateway net.IP `json:"gateway,omitempty"`
}

// NetworkInspect embeds Network and adds the per-container map.
type NetworkInspect struct {
	Network
	Containers map[string]NetworkContainerInfo `json:"containers"`
}

func trimJSON(b []byte) []byte {
	start := 0
	end := len(b)
	for start < end {
		switch b[start] {
		case ' ', '\t', '\r', '\n':
			start++
		default:
			goto trimEnd
		}
	}
trimEnd:
	for end > start {
		switch b[end-1] {
		case ' ', '\t', '\r', '\n':
			end--
		default:
			return b[start:end]
		}
	}
	return b[start:end]
}
