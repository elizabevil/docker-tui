package runtime

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestCapabilitySetDefaultsToUnsupported(t *testing.T) {
	capabilities := CapabilitySet{
		CapabilityContainers: {Support: Available},
		CapabilityEvents:     {Support: Degraded, Reason: "polling fallback"},
	}

	if !capabilities.Supports(CapabilityContainers) {
		t.Fatal("available capability should be supported")
	}
	if !capabilities.Supports(CapabilityEvents) {
		t.Fatal("degraded capability should be supported")
	}
	if capabilities.Supports(CapabilityExec) {
		t.Fatal("missing capability should be unsupported")
	}
}

func TestErrorPreservesKindAndCause(t *testing.T) {
	cause := errors.New("socket unavailable")
	err := NewError(ErrorConnection, "ping", "local-podman", cause)

	if !IsErrorKind(err, ErrorConnection) {
		t.Fatalf("expected connection error, got %v", err)
	}
	if !errors.Is(err, cause) {
		t.Fatal("runtime error should preserve its cause")
	}
	if got := err.Error(); got != "ping: connection local-podman: socket unavailable" {
		t.Fatalf("unexpected error text: %q", got)
	}
}

func TestFilterSetCloneDoesNotShareValues(t *testing.T) {
	filters := FilterSet{"label": {"app=api"}}
	clone := filters.Clone()
	clone.Add("label", "tier=backend")

	if got := len(filters.Values("label")); got != 1 {
		t.Fatalf("clone mutated source values: got %d", got)
	}
	if got := len(clone.Values("label")); got != 2 {
		t.Fatalf("expected two cloned values, got %d", got)
	}
}

func TestHTTPErrorClassification(t *testing.T) {
	tests := map[int]ErrorKind{
		400: ErrorInvalid,
		401: ErrorAuthentication,
		403: ErrorPermission,
		404: ErrorNotFound,
		409: ErrorConflict,
		429: ErrorRateLimited,
		503: ErrorUnavailable,
		500: ErrorInternal,
	}
	for status, want := range tests {
		if got := ClassifyHTTPStatus(status); got != want {
			t.Errorf("status %d: got %q, want %q", status, got, want)
		}
	}
}

func TestRetryableErrorKinds(t *testing.T) {
	for _, kind := range []ErrorKind{ErrorConnection, ErrorTimeout, ErrorRateLimited, ErrorUnavailable} {
		if !IsRetryableKind(kind) {
			t.Errorf("expected %q to be retryable", kind)
		}
	}
	for _, kind := range []ErrorKind{ErrorCanceled, ErrorPermission, ErrorInvalid, ErrorConflict, ErrorUnsupported} {
		if IsRetryableKind(kind) {
			t.Errorf("expected %q not to be retryable", kind)
		}
	}
}

func TestPruneResultCountsPartialFailures(t *testing.T) {
	result := PruneResult{Resources: []ResourceResult{{ID: "deleted"}, {ID: "busy", Error: errors.New("in use")}}}
	succeeded, failed := result.Counts()
	if succeeded != 1 || failed != 1 {
		t.Fatalf("Counts() = %d, %d", succeeded, failed)
	}
}

func TestImageListOptionsPushesDownFirstANDValue(t *testing.T) {
	options := ImageListOptions{Filters: FilterSet{ImageFilterLabel: {"app=api", "tier=backend"}}}
	filters, err := options.NativeFilters()
	if err != nil || len(filters[ImageFilterLabel]) != 1 || filters[ImageFilterLabel][0] != "app=api" {
		t.Fatalf("native filters = %#v, err=%v", filters, err)
	}
}

func TestImageListOptionsCopiesNativeFilters(t *testing.T) {
	options := ImageListOptions{Filters: FilterSet{ImageFilterReference: {"example/*"}}}
	filters, err := options.NativeFilters()
	if err != nil {
		t.Fatalf("NativeFilters() error = %v", err)
	}
	filters[ImageFilterReference][0] = "changed"
	if options.Filters.Values(ImageFilterReference)[0] != "example/*" {
		t.Fatal("native filters mutated input options")
	}
}

