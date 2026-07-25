package podman

import (
	"context"
	"fmt"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/driver/podman"
)

type PodmanEngine struct {
	client *podman.Client
}

func NewPodmanEngine(config PodmanEngineConfig) (*PodmanEngine, error) {
	rest := config.REST
	if rest == nil {
		return nil, fmt.Errorf("Podman REST client is required")
	}
	return &PodmanEngine{client: &podman.Client{REST: rest, Host: config.Host, TLS: config.TLS}}, nil
}

type PodmanEngineConfig struct {
	REST *podman.RESTClient
	Host string
	TLS  bool
}

func (e *PodmanEngine) Identity() runtimeapi.Identity {
	version := ""
	if v, err := e.client.REST.APIVersion(context.Background()); err == nil {
		version = v
	}
	return runtimeapi.Identity{Type: runtimeapi.Podman, Endpoint: e.client.Host, Version: version}
}

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

func (e *PodmanEngine) PingContext(ctx context.Context) error {
	_, err := e.client.REST.APIVersion(ctx)
	return err
}

func (e *PodmanEngine) Containers() runtimeapi.ContainerService {
	return PodmanContainerService{Client: e.client}
}
func (e *PodmanEngine) Volumes() runtimeapi.VolumeService {
	return PodmanVolumeService{Client: e.client}
}
func (e *PodmanEngine) Networks() runtimeapi.NetworkService {
	return PodmanNetworkService{Client: e.client}
}
func (e *PodmanEngine) Images() runtimeapi.ImageService { return PodmanImageService{Client: e.client} }
func (e *PodmanEngine) ImageTransfers() runtimeapi.ImageTransferService {
	return PodmanImageTransferService{Client: e.client}
}
func (e *PodmanEngine) Actions() runtimeapi.ResourceActionService {
	return PodmanResourceActionService{Client: e.client}
}
func (e *PodmanEngine) Exec() runtimeapi.ExecService {
	return PodmanExecService{Client: e.client}
}
func (e *PodmanEngine) Events() runtimeapi.EventService { return PodmanEventService{Client: e.client} }
func (e *PodmanEngine) Close() error                    { return nil }
