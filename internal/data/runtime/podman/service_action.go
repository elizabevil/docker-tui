package podman

import (
	"context"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	podman "github.com/elizabevil/docker-tui/internal/driver/podman"
)

// PodmanResourceActionService implements ResourceActionService for the Podman adapter.
type PodmanResourceActionService struct {
	Client *podman.Client
}

func (s PodmanResourceActionService) Execute(ctx context.Context, ref runtimeapi.ResourceRef, action runtimeapi.Action, options runtimeapi.ActionOptions) (runtimeapi.ActionResult, error) {
	result := runtimeapi.ActionResult{Resource: ref, Action: action}
	err := s.execute(ctx, ref, action, options, &result)
	if err != nil {
		return runtimeapi.ActionResult{}, runtimeapi.MapRuntimeError(err, string(ref.Type)+"."+string(action), ref, runtimeapi.Podman)
	}
	return result, nil
}

func (s PodmanResourceActionService) execute(ctx context.Context, ref runtimeapi.ResourceRef, action runtimeapi.Action, options runtimeapi.ActionOptions, result *runtimeapi.ActionResult) error {
	switch ref.Type {
	case runtimeapi.ResourceContainer:
		return s.Client.REST.ExecuteContainerAction(ctx, ref.ID, string(action), options.Timeout, options.Force, options.Signal, options.Name)
	case runtimeapi.ResourceImage:
		stream, err := s.Client.REST.ExecuteImageAction(ctx, ref.ID, string(action), options.Force, &result.SpaceReclaimed)
		if err != nil {
			return err
		}
		if stream != nil {
			defer stream.Close()
			return runtimeapi.DecodeImageProgress(ctx, stream, nil)
		}
		return nil
	case runtimeapi.ResourceVolume:
		if action != runtimeapi.ActionRemove {
			return runtimeapi.UnsupportedError(runtimeapi.Operation(runtimeapi.ResourceVolume, string(action)))
		}
		return s.Client.REST.RemoveVolume(ctx, ref.ID, options.Force)
	case runtimeapi.ResourceNetwork:
		if action != runtimeapi.ActionRemove {
			return runtimeapi.UnsupportedError(runtimeapi.Operation(runtimeapi.ResourceNetwork, string(action)))
		}
		return s.Client.REST.RemoveNetwork(ctx, ref.ID)
	default:
		return runtimeapi.NewError(runtimeapi.ErrorInvalid, "resource.action", ref.ID, nil)
	}
}
