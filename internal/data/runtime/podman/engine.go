// Package podman implements the runtime.Engine interface for Podman
// containers using the Podman REST API. It adapts the Podman-specific
// driver into the unified runtime abstraction shared with Docker.
package podman

import (
	"context"
	"fmt"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/driver/podman"
)

// PodmanEngine implements runtime.Engine by delegating to a Podman
// REST client. All resource-specific services are created lazily
// on each call and share the same underlying connection.
type PodmanEngine struct {
	client *podman.Client
}

// NewPodmanEngine creates a new PodmanEngine from the given configuration.
// The REST client must be non-nil; Host and TLS configure the connection
// target and transport security.
func NewPodmanEngine(config PodmanEngineConfig) (*PodmanEngine, error) {
	rest := config.REST
	if rest == nil {
		return nil, fmt.Errorf("Podman REST client is required")
	}
	return &PodmanEngine{client: &podman.Client{REST: rest, Host: config.Host, TLS: config.TLS}}, nil
}

// PodmanEngineConfig holds the connection parameters for a PodmanEngine.
type PodmanEngineConfig struct {
	REST *podman.RESTClient // REST is the Podman REST API client.
	Host string             // Host is the Podman socket endpoint.
	TLS  bool               // TLS indicates whether to use TLS transport.
}

// Identity returns the runtime identity containing the Podman version
// and the connected socket endpoint.
func (e *PodmanEngine) Identity() runtimeapi.Identity {
	version := ""
	if v, err := e.client.REST.APIVersion(context.Background()); err == nil {
		version = v
	}
	return runtimeapi.Identity{Type: runtimeapi.Podman, Endpoint: e.client.Host, Version: version}
}

// Capabilities reports which runtime capabilities Podman supports and
// any degraded modes such as adapter-level post-filtering.
func (e *PodmanEngine) Capabilities() runtimeapi.CapabilitySet {
	return runtimeapi.CapabilitySet{
		runtimeapi.CapabilityContainers: {Support: runtimeapi.Available},
		runtimeapi.CapabilityImages:     {Support: runtimeapi.Available},
		runtimeapi.CapabilityVolumes:    {Support: runtimeapi.Available},
		runtimeapi.CapabilityNetworks:   {Support: runtimeapi.Available},
		runtimeapi.CapabilityEvents:     {Support: runtimeapi.Available},
		runtimeapi.CapabilityExec:       {Support: runtimeapi.Available},
		runtimeapi.CapabilityFiltering: {
			Support:    runtimeapi.Degraded,
			Reason:     "the first same-field value is native; remaining values use equivalent adapter post-filtering",
			ReasonCode: "same_field_and_post_filter",
		},
		runtimeapi.CapabilityContainerListFilter: {
			Support:    runtimeapi.Degraded,
			Reason:     "the first same-field value is native; remaining values use equivalent adapter post-filtering",
			ReasonCode: "same_field_and_post_filter",
		},
		runtimeapi.CapabilityImageListFilter: {
			Support:    runtimeapi.Degraded,
			Reason:     "the first same-field AND uses adapter post-filtering; relative before, since, and until combinations are rejected",
			ReasonCode: "partial_same_field_and_post_filter",
		},
		runtimeapi.CapabilityVolumeListFilter: {
			Support:    runtimeapi.Degraded,
			Reason:     "the first same-field AND uses adapter post-filtering; multiple dangling conditions are rejected",
			ReasonCode: "partial_same_field_and_post_filter",
		},
		runtimeapi.CapabilityNetworkListFilter: {
			Support:    runtimeapi.Degraded,
			Reason:     "the first same-field AND uses adapter post-filtering; multiple type conditions are rejected",
			ReasonCode: "partial_same_field_and_post_filter",
		},
		runtimeapi.CapabilityEventFilter: {Support: runtimeapi.Available},
		runtimeapi.CapabilityExecResize:  {Support: runtimeapi.Available},
		runtimeapi.CapabilityStatsStream: {Support: runtimeapi.Available},
	}
}

// PingContext probes the Podman daemon by requesting its API version.
func (e *PodmanEngine) PingContext(ctx context.Context) error {
	_, err := e.client.REST.APIVersion(ctx)
	return err
}

// Containers returns a container service backed by the Podman REST API.
func (e *PodmanEngine) Containers() runtimeapi.ContainerService {
	return PodmanContainerService{Client: e.client}
}

// Volumes returns a volume service backed by the Podman REST API.
func (e *PodmanEngine) Volumes() runtimeapi.VolumeService {
	return PodmanVolumeService{Client: e.client}
}

// Networks returns a network service backed by the Podman REST API.
func (e *PodmanEngine) Networks() runtimeapi.NetworkService {
	return PodmanNetworkService{Client: e.client}
}

// Images returns an image service backed by the Podman REST API.
func (e *PodmanEngine) Images() runtimeapi.ImageService { return PodmanImageService{Client: e.client} }

// ImageTransfers returns an image transfer service for pull/push/save/load.
func (e *PodmanEngine) ImageTransfers() runtimeapi.ImageTransferService {
	return PodmanImageTransferService{Client: e.client}
}

// Actions returns a resource action service for start/stop/restart etc.
func (e *PodmanEngine) Actions() runtimeapi.ResourceActionService {
	return PodmanResourceActionService{Client: e.client}
}

// Exec returns an exec service for attaching to running containers.
func (e *PodmanEngine) Exec() runtimeapi.ExecService {
	return PodmanExecService{Client: e.client}
}

// Events returns an event subscription service.
func (e *PodmanEngine) Events() runtimeapi.EventService { return PodmanEventService{Client: e.client} }

// Close releases any resources held by the engine. The Podman REST
// client is stateless so this is a no-op.
func (e *PodmanEngine) Close() error { return nil }
