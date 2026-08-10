package docker

import (
	"encoding/json"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

func TestListContainersPopulatesNetworksAndIPs(t *testing.T) {
	raw := container.Summary{
		ID:    "10b2c97269cc",
		Names: []string{"/redis"},
		Image: "redis:7-alpine",
		State: "running",
		NetworkSettings: &container.NetworkSettingsSummary{
			Networks: map[string]*network.EndpointSettings{
				"bridge": {IPAddress: "172.17.0.42"},
				"myNet":  {IPAddress: "10.0.0.2"},
			},
		},
	}
	payload, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	var decoded container.Summary
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatal(err)
	}
	got := runtimeapi.ContainerSummary{
		Networks: map[string]string{},
	}
	if decoded.NetworkSettings != nil {
		for name, net := range decoded.NetworkSettings.Networks {
			if net.IPAddress != "" {
				got.Networks[name] = net.IPAddress
				got.IPs = append(got.IPs, net.IPAddress)
			}
		}
	}
	if got.Networks["bridge"] != "172.17.0.42" || got.Networks["myNet"] != "10.0.0.2" {
		t.Fatalf("networks map missing entries: %+v", got.Networks)
	}
	if len(got.IPs) != 2 {
		t.Fatalf("IPs slice missing entries: %+v", got.IPs)
	}
}