func TestContainerListOptionsPushesDownFirstANDValue(t *testing.T) {
	options := ContainerListOptions{Filters: FilterSet{ContainerFilterLabel: {"app=api", "tier=backend"}}}
	filters, err := options.NativeFilters()
	if err != nil || len(filters[ContainerFilterLabel]) != 1 || filters[ContainerFilterLabel][0] != "app=api" {
		t.Fatalf("native filters = %#v, err=%v", filters, err)
	}
}

func TestResourceListOptionsPushDownFirstANDValue(t *testing.T) {
	volume, volumeErr := (VolumeListOptions{Filters: FilterSet{VolumeFilterLabel: {"a=1", "b=2"}}}).NativeFilters()
	network, networkErr := (NetworkListOptions{Filters: FilterSet{NetworkFilterName: {"a", "b"}}}).NativeFilters()
	if volumeErr != nil || networkErr != nil || len(volume[VolumeFilterLabel]) != 1 || len(network[NetworkFilterName]) != 1 {
		t.Fatalf("native filters volume=%#v network=%#v errors=%v/%v", volume, network, volumeErr, networkErr)
	}
}

func TestVolumeDetailJSONRoundTrip(t *testing.T) {
	vol := VolumeDetail{
		Name:       "test-vol",
		Driver:     "local",
		Mountpoint: "/var/lib/docker/volumes/test-vol/_data",
		CreatedAt:  "2025-01-15T10:30:00Z",
		Labels:     map[string]string{"app": "api"},
		Scope:      "local",
		Options:    map[string]string{"type": "none"},
		Status:     map[string]any{"availability": "online"},
	}
	raw, err := json.Marshal(vol)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	var decoded VolumeDetail
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if decoded.Name != vol.Name || decoded.Driver != vol.Driver {
		t.Errorf("name/driver mismatch: %q/%q", decoded.Name, decoded.Driver)
	}
	if decoded.Labels["app"] != "api" {
		t.Error("labels not preserved")
	}
	if decoded.Status["availability"] != "online" {
		t.Error("status not preserved")
	}
}

func TestNetworkDetailJSONRoundTrip(t *testing.T) {
	net := NetworkDetail{
		Name:       "test-net",
		ID:         "abc123",
		Created:    "2025-01-15T10:30:00Z",
		Scope:      "local",
		Driver:     "bridge",
		EnableIPv4: true,
		Internal:   false,
		IPAM: NetworkIPAM{
			Config: []NetworkIPAMConfig{
				{Subnet: "172.20.0.0/16", Gateway: "172.20.0.1"},
			},
		},
		Containers: map[string]NetworkEndpoint{
			"ep1": {Name: "c1", IPv4Address: "172.20.0.2/16"},
		},
		Labels: map[string]string{"env": "dev"},
	}
	raw, err := json.Marshal(net)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	var decoded NetworkDetail
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if decoded.Name != "test-net" || decoded.ID != "abc123" {
		t.Errorf("name/id mismatch: %q/%q", decoded.Name, decoded.ID)
	}
	if len(decoded.IPAM.Config) != 1 || decoded.IPAM.Config[0].Subnet != "172.20.0.0/16" {
		t.Error("IPAM config not preserved")
	}
	if ep, ok := decoded.Containers["ep1"]; !ok || ep.IPv4Address != "172.20.0.2/16" {
		t.Error("container endpoint not preserved")
	}
}

func TestRefreshesContainers(t *testing.T) {
	for _, action := range []Action{ActionStart, ActionStop, ActionDie, ActionKill, ActionPause, ActionUnpause, ActionRename, ActionDestroy, ActionCreate} {
		if !RefreshesContainers(action) {
			t.Errorf("%q should refresh containers", action)
		}
	}
	if RefreshesContainers(ActionRestart) {
		t.Error("restart is not emitted as a list event")
	}
}

func TestOperationComposesResourceAndVerb(t *testing.T) {
	if got := Operation(ResourceContainer, "list"); got != "container.list" {
		t.Errorf("Operation(container, list) = %q, want %q", got, "container.list")
	}
	if got := Operation(ResourceVolume, "prune"); got != "volume.prune" {
		t.Errorf("Operation(volume, prune) = %q, want %q", got, "volume.prune")
	}
	if got := Operation(ResourceEvent, "subscribe"); got != "events.subscribe" {
		t.Errorf("Operation(events, subscribe) = %q, want %q", got, "events.subscribe")
	}
}
