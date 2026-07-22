package podman

import (
	"context"
	"errors"
	"testing"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/data/runtime/podman/dto"
)

// stubDriver is a minimal DriverBackend used to verify the Client's
// "driver-first, REST-fallback" dispatch.
type stubDriver struct {
	called bool
}

func (d *stubDriver) ListContainers(ctx context.Context, all bool, limit int, filters map[string][]string) ([]ContainerItem, error) {
	d.called = true
	return []ContainerItem{{ID: "driver-list"}}, nil
}
func (d *stubDriver) ListImages(ctx context.Context, all bool, filters map[string][]string) ([]ImageItem, error) {
	d.called = true
	return []ImageItem{{ID: "driver-image"}}, nil
}
func (d *stubDriver) ListNetworks(ctx context.Context, filters map[string][]string) ([]Network, error) {
	d.called = true
	return []Network{{ID: "driver-net"}}, nil
}
func (d *stubDriver) InspectNetwork(ctx context.Context, id string) (*NetworkInspectItem, error) {
	d.called = true
	return &NetworkInspectItem{Network: Network{ID: id}}, nil
}
func (d *stubDriver) CreateNetwork(ctx context.Context, name, driver string, internal, ipv6 bool, labels, options map[string]string) (*Network, error) {
	d.called = true
	return &Network{ID: "new-net"}, nil
}
func (d *stubDriver) RemoveNetwork(ctx context.Context, id string) error {
	d.called = true
	return nil
}
func (d *stubDriver) PruneNetworks(ctx context.Context, filters map[string][]string) ([]dto.NetworkPruneReportItem, error) {
	d.called = true
	return nil, nil
}
func (d *stubDriver) ListVolumes(ctx context.Context, filters map[string][]string) ([]VolumeItem, error) {
	d.called = true
	return []VolumeItem{{Name: "driver-vol"}}, nil
}
func (d *stubDriver) InspectVolume(ctx context.Context, name string) (*VolumeItem, error) {
	d.called = true
	return &VolumeItem{Name: name}, nil
}
func (d *stubDriver) CreateVolume(ctx context.Context, name, driver string, labels, options map[string]string) (*VolumeItem, error) {
	d.called = true
	return &VolumeItem{Name: "new-vol"}, nil
}
func (d *stubDriver) RemoveVolume(ctx context.Context, name string, force bool) error {
	d.called = true
	return nil
}
func (d *stubDriver) PruneVolumes(ctx context.Context, filters map[string][]string) ([]dto.VolumePruneReportItem, error) {
	d.called = true
	return nil, nil
}
func (d *stubDriver) Close() error { return nil }

// TestClientPrefersDriver verifies that driver-backed methods route through
// the driver when one is installed.
func TestClientPrefersDriver(t *testing.T) {
	d := &stubDriver{}
	c := &Client{Driver: d, REST: nil}

	if _, err := c.ListContainers(context.Background(), runtimeapi.ContainerListOptions{}); err != nil {
		t.Fatalf("ListContainers via driver: %v", err)
	}
	if !d.called {
		t.Fatal("driver.ListContainers was not invoked")
	}
	d.called = false

	if _, err := c.ListImages(context.Background(), runtimeapi.ImageListOptions{}); err != nil {
		t.Fatalf("ListImages via driver: %v", err)
	}
	if !d.called {
		t.Fatal("driver.ListImages was not invoked")
	}
	d.called = false

	if _, err := c.ListNetworks(context.Background(), runtimeapi.NetworkListOptions{}); err != nil {
		t.Fatalf("ListNetworks via driver: %v", err)
	}
	if !d.called {
		t.Fatal("driver.ListNetworks was not invoked")
	}
	d.called = false

	if _, err := c.ListVolumes(context.Background(), runtimeapi.VolumeListOptions{}); err != nil {
		t.Fatalf("ListVolumes via driver: %v", err)
	}
	if !d.called {
		t.Fatal("driver.ListVolumes was not invoked")
	}
}

// TestClientFallsBackToRESTWithoutDriver confirms that methods which
// require a REST transport surface a clear error when REST is nil.
func TestClientFallsBackToRESTWithoutDriver(t *testing.T) {
	c := &Client{Driver: nil, REST: nil}
	_, err := c.ListContainers(context.Background(), runtimeapi.ContainerListOptions{})
	if !errors.Is(err, errPodmanRESTNotReady) {
		t.Fatalf("expected errPodmanRESTNotReady, got %v", err)
	}
}

// TestClientServiceAccessors exercise the service accessors, which always
// use the REST transport today.
func TestClientServiceAccessors(t *testing.T) {
	c := &Client{Driver: &stubDriver{}, REST: nil}
	if c.ContainerService() == nil {
		t.Fatal("ContainerService accessor returned nil")
	}
	if c.VolumeService() == nil {
		t.Fatal("VolumeService accessor returned nil")
	}
	if c.NetworkService() == nil {
		t.Fatal("NetworkService accessor returned nil")
	}
	if c.ImageService() == nil {
		t.Fatal("ImageService accessor returned nil")
	}
	if c.ExecService() == nil {
		t.Fatal("ExecService accessor returned nil")
	}
	if c.EventService() == nil {
		t.Fatal("EventService accessor returned nil")
	}
	if c.ImageTransferService() == nil {
		t.Fatal("ImageTransferService accessor returned nil")
	}
	if c.ResourceActionService() == nil {
		t.Fatal("ResourceActionService accessor returned nil")
	}
}

// TestNewClientWithoutRESTReturnsErrorPing verifies Ping surfaces the
// "REST not ready" error when constructed without a REST transport.
func TestNewClientWithoutRESTReturnsErrorPing(t *testing.T) {
	c := NewClient(nil, "unix:///run/podman/podman.sock", false)
	if err := c.Ping(context.Background()); !errors.Is(err, errPodmanRESTNotReady) {
		t.Fatalf("expected errPodmanRESTNotReady, got %v", err)
	}
}
