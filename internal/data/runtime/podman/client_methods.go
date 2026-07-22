package podman

import (
	"context"
	"fmt"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/data/runtime/podman/dto"
)

// Client methods prefer Driver (the optional CGO binding transport) and
// fall back to the REST transport when the driver is nil or the method
// is not implemented on it. Streaming/specialized operations are exposed
// through service accessors rather than direct methods on Client.

// ---------- Containers ----------

// ListContainers fetches containers via the Podman driver when available,
// otherwise via the REST transport.
func (c *Client) ListContainers(ctx context.Context, options runtimeapi.ContainerListOptions) ([]runtimeapi.ContainerSummary, error) {
	var raw []ContainerItem
	var err error
	filters := map[string][]string(options.Filters)
	if c.Driver != nil {
		raw, err = c.Driver.ListContainers(ctx, options.All, options.Limit, filters)
	} else {
		raw, err = ListContainersREST(ctx, c, options.All, options.Limit, filters)
	}
	if err != nil {
		return nil, err
	}
	return MapContainerSummaries(raw), nil
}

// InspectContainer fetches container detail via the REST transport.
func (c *Client) InspectContainer(ctx context.Context, id string) (*runtimeapi.ContainerDetail, error) {
	raw, err := InspectContainerREST(ctx, c, id)
	if err != nil {
		return nil, err
	}
	return MapContainerInspectResponse(*raw), nil
}

// ContainerTop fetches the process listing for a container.
func (c *Client) ContainerTop(ctx context.Context, id string) (runtimeapi.ContainerProcesses, error) {
	raw, err := ContainerTopREST(ctx, c, id)
	if err != nil {
		return runtimeapi.ContainerProcesses{}, err
	}
	return runtimeapi.ContainerProcesses{
		Titles:    raw.Titles,
		Processes: raw.Processes,
	}, nil
}

// ContainerStats fetches container stats via the REST transport.
func (c *Client) ContainerStats(ctx context.Context, id string) (runtimeapi.ContainerStats, error) {
	raw, err := ContainerStatsREST(ctx, c, id)
	if err != nil {
		return runtimeapi.ContainerStats{}, err
	}
	var networkRx, networkTx float64
	for _, net := range raw.Networks {
		networkRx += float64(net.RxBytes)
		networkTx += float64(net.TxBytes)
	}
	var memPct float64
	if raw.MemoryStats.Limit > 0 {
		memPct = float64(raw.MemoryStats.Usage) / float64(raw.MemoryStats.Limit) * 100
	}
	return runtimeapi.ContainerStats{
		MemoryUsage:   float64(raw.MemoryStats.Usage),
		MemoryLimit:   float64(raw.MemoryStats.Limit),
		MemoryPercent: memPct,
		NetworkRx:     networkRx,
		NetworkTx:     networkTx,
	}, nil
}

// ExecuteContainerAction performs a lifecycle action on a container.
func (c *Client) ExecuteContainerAction(ctx context.Context, id string, action runtimeapi.Action, options runtimeapi.ActionOptions) error {
	return ExecuteContainerActionREST(ctx, c, id, action, options)
}

// ---------- Images ----------

// ListImages fetches images via the Podman driver when available,
// otherwise via the REST transport.
func (c *Client) ListImages(ctx context.Context, options runtimeapi.ImageListOptions) ([]runtimeapi.ImageSummary, error) {
	var raw []ImageItem
	var err error
	filters := map[string][]string(options.Filters)
	if c.Driver != nil {
		raw, err = c.Driver.ListImages(ctx, options.All, filters)
	} else {
		raw, err = ListImagesREST(ctx, c, options.All, filters)
	}
	if err != nil {
		return nil, err
	}
	return MapImageSummaries(raw), nil
}

// ExecuteImageAction performs a lifecycle action on an image.
func (c *Client) ExecuteImageAction(ctx context.Context, id string, action runtimeapi.Action, options runtimeapi.ActionOptions, result *runtimeapi.ActionResult) error {
	return ExecuteImageActionREST(ctx, c, id, action, options, result)
}

// TagImage tags an image via the REST transport.
func (c *Client) TagImage(ctx context.Context, source, destination string) error {
	return TagImageREST(ctx, c, source, destination)
}

// PushImage pushes an image via the REST transport.
func (c *Client) PushImage(ctx context.Context, request runtimeapi.ImageTransferRequest, output chan<- runtimeapi.ImageTransferEvent) error {
	return PushImageREST(ctx, c, request, output)
}

// SaveImage saves an image via the REST transport.
func (c *Client) SaveImage(ctx context.Context, request runtimeapi.ImageTransferRequest, output chan<- runtimeapi.ImageTransferEvent) error {
	return SaveImageREST(ctx, c, request, output)
}

