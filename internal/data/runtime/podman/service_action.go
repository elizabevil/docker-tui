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
		return s.executeContainer(ctx, ref.ID, action, options)
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

// executeContainer dispatches TASK-019 advanced container actions to
// the dedicated ContainerService methods, falling back to the existing
// lifecycle path for the original actions.
func (s PodmanResourceActionService) executeContainer(ctx context.Context, id string, action runtimeapi.Action, options runtimeapi.ActionOptions) error {
	switch action {
	case runtimeapi.ActionUpdate, runtimeapi.ActionDiff, runtimeapi.ActionExport,
		runtimeapi.ActionCommit, runtimeapi.ActionWait, runtimeapi.ActionCopy:
		return s.dispatchAdvanced(ctx, id, action, options)
	default:
		return s.Client.REST.ExecuteContainerAction(ctx, id, string(action), options.Timeout, options.Force, options.Signal, options.Name)
	}
}

// dispatchAdvanced mirrors the Docker adapter's TASK-019 routing. It
// funnels the advanced actions through PodmanContainerService so the
// REST client stays the single integration point.
func (s PodmanResourceActionService) dispatchAdvanced(ctx context.Context, id string, action runtimeapi.Action, options runtimeapi.ActionOptions) error {
	svc := PodmanContainerService{Client: s.Client}
	switch action {
	case runtimeapi.ActionUpdate:
		_, err := svc.Update(ctx, id, runtimeapi.ContainerUpdateOptions{
			Memory:            options.Memory,
			NanoCPUs:          options.NanoCPUs,
			RestartPolicy:     options.RestartPolicy,
			RestartMaxRetries: options.RestartMaxRetries,
		})
		return err
	case runtimeapi.ActionDiff:
		_, err := svc.Diff(ctx, id)
		return err
	case runtimeapi.ActionExport:
		rc, err := svc.Export(ctx, id)
		if err != nil {
			return err
		}
		return rc.Close()
	case runtimeapi.ActionCommit:
		_, err := svc.Commit(ctx, id, runtimeapi.ContainerCommitOptions{
			Repository: options.Repository,
			Tag:        options.Tag,
			Comment:    options.Comment,
			Author:     options.Author,
			Pause:      options.Pause,
		})
		return err
	case runtimeapi.ActionWait:
		_, err := svc.Wait(ctx, id, options.Condition)
		return err
	case runtimeapi.ActionCopy:
		rc, err := svc.CopyFromContainer(ctx, id, options.SourcePath)
		if err != nil {
			return err
		}
		return rc.Close()
	}
	return runtimeapi.UnsupportedError(runtimeapi.Operation(runtimeapi.ResourceContainer, string(action)))
}
