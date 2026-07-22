//go:build !cgo

package podman

import "github.com/elizabevil/docker-tui/internal/data/runtime/podman/dto"

// Network is an alias for dto.Network.
type Network = dto.Network

// NetworkSubnet is an alias for dto.Subnet.
type NetworkSubnet = dto.Subnet

// NetInterface is an alias for dto.NetInterface.
type NetInterface = dto.NetInterface

// NetAddress is an alias for dto.NetAddress.
type NetAddress = dto.NetAddress

// NetworkInspectItem is an alias for dto.NetworkInspect.
type NetworkInspectItem = dto.NetworkInspect

// NetworkContainerInfo is an alias for dto.NetworkContainerInfo.
type NetworkContainerInfo = dto.NetworkContainerInfo