// LoadImage loads an image via the REST transport.
func (c *Client) LoadImage(ctx context.Context, request runtimeapi.ImageTransferRequest, output chan<- runtimeapi.ImageTransferEvent) ([]string, error) {
	return LoadImageREST(ctx, c, request, output)
}

// ---------- Networks ----------

// ListNetworks fetches networks via the Podman driver when available,
// otherwise via the REST transport.
func (c *Client) ListNetworks(ctx context.Context, options runtimeapi.NetworkListOptions) ([]runtimeapi.Network, error) {
	var raw []Network
	var err error
	filters := map[string][]string(options.Filters)
	if c.Driver != nil {
		raw, err = c.Driver.ListNetworks(ctx, filters)
	} else {
		raw, err = ListNetworksREST(ctx, c, filters)
	}
	if err != nil {
		return nil, err
	}
	return MapNetworks(raw), nil
}

// InspectNetwork fetches network detail via the driver when available,
// otherwise via REST.
func (c *Client) InspectNetwork(ctx context.Context, id string) (*runtimeapi.NetworkDetail, error) {
	var raw *NetworkInspectItem
	var err error
	if c.Driver != nil {
		raw, err = c.Driver.InspectNetwork(ctx, id)
	} else {
		raw, err = InspectNetworkREST(ctx, c, id)
	}
	if err != nil {
		return nil, err
	}
	return MapNetworkInspect(*raw), nil
}

// CreateNetwork creates a network via the driver when available.
func (c *Client) CreateNetwork(ctx context.Context, options runtimeapi.NetworkCreateOptions) (*runtimeapi.Network, error) {
	var raw *Network
	var err error
	if c.Driver != nil {
		raw, err = c.Driver.CreateNetwork(ctx, options.Name, options.Driver, options.Internal, options.EnableIPv6, options.Labels, options.Options)
	} else {
		raw, err = CreateNetworkREST(ctx, c, options.Name, options.Driver, options.Internal, options.EnableIPv6, options.Labels, options.Options)
	}
	if err != nil {
		return nil, err
	}
	networks := MapNetworks([]Network{*raw})
	if len(networks) == 0 {
		return nil, nil
	}
	return &networks[0], nil
}

// RemoveNetwork deletes a network via the driver when available.
func (c *Client) RemoveNetwork(ctx context.Context, id string) error {
	if c.Driver != nil {
		return c.Driver.RemoveNetwork(ctx, id)
	}
	return RemoveNetworkREST(ctx, c, id)
}

// PruneNetworks removes unused networks via the driver when available.
func (c *Client) PruneNetworks(ctx context.Context, options runtimeapi.PruneOptions) (runtimeapi.PruneResult, error) {
	filters := map[string][]string(options.Filters)
	var reports []dto.NetworkPruneReportItem
	var err error
	if c.Driver != nil {
		reports, err = c.Driver.PruneNetworks(ctx, filters)
	} else {
		reports, err = PruneNetworksREST(ctx, c, filters)
	}
	if err != nil {
		return runtimeapi.PruneResult{}, err
	}
	return assembleNetworkPruneResult(reports), nil
}

// ---------- Volumes ----------

// ListVolumes fetches volumes via the Podman driver when available,
// otherwise via the REST transport.
func (c *Client) ListVolumes(ctx context.Context, options runtimeapi.VolumeListOptions) ([]runtimeapi.Volume, error) {
	var raw []VolumeItem
	var err error
	filters := map[string][]string(options.Filters)
	if c.Driver != nil {
		raw, err = c.Driver.ListVolumes(ctx, filters)
	} else {
		raw, err = ListVolumesREST(ctx, c, filters)
	}
	if err != nil {
		return nil, err
	}
	return MapVolumes(raw), nil
}

// InspectVolume fetches volume detail via the driver when available.
func (c *Client) InspectVolume(ctx context.Context, name string) (*runtimeapi.VolumeDetail, error) {
	var raw *VolumeItem
	var err error
	if c.Driver != nil {
		raw, err = c.Driver.InspectVolume(ctx, name)
	} else {
		raw, err = InspectVolumeREST(ctx, c, name)
	}
	if err != nil {
		return nil, err
	}
	return MapVolumeInspect(*raw), nil
}

// CreateVolume creates a volume via the driver when available.
func (c *Client) CreateVolume(ctx context.Context, options runtimeapi.VolumeCreateOptions) (*runtimeapi.Volume, error) {
	var raw *VolumeItem
	var err error
	if c.Driver != nil {
		raw, err = c.Driver.CreateVolume(ctx, options.Name, options.Driver, options.Labels, options.Options)
	} else {
		raw, err = CreateVolumeREST(ctx, c, options.Name, options.Driver, options.Labels, options.Options)
	}
	if err != nil {
		return nil, err
	}
	volumes := MapVolumes([]VolumeItem{*raw})
	if len(volumes) == 0 {
		return nil, nil
	}
	return &volumes[0], nil
}

