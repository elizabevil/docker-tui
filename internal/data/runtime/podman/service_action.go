package podman

import (
	"context"
	"fmt"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/driver/podman"
)

// PodmanResourceActionService implements ResourceActionService for the Podman adapter.
type PodmanResourceActionService struct {
	Client *podman.Client
}

func (s PodmanResourceActionService) Execute(ctx context.Context, ref runtimeapi.ResourceRef, action runtimeapi.Action, options runtimeapi.ActionOptions) (runtimeapi.ActionResult, error) {
	result := runtimeapi.ActionResult{Resource: ref, Action: action}
	err := s.execute(ctx, ref, action, options, &result)
	if err != nil {
		return runtimeapi.ActionResult{}, runtimeapi.MapRuntimeError(err,
			string(ref.Type)+"."+string(action), ref, runtimeapi.Podman)
	}
	return result, nil
}

func (s PodmanResourceActionService) execute(ctx context.Context, ref runtimeapi.ResourceRef, action runtimeapi.Action, options runtimeapi.ActionOptions, result *runtimeapi.ActionResult) error {
	lc := options.Lifecycle
	switch ref.Type {
	case runtimeapi.ResourceContainer:
		return s.executeContainer(ctx, ref.ID, action, options)
	case runtimeapi.ResourceImage:
		stream, err := s.Client.REST.ExecuteImageAction(ctx, ref.ID, string(action), lc.Force, &result.SpaceReclaimed)
		if err != nil {
			return err
		}
		if stream != nil {
			defer func() { _ = stream.Close() }() //nolint:errcheck // progress stream drained via DecodeImageProgress.
			return runtimeapi.DecodeImageProgress(ctx, stream, nil)
		}
		return nil
	case runtimeapi.ResourceVolume:
		if action != runtimeapi.ActionRemove {
			return runtimeapi.UnsupportedError(runtimeapi.Operation(runtimeapi.ResourceVolume, string(action)))
		}
		return s.Client.REST.RemoveVolume(ctx, ref.ID, lc.Force)
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
		lc := options.Lifecycle
		return s.Client.REST.ExecuteContainerAction(ctx, id, string(action), podman.ContainerActionOptions{
			Timeout: lc.Timeout,
			Force:   lc.Force,
			Signal:  lc.Signal,
			Name:    lc.Name,
		})
	}
}

// dispatchAdvanced mirrors the Docker adapter's TASK-019 routing. It
// funnels the advanced actions through PodmanContainerService so the
// REST client stays the single integration point.
func (s PodmanResourceActionService) dispatchAdvanced(ctx context.Context, id string, action runtimeapi.Action, options runtimeapi.ActionOptions) error {
	svc := PodmanContainerService(s)
	switch action {
	case runtimeapi.ActionUpdate:
		if options.Update == nil {
			return invalidContainerActionPodman(action, id, "update options required")
		}
		_, err := svc.Update(ctx, id, runtimeapi.ContainerUpdateOptions{
			Memory:            options.Update.Memory,
			NanoCPUs:          options.Update.NanoCPUs,
			RestartPolicy:     options.Update.RestartPolicy,
			RestartMaxRetries: options.Update.RestartMaxRetries,
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
		if options.Commit == nil {
			return invalidContainerActionPodman(action, id, "commit options required")
		}
		_, err := svc.Commit(ctx, id, runtimeapi.ContainerCommitOptions{
			Repository: options.Commit.Repository,
			Tag:        options.Commit.Tag,
			Comment:    options.Commit.Comment,
			Author:     options.Commit.Author,
			Pause:      options.Commit.Pause,
		})
		return err
	case runtimeapi.ActionWait:
		var cond string
		if options.Wait != nil {
			cond = options.Wait.Condition
		}
		_, err := svc.Wait(ctx, id, cond)
		return err
	case runtimeapi.ActionCopy:
		if options.Copy == nil {
			return invalidContainerActionPodman(action, id, "copy options required")
		}
		rc, err := svc.CopyFromContainer(ctx, id, options.Copy.SourcePath)
		if err != nil {
			return err
		}
		return rc.Close()
	}
	return runtimeapi.UnsupportedError(runtimeapi.Operation(runtimeapi.ResourceContainer, string(action)))
}

// invalidContainerActionPodman produces the standard "missing option" error
// used by advanced container actions that require typed sub-payloads.
func invalidContainerActionPodman(action runtimeapi.Action, id, message string) error {
	return runtimeapi.NewError(runtimeapi.ErrorInvalid,
		runtimeapi.Operation(runtimeapi.ResourceContainer, string(action)), id, fmt.Errorf("%s", message))
}
