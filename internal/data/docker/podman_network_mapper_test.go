package docker

import (
	"testing"
	"time"
)

func TestMapPodmanNetworks(t *testing.T) {
	created := time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC)
	result := mapPodmanNetworks([]podmanNetworkItem{{
		Name:    "podman",
		ID:      "abc123",
		Driver:  "bridge",
		Created: created,
		Subnets: []podmanSubnet{
			{Subnet: "10.88.0.0/16", Gateway: "10.88.0.1"},
		},
		Internal: false,
		Labels:   map[string]string{"env": "dev"},
	}})
	if len(result) != 1 {
		t.Fatalf("expected one network, got %d", len(result))
	}
	net := result[0]
	if net.Name != "podman" || net.ID != "abc123" {
		t.Errorf("name/id = %q/%q", net.Name, net.ID)
	}
	if net.Scope != "local" {
		t.Errorf("scope = %q, want local", net.Scope)
	}
	if net.Driver != "bridge" {
		t.Errorf("driver = %q", net.Driver)
	}
	if len(net.IPAM) != 1 || net.IPAM[0] != "10.88.0.0/16" {
		t.Errorf("IPAM = %v", net.IPAM)
	}
	if net.Created != created.Unix() {
		t.Errorf("created = %d, want %d", net.Created, created.Unix())
	}
}

func TestMapPodmanNetworkInspect(t *testing.T) {
	created := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	item := podmanNetworkItem{
		Name:        "test-net",
		ID:          "def456",
		Driver:      "bridge",
		Created:     created,
		IPv6Enabled: true,
		Internal:    true,
		Subnets: []podmanSubnet{
			{Subnet: "172.20.0.0/16", Gateway: "172.20.0.1"},
		},
		Options: map[string]string{"com.docker.network.bridge.name": "br0"},
		Labels:  map[string]string{"tier": "backend"},
		Containers: map[string]podmanNetworkContainer{
			"ep1": {
				Name: "c1",
				Interfaces: map[string]podmanNetInterface{
					"eth0": {
						Subnets: []podmanNetAddress{
							{IPNet: "172.20.0.2/16", Gateway: "172.20.0.1"},
						},
						MacAddress: "aa:bb:cc:dd:ee:ff",
					},
				},
			},
		},
	}
	result := mapPodmanNetworkInspect(item)
	if result.Name != "test-net" || result.ID != "def456" {
		t.Errorf("name/id = %q/%q", result.Name, result.ID)
	}
	if !result.EnableIPv6 {
		t.Error("expected EnableIPv6=true")
	}
	if !result.EnableIPv4 {
		t.Error("expected IPv4 subnet to keep EnableIPv4=true on a dual-stack network")
	}
	if !result.Internal {
		t.Error("expected Internal=true")
	}
	if len(result.IPAM.Config) != 1 || result.IPAM.Config[0].Subnet != "172.20.0.0/16" {
		t.Errorf("IPAM.Config = %v", result.IPAM.Config)
	}
	ep, ok := result.Containers["ep1"]
	if !ok {
		t.Fatal("missing container endpoint")
	}
	if ep.Name != "c1" || ep.MacAddress != "aa:bb:cc:dd:ee:ff" {
		t.Errorf("endpoint = %+v", ep)
	}
	if ep.IPv4Address != "172.20.0.2/16" {
		t.Errorf("IPv4Address = %q", ep.IPv4Address)
	}
}

func TestMapPodmanNetworksEmpty(t *testing.T) {
	result := mapPodmanNetworks([]podmanNetworkItem{})
	if len(result) != 0 {
		t.Errorf("expected empty, got %d", len(result))
	}
}
