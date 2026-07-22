package podman

import (
	"context"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

// VolumeService implements runtime.VolumeService for the Podman adapter.
type VolumeService struct {
	Client *Client
}

func (s VolumeService) List(ctx context.Context, options runtimeapi.VolumeListOptions) ([]runtimeapi.Volume, error) {
	filters := map[string][]string(options.Filters)
	raw, err := ListVolumesREST(ctx, s.Client, filters)
	if err == nil {
		raw, err = PostFilterVolumes(raw, options.Filters)
	}
	if err != nil {
		return nil, mapRuntimeError(err, "volume.list", runtimeapi.ResourceRef{Type: runtimeapi.ResourceVolume}, runtimeapi.Podman)
	}
	return MapVolumes(raw), nil
}

func (s VolumeService) Inspect(ctx context.Context, name string) (*runtimeapi.VolumeDetail, error) {
	raw, err := InspectVolumeREST(ctx, s.Client, name)
	if err != nil {
		return nil, mapRuntimeError(err, "volume.inspect", runtimeapi.ResourceRef{Type: runtimeapi.ResourceVolume, ID: name}, runtimeapi.Podman)
	}
	return MapVolumeInspect(*raw), nil
}

func (s VolumeService) Create(ctx context.Context, options runtimeapi.VolumeCreateOptions) (*runtimeapi.Volume, error) {
	raw, err := CreateVolumeREST(ctx, s.Client, options.Name, options.Driver, options.Labels, options.Options)
	if err != nil {
		return nil, mapRuntimeError(err, "volume.create", runtimeapi.ResourceRef{Type: runtimeapi.ResourceVolume, ID: options.Name}, runtimeapi.Podman)
	}
	volumes := MapVolumes([]VolumeItem{*raw})
	if len(volumes) == 0 {
		return nil, nil
	}
	return &volumes[0], nil
}

func (s VolumeService) Remove(ctx context.Context, name string, force bool) error {
	err := RemoveVolumeREST(ctx, s.Client, name, force)
	return mapRuntimeError(err, "volume.remove", runtimeapi.ResourceRef{Type: runtimeapi.ResourceVolume, ID: name}, runtimeapi.Podman)
}

func (s VolumeService) Prune(ctx context.Context, options runtimeapi.PruneOptions) (runtimeapi.PruneResult, error) {
	filters := map[string][]string(options.Filters)
	reports, err := PruneVolumesREST(ctx, s.Client, filters)
	if err != nil {
		return runtimeapi.PruneResult{}, mapRuntimeError(err, "volume.prune", runtimeapi.ResourceRef{Type: runtimeapi.ResourceVolume}, runtimeapi.Podman)
	}
	return assembleVolumePruneResult(reports), nil
}

// NetworkService implements runtime.NetworkService for the Podman adapter.
type NetworkService struct {
	Client *Client
}

func (s NetworkService) List(ctx context.Context, options runtimeapi.NetworkListOptions) ([]runtimeapi.Network, error) {
	filters := map[string][]string(options.Filters)
	raw, err := ListNetworksREST(ctx, s.Client, filters)
	if err == nil {
		raw, err = PostFilterNetworks(raw, options.Filters)
	}
	if err != nil {
		return nil, mapRuntimeError(err, "network.list", runtimeapi.ResourceRef{Type: runtimeapi.ResourceNetwork}, runtimeapi.Podman)
	}
	return MapNetworks(raw), nil
}

func (s NetworkService) Inspect(ctx context.Context, id string) (*runtimeapi.NetworkDetail, error) {
	raw, err := InspectNetworkREST(ctx, s.Client, id)
	if err != nil {
		return nil, mapRuntimeError(err, "network.inspect", runtimeapi.ResourceRef{Type: runtimeapi.ResourceNetwork, ID: id}, runtimeapi.Podman)
	}
	return MapNetworkInspect(*raw), nil
}

func (s NetworkService) Create(ctx context.Context, options runtimeapi.NetworkCreateOptions) (*runtimeapi.Network, error) {
	raw, err := CreateNetworkREST(ctx, s.Client, options.Name, options.Driver, options.Internal, options.EnableIPv6, options.Labels, options.Options)
	if err != nil {
		return nil, mapRuntimeError(err, "network.create", runtimeapi.ResourceRef{Type: runtimeapi.ResourceNetwork, ID: options.Name}, runtimeapi.Podman)
	}
	networks := MapNetworks([]Network{*raw})
	if len(networks) == 0 {
		return nil, nil
	}
	return &networks[0], nil
}

func (s NetworkService) Remove(ctx context.Context, id string) error {
	err := RemoveNetworkREST(ctx, s.Client, id)
	return mapRuntimeError(err, "network.remove", runtimeapi.ResourceRef{Type: runtimeapi.ResourceNetwork, ID: id}, runtimeapi.Podman)
}

func (s NetworkService) Prune(ctx context.Context, options runtimeapi.PruneOptions) (runtimeapi.PruneResult, error) {
	filters := map[string][]string(options.Filters)
	reports, err := PruneNetworksREST(ctx, s.Client, filters)
	if err != nil {
		return runtimeapi.PruneResult{}, mapRuntimeError(err, "network.prune", runtimeapi.ResourceRef{Type: runtimeapi.ResourceNetwork}, runtimeapi.Podman)
	}
	return assembleNetworkPruneResult(reports), nil
}

// ImageService implements runtime.ImageService for the Podman adapter.
type ImageService struct {
	Client *Client
}

func (s ImageService) List(ctx context.Context, options runtimeapi.ImageListOptions) ([]runtimeapi.ImageSummary, error) {
	filters := map[string][]string(options.Filters)
	raw, err := ListImagesREST(ctx, s.Client, options.All, filters)
	if err == nil {
		raw, err = PostFilterImages(raw, options.Filters)
	}
	if err != nil {
		return nil, mapRuntimeError(err, "image.list", runtimeapi.ResourceRef{Type: runtimeapi.ResourceImage}, runtimeapi.Podman)
	}
	return MapImageSummaries(raw), nil
}

func (s ImageService) Inspect(ctx context.Context, summary runtimeapi.ImageSummary) (*runtimeapi.ImageDetail, error) {
	// TODO: implement image inspect via Podman REST API
	return nil, runtimeapi.NewError(runtimeapi.ErrorUnsupported, "image.inspect", summary.ID, nil)
}

// ImageTransferService implements runtime.ImageTransferService for the Podman adapter.
type ImageTransferService struct {
	Client *Client
}

func (s ImageTransferService) Run(ctx context.Context, request runtimeapi.ImageTransferRequest) (<-chan runtimeapi.ImageTransferEvent, error) {
	output := make(chan runtimeapi.ImageTransferEvent, 32)
	go func() {
		defer close(output)
		result, err := s.execute(ctx, request, output)
		if err != nil {
			err = mapRuntimeError(err, "image."+string(request.Operation), runtimeapi.ResourceRef{Type: runtimeapi.ResourceImage, ID: request.Source}, runtimeapi.Podman)
		}
		output <- runtimeapi.ImageTransferEvent{Result: result, Error: err, Done: true}
	}()
	return output, nil
}

func (s ImageTransferService) execute(ctx context.Context, request runtimeapi.ImageTransferRequest, output chan<- runtimeapi.ImageTransferEvent) (*runtimeapi.ImageTransferResult, error) {
	result := &runtimeapi.ImageTransferResult{Operation: request.Operation, Source: request.Source, Destination: request.Destination, Path: request.Path}
	switch request.Operation {
	case runtimeapi.ImageTransferTag:
		return result, TagImageREST(ctx, s.Client, request.Source, request.Destination)
	case runtimeapi.ImageTransferPush:
		return result, PushImageREST(ctx, s.Client, request, output)
	case runtimeapi.ImageTransferSave:
		return result, SaveImageREST(ctx, s.Client, request, output)
	case runtimeapi.ImageTransferLoad:
		refs, err := LoadImageREST(ctx, s.Client, request, output)
		result.References = refs
		return result, err
	default:
		return nil, runtimeapi.UnsupportedError("image." + string(request.Operation))
	}
}

// ResourceActionService implements runtime.ResourceActionService for the Podman adapter.
type ResourceActionService struct {
	Client *Client
}

func (s ResourceActionService) Execute(ctx context.Context, ref runtimeapi.ResourceRef, action runtimeapi.Action, options runtimeapi.ActionOptions) (runtimeapi.ActionResult, error) {
	result := runtimeapi.ActionResult{Resource: ref, Action: action}
	err := s.execute(ctx, ref, action, options, &result)
	if err != nil {
		return runtimeapi.ActionResult{}, mapRuntimeError(err, string(ref.Type)+"."+string(action), ref, runtimeapi.Podman)
	}
	return result, nil
}

func (s ResourceActionService) execute(ctx context.Context, ref runtimeapi.ResourceRef, action runtimeapi.Action, options runtimeapi.ActionOptions, result *runtimeapi.ActionResult) error {
	switch ref.Type {
	case runtimeapi.ResourceContainer:
		return ExecuteContainerActionREST(ctx, s.Client, ref.ID, action, options)
	case runtimeapi.ResourceImage:
		return ExecuteImageActionREST(ctx, s.Client, ref.ID, action, options, result)
	case runtimeapi.ResourceVolume:
		if action != runtimeapi.ActionRemove {
			return runtimeapi.UnsupportedError("volume." + string(action))
		}
		return RemoveVolumeREST(ctx, s.Client, ref.ID, options.Force)
	case runtimeapi.ResourceNetwork:
		if action != runtimeapi.ActionRemove {
			return runtimeapi.UnsupportedError("network." + string(action))
		}
		return RemoveNetworkREST(ctx, s.Client, ref.ID)
	default:
		return runtimeapi.NewError(runtimeapi.ErrorInvalid, "resource.action", ref.ID, nil)
	}
}
