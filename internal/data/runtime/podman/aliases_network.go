//go:build cgo

package podman

// Type aliases to Podman native types.
// The HTTP REST JSON responses use the same field names and types as these
// Go structs, so they can be deserialized directly without custom DTOs.
//
// Aliases avoid duplication: the CGO bindings path and the REST path share
// the same struct definitions from the Podman driver.

import (
	commonTypes "go.podman.io/common/libnetwork/types"
	"go.podman.io/podman/v6/pkg/domain/entities/types"
)

// Network is an alias for the Podman network list DTO.
// Used for list and create responses (no Containers field).
// Verified compatible with Podman 5.4.2 /networks/json response.
type Network = commonTypes.Network

// NetworkSubnet is an alias for the Podman network subnet DTO.
type NetworkSubnet = commonTypes.Subnet

// NetInterface is an alias for the Podman network interface DTO.
type NetInterface = commonTypes.NetInterface

// NetAddress is an alias for the Podman network address DTO.
type NetAddress = commonTypes.NetAddress

// NetworkInspectItem is the Podman network inspect response.
// It wraps Network and adds a Containers map.
type NetworkInspectItem = types.NetworkInspectReport

// NetworkContainerInfo holds per-container endpoint info on a network.
type NetworkContainerInfo = types.NetworkContainerInfo
