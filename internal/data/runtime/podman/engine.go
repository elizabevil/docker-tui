package podman

import (
	"context"
	"fmt"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

// Engine implements runtime.Engine for the Podman backend. It selects
// between CGO bindings and REST transport based on the client configuration.
type Engine struct {
	client *Client
}

// EngineConfig holds the parameters needed to create a Podman engine.
type EngineConfig struct {
	REST *RESTClient
	Host string
	TLS  bool
}

// NewEngine creates a Podman runtime engine.
func NewEngine(config EngineConfig) (*Engine, error) {
	rest := config.REST
	if rest == nil {
		return nil, fmt.Errorf("Podman REST client is required")
	}
	return &Engine{client: &Client{REST: rest, Host: config.Host, TLS: config.TLS}}, nil
}

func (e *Engine) Identity() runtimeapi.Identity {
	version := ""
	if v, err := e.client.REST.APIVersion(context.Background()); err == nil {
		version = v
	}
	return runtimeapi.Identity{Type: runtimeapi.Podman, Endpoint: e.client.Host, Version: version}
}

func (e *Engine) Capabilities() runtimeapi.CapabilitySet {
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

func (e *Engine) PingContext(ctx context.Context) error {
	_, err := e.client.REST.APIVersion(ctx)
	return err
}

func (e *Engine) Containers() runtimeapi.ContainerService { return ContainerService{Client: e.client} }
func (e *Engine) Volumes() runtimeapi.VolumeService       { return VolumeService{Client: e.client} }
func (e *Engine) Networks() runtimeapi.NetworkService     { return NetworkService{Client: e.client} }
func (e *Engine) Images() runtimeapi.ImageService         { return ImageService{Client: e.client} }
func (e *Engine) ImageTransfers() runtimeapi.ImageTransferService {
	return ImageTransferService{Client: e.client}
}
func (e *Engine) Actions() runtimeapi.ResourceActionService {
	return ResourceActionService{Client: e.client}
}
func (e *Engine) Exec() runtimeapi.ExecService    { return ExecService{Client: e.client} }
func (e *Engine) Events() runtimeapi.EventService { return EventService{Client: e.client} }
func (e *Engine) Close() error                    { return nil }