// RemoveVolume deletes a volume via the driver when available.
func (c *Client) RemoveVolume(ctx context.Context, name string, force bool) error {
	if c.Driver != nil {
		return c.Driver.RemoveVolume(ctx, name, force)
	}
	return RemoveVolumeREST(ctx, c, name, force)
}

// PruneVolumes removes unused volumes via the driver when available.
func (c *Client) PruneVolumes(ctx context.Context, options runtimeapi.PruneOptions) (runtimeapi.PruneResult, error) {
	filters := map[string][]string(options.Filters)
	var reports []dto.VolumePruneReportItem
	var err error
	if c.Driver != nil {
		reports, err = c.Driver.PruneVolumes(ctx, filters)
	} else {
		reports, err = PruneVolumesREST(ctx, c, filters)
	}
	if err != nil {
		return runtimeapi.PruneResult{}, err
	}
	return assembleVolumePruneResult(reports), nil
}

// ---------- Service accessors ----------

// ContainerService returns a runtimeapi-compatible service facade for
// streaming operations (Logs) not exposed as direct Client methods.
func (c *Client) ContainerService() runtimeapi.ContainerService {
	return ContainerService{Client: c}
}

// VolumeService returns a runtimeapi-compatible service facade.
func (c *Client) VolumeService() runtimeapi.VolumeService {
	return VolumeService{Client: c}
}

// NetworkService returns a runtimeapi-compatible service facade.
func (c *Client) NetworkService() runtimeapi.NetworkService {
	return NetworkService{Client: c}
}

// ImageService returns a runtimeapi-compatible service facade.
func (c *Client) ImageService() runtimeapi.ImageService {
	return ImageService{Client: c}
}

// ExecService returns a runtimeapi-compatible exec facade.
func (c *Client) ExecService() runtimeapi.ExecService {
	return ExecService{Client: c}
}

// EventService returns a runtimeapi-compatible event facade.
func (c *Client) EventService() runtimeapi.EventService {
	return EventService{Client: c}
}

// ImageTransferService returns a runtimeapi-compatible image transfer facade.
func (c *Client) ImageTransferService() runtimeapi.ImageTransferService {
	return ImageTransferService{Client: c}
}

// ResourceActionService returns a runtimeapi-compatible resource action facade.
func (c *Client) ResourceActionService() runtimeapi.ResourceActionService {
	return ResourceActionService{Client: c}
}

// ---------- Meta ----------

// APIVersion returns the negotiated Libpod API version.
func (c *Client) APIVersion(ctx context.Context) (string, error) {
	if c.REST == nil {
		return "", errPodmanRESTNotReady
	}
	return c.REST.APIVersion(ctx)
}

// Ping verifies the runtime is reachable.
func (c *Client) Ping(ctx context.Context) error {
	if c.REST == nil {
		return errPodmanRESTNotReady
	}
	_, err := c.REST.APIVersion(ctx)
	return err
}

// Close releases the driver and REST transport.
func (c *Client) Close() error {
	if c.Driver != nil {
		_ = c.Driver.Close()
	}
	return nil
}

// assembleNetworkPruneResult converts raw Podman network prune reports into
// a runtime.PruneResult.
func assembleNetworkPruneResult(reports []dto.NetworkPruneReportItem) runtimeapi.PruneResult {
	result := runtimeapi.PruneResult{}
	for _, report := range reports {
		if report.Error != "" {
			result.Resources = append(result.Resources, runtimeapi.ResourceResult{ID: report.Name, Error: fmt.Errorf("%s", report.Error)})
		} else {
			result.Resources = append(result.Resources, runtimeapi.ResourceResult{ID: report.Name})
		}
	}
	return result
}

// assembleVolumePruneResult converts raw Podman volume prune reports into
// a runtime.PruneResult.
func assembleVolumePruneResult(reports []dto.VolumePruneReportItem) runtimeapi.PruneResult {
	result := runtimeapi.PruneResult{}
	for _, report := range reports {
		if report.Err != "" {
			result.Resources = append(result.Resources, runtimeapi.ResourceResult{ID: report.ID, Error: fmt.Errorf("%s", report.Err)})
		} else {
			result.Resources = append(result.Resources, runtimeapi.ResourceResult{ID: report.ID})
			if report.Size > 0 {
				result.SpaceReclaimed += uint64(report.Size)
			}
		}
	}
	return result
}
